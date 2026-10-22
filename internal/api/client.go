package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ricardoruwer/tt-cli/internal/model"
)

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTP:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Month fetches every entry of the given month. The separated flag makes the
// API return code reviews and deploys as their own fields instead of folding
// them into the description.
func (c *Client) Month(ctx context.Context, year int, month time.Month) ([]model.Entry, error) {
	url := fmt.Sprintf("%s/api/time-tracking/%d/%d?separated=true", c.BaseURL, year, int(month))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("%s: token rejected, check the token in your config", resp.Status)
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%s: %s", resp.Status, snippet(body))
	}

	var payload struct {
		Data []model.Entry `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return payload.Data, nil
}

// Update is the editable part of a day's entry. Time accepts "H", "H:M" or
// "H:M:S"; an empty string clears it.
type Update struct {
	Time        string `json:"time"`
	Description string `json:"description"`
	CodeReviews string `json:"code_reviews"`
	Deploys     string `json:"deploys"`
}

// Save creates or updates the entry of the given date and returns the stored
// record.
func (c *Client) Save(ctx context.Context, date time.Time, u Update) (model.Entry, error) {
	body, err := json.Marshal(u)
	if err != nil {
		return model.Entry{}, err
	}
	url := fmt.Sprintf("%s/api/time-tracking/%d/%d/%d?separated=true",
		c.BaseURL, date.Year(), int(date.Month()), date.Day())
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return model.Entry{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return model.Entry{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return model.Entry{}, err
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return model.Entry{}, fmt.Errorf("%s: token rejected, check the token in your config", resp.Status)
	case resp.StatusCode == http.StatusUnprocessableEntity:
		return model.Entry{}, fmt.Errorf("rejected: %s", validationErrors(respBody))
	case resp.StatusCode != http.StatusOK:
		return model.Entry{}, fmt.Errorf("%s: %s", resp.Status, snippet(respBody))
	}

	var payload struct {
		Data model.Entry `json:"data"`
	}
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return model.Entry{}, fmt.Errorf("decoding response: %w", err)
	}
	return payload.Data, nil
}

// validationErrors renders a {"errors": {"field": ["msg"]}} body as
// "field msg; field msg", falling back to the raw body.
func validationErrors(body []byte) string {
	var payload struct {
		Errors map[string][]string `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || len(payload.Errors) == 0 {
		return snippet(body)
	}
	fields := make([]string, 0, len(payload.Errors))
	for f := range payload.Errors {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	var parts []string
	for _, f := range fields {
		parts = append(parts, f+" "+strings.Join(payload.Errors[f], ", "))
	}
	return strings.Join(parts, "; ")
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}
