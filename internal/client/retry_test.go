package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// GET requests retry on 5xx until they succeed.
func TestRetriesGetOn5xx(t *testing.T) {
	t.Parallel()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groups": []map[string]any{{"id": 1, "name": "g", "zones": []map[string]any{{"id": 2}}}},
		})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "t", Timeout: 5 * time.Second, MaxRetries: 3})
	if err != nil {
		t.Fatal(err)
	}
	group, err := c.GetGroupByName(context.Background(), "g")
	if err != nil {
		t.Fatal(err)
	}
	if group.ID != 1 {
		t.Fatalf("unexpected group: %#v", group)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

// Non-idempotent POSTs must NOT retry on 5xx (avoids duplicate writes).
func TestDoesNotRetryPostOn5xx(t *testing.T) {
	t.Parallel()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "t", Timeout: 5 * time.Second, MaxRetries: 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateSecurityGroup(context.Background(), "n", ""); err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("POST should not retry on 5xx; attempts=%d", got)
	}
}

// 429 means the request was not processed, so it is retried for any method.
func TestRetriesPostOn429(t *testing.T) {
	t.Parallel()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"securityGroup": map[string]any{"id": 5, "name": "n"}})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "t", Timeout: 5 * time.Second, MaxRetries: 3})
	if err != nil {
		t.Fatal(err)
	}
	sg, err := c.CreateSecurityGroup(context.Background(), "n", "")
	if err != nil {
		t.Fatal(err)
	}
	if sg.ID != 5 {
		t.Fatalf("unexpected sg: %#v", sg)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 attempts (429 then 200), got %d", got)
	}
}

// MaxRetries < 0 disables retries entirely.
func TestRetriesDisabled(t *testing.T) {
	t.Parallel()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "t", Timeout: 5 * time.Second, MaxRetries: -1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetGroupByName(context.Background(), "g"); err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 attempt with retries disabled, got %d", got)
	}
}
