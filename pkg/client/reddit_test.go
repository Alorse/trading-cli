package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func buildRedditResponse(posts []RedditPost) map[string]interface{} {
	children := make([]map[string]interface{}, len(posts))
	for i, p := range posts {
		children[i] = map[string]interface{}{"data": p}
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"children": children,
		},
	}
}

func TestParseRedditResponse(t *testing.T) {
	posts := []RedditPost{
		{Title: "BTC to the moon", Subreddit: "CryptoCurrency", Score: 500, NumComments: 123},
		{Title: "ETH analysis", Subreddit: "ethereum", Score: 200, NumComments: 45},
	}

	resp := buildRedditResponse(posts)
	data, _ := json.Marshal(resp)

	result, err := parseRedditResponse(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result))
	}
	if result[0].Title != "BTC to the moon" {
		t.Errorf("expected 'BTC to the moon', got %s", result[0].Title)
	}
	if result[0].Score != 500 {
		t.Errorf("expected score=500, got %d", result[0].Score)
	}
}

func TestRedditSearchHTTP(t *testing.T) {
	posts := []RedditPost{
		{Title: "Bitcoin news", Subreddit: "Bitcoin", Score: 100},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			t.Error("expected query parameter 'q'")
		}
		ua := r.Header.Get("User-Agent")
		if ua == "" {
			t.Error("expected User-Agent header")
		}
		resp := buildRedditResponse(posts)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	httpClient := testHTTPClient()
	ctx := context.Background()

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 Chrome/120.0.0.0",
	}
	data, err := httpClient.GetWithHeaders(ctx, srv.URL+"?q=bitcoin", headers)
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseRedditResponse(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestRedditLimitClamping(t *testing.T) {
	cases := []struct {
		input    int
		expected int
	}{
		{0, 25},
		{-1, 25},
		{50, 50},
		{100, 100},
		{200, 100},
	}

	for _, tc := range cases {
		limit := tc.input
		if limit <= 0 {
			limit = 25
		}
		if limit > 100 {
			limit = 100
		}
		if limit != tc.expected {
			t.Errorf("input=%d: expected %d, got %d", tc.input, tc.expected, limit)
		}
	}
}

func TestParseRedditEmpty(t *testing.T) {
	resp := buildRedditResponse(nil)
	data, _ := json.Marshal(resp)

	result, err := parseRedditResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 posts, got %d", len(result))
	}
}
