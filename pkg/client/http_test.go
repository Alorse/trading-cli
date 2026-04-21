package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alorse/trading-cli/internal/config"
)

func TestNewHTTPClient(t *testing.T) {
	cfg := &config.Config{
		HTTPTimeout: 10 * time.Second,
		MaxRetries:  2,
		RetryDelay:  100 * time.Millisecond,
		UserAgent:   "test/1.0",
	}
	c := NewHTTPClient(cfg)
	if c == nil {
		t.Fatal("client is nil")
	}
	if c.maxRetries != 2 {
		t.Errorf("expected maxRetries=2, got %d", c.maxRetries)
	}
}

func TestGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "test/1.0" {
			t.Errorf("unexpected User-Agent: %s", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		HTTPTimeout: 5 * time.Second,
		MaxRetries:  0,
		RetryDelay:  0,
		UserAgent:   "test/1.0",
	}
	c := NewHTTPClient(cfg)

	data, err := c.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"status":"ok"}` {
		t.Errorf("unexpected body: %s", string(data))
	}
}

func TestGet_ServerError_Retry(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`ok`))
	}))
	defer server.Close()

	cfg := &config.Config{
		HTTPTimeout: 5 * time.Second,
		MaxRetries:  3,
		RetryDelay:  10 * time.Millisecond,
		UserAgent:   "test/1.0",
	}
	c := NewHTTPClient(cfg)

	data, err := c.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "ok" {
		t.Errorf("unexpected body: %s", string(data))
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestGet_ClientError_NoRetry(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(400)
		w.Write([]byte(`bad request`))
	}))
	defer server.Close()

	cfg := &config.Config{
		HTTPTimeout: 5 * time.Second,
		MaxRetries:  3,
		RetryDelay:  10 * time.Millisecond,
		UserAgent:   "test/1.0",
	}
	c := NewHTTPClient(cfg)

	_, err := c.Get(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if calls != 1 {
		t.Errorf("client errors should not retry, got %d calls", calls)
	}
}

func TestGet_AllRetriesFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	cfg := &config.Config{
		HTTPTimeout: 5 * time.Second,
		MaxRetries:  2,
		RetryDelay:  10 * time.Millisecond,
		UserAgent:   "test/1.0",
	}
	c := NewHTTPClient(cfg)

	_, err := c.Get(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error after all retries fail")
	}
}

func TestPost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`posted`))
	}))
	defer server.Close()

	cfg := &config.Config{
		HTTPTimeout: 5 * time.Second,
		MaxRetries:  0,
		RetryDelay:  0,
		UserAgent:   "test/1.0",
	}
	c := NewHTTPClient(cfg)

	data, err := c.Post(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "posted" {
		t.Errorf("unexpected body: %s", string(data))
	}
}

func TestGet_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Errorf("missing custom header")
		}
		w.WriteHeader(200)
		w.Write([]byte(`ok`))
	}))
	defer server.Close()

	cfg := &config.Config{
		HTTPTimeout: 5 * time.Second,
		MaxRetries:  0,
		RetryDelay:  0,
		UserAgent:   "test/1.0",
	}
	c := NewHTTPClient(cfg)

	data, err := c.GetWithHeaders(context.Background(), server.URL, map[string]string{"X-Custom": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "ok" {
		t.Errorf("unexpected body: %s", string(data))
	}
}
