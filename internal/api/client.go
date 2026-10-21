package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	defer resp.Body.Close()

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

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}
