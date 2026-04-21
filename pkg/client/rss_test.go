package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>Bitcoin Hits New High</title>
      <link>https://example.com/bitcoin-high</link>
      <description>Bitcoin reaches $100,000</description>
      <pubDate>Mon, 01 Jan 2024 12:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Ethereum Upgrade Complete</title>
      <link>https://example.com/eth-upgrade</link>
      <description>Ethereum network upgrade successful</description>
      <pubDate>Tue, 02 Jan 2024 08:30:00 +0000</pubDate>
    </item>
  </channel>
</rss>`

func TestParseRSS(t *testing.T) {
	items, err := parseRSS([]byte(sampleRSS), "testfeed")
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "Bitcoin Hits New High" {
		t.Errorf("expected 'Bitcoin Hits New High', got %s", items[0].Title)
	}
	if items[0].Link != "https://example.com/bitcoin-high" {
		t.Errorf("unexpected link: %s", items[0].Link)
	}
	if items[0].Source != "testfeed" {
		t.Errorf("expected source=testfeed, got %s", items[0].Source)
	}
	if items[0].PubDate.IsZero() {
		t.Error("expected non-zero pub date")
	}
}

func TestRSSFetchURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	httpClient := testHTTPClient()
	rssClient := NewRSSClient(httpClient)

	ctx := context.Background()
	items, err := rssClient.FetchURL(ctx, srv.URL, "test")
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[1].Title != "Ethereum Upgrade Complete" {
		t.Errorf("unexpected second item title: %s", items[1].Title)
	}
}

func TestRSSUnknownFeed(t *testing.T) {
	httpClient := testHTTPClient()
	rssClient := NewRSSClient(httpClient)

	ctx := context.Background()
	_, err := rssClient.FetchFeed(ctx, "nonexistent-feed")
	if err == nil {
		t.Error("expected error for unknown feed")
	}
}

func TestRSSMultipleFetches(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	httpClient := testHTTPClient()
	rssClient := NewRSSClient(httpClient)

	ctx := context.Background()
	items1, err := rssClient.FetchURL(ctx, srv.URL, "feed1")
	if err != nil {
		t.Fatal(err)
	}
	items2, err := rssClient.FetchURL(ctx, srv.URL, "feed2")
	if err != nil {
		t.Fatal(err)
	}

	total := append(items1, items2...)
	if len(total) != 4 {
		t.Errorf("expected 4 items total, got %d", len(total))
	}
	if callCount != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", callCount)
	}
}
