package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSaveSendsPutWithBody(t *testing.T) {
	var gotPath, gotQuery, gotAuth, gotMethod string
	var gotBody Update
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(`{"data":{"date":"2026-10-09","time":"08:30:00","description":"- a","code_reviews":null,"deploys":"VIRT-1"}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	entry, err := c.Save(context.Background(), time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC),
		Update{Time: "8:30", Description: "- a", Deploys: "VIRT-1"})
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPut || gotPath != "/api/time-tracking/2026/10/9" || gotQuery != "separated=true" {
		t.Errorf("request = %s %s?%s", gotMethod, gotPath, gotQuery)
	}
	if gotAuth != "Bearer tok" {
		t.Errorf("auth = %q", gotAuth)
	}
	if gotBody.Time != "8:30" || gotBody.Deploys != "VIRT-1" {
		t.Errorf("body = %+v", gotBody)
	}
	if entry.Time != "08:30:00" || entry.CodeReviews != "" || entry.Deploys != "VIRT-1" {
		t.Errorf("entry = %+v", entry)
	}
}

func TestSaveReportsValidationErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errors":{"time":["is invalid"]}}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL, "tok").Save(context.Background(), time.Now(), Update{Time: "99"})
	if err == nil || !strings.Contains(err.Error(), "time is invalid") {
		t.Errorf("err = %v", err)
	}
}

func TestSaveTokenRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := New(srv.URL, "bad").Save(context.Background(), time.Now(), Update{})
	if err == nil || !strings.Contains(err.Error(), "token rejected") {
		t.Errorf("err = %v", err)
	}
}
