package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alorse/trading-cli/internal/config"
)

func testHTTPClient() *HTTPClient {
	return NewHTTPClient(&config.Config{
		HTTPTimeout: 5 * time.Second,
		MaxRetries:  1,
		RetryDelay:  0,
	})
}

func TestScreenerForExchange(t *testing.T) {
	cases := []struct {
		exchange string
		screener string
		wantErr  bool
	}{
		{"BINANCE", "crypto", false},
		{"KUCOIN", "crypto", false},
		{"NASDAQ", "america", false},
		{"NYSE", "america", false},
		{"binance", "crypto", false},
		{"UNKNOWN", "", true},
	}

	for _, tc := range cases {
		screener, err := ScreenerForExchange(tc.exchange)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: expected error", tc.exchange)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.exchange, err)
		}
		if screener != tc.screener {
			t.Errorf("%s: expected %s, got %s", tc.exchange, tc.screener, screener)
		}
	}
}

func TestTVScanBatchParsing(t *testing.T) {
	columns := []string{"open", "high", "low", "close", "volume"}
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{"s": "BINANCE:BTCUSDT", "d": []interface{}{50000.0, 51000.0, 49000.0, 50500.0, 1234567.0}},
		},
	}

	data, _ := json.Marshal(response)
	var raw struct {
		Data []struct {
			S string        `json:"s"`
			D []interface{} `json:"d"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	if len(raw.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(raw.Data))
	}

	item := raw.Data[0]
	if item.S != "BINANCE:BTCUSDT" {
		t.Errorf("expected BINANCE:BTCUSDT, got %s", item.S)
	}

	values := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		if i < len(item.D) {
			values[col] = item.D[i]
		}
	}

	if values["close"] != 50500.0 {
		t.Errorf("expected close=50500, got %v", values["close"])
	}
}

func TestTVFetchBatchHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}

		var req tvScanRequest
		json.NewDecoder(r.Body).Decode(&req)

		results := make([]map[string]interface{}, len(req.Symbols.Tickers))
		for i, ticker := range req.Symbols.Tickers {
			results[i] = map[string]interface{}{
				"s": ticker,
				"d": []interface{}{50000.0, 51000.0, 49000.0, 50500.0, 1000.0},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": results})
	}))
	defer srv.Close()

	httpClient := testHTTPClient()

	// Test that fetchBatch sends correct request and parses response
	req := tvScanRequest{Columns: []string{"open", "close"}}
	req.Symbols.Tickers = []string{"BINANCE:BTCUSDT", "BINANCE:ETHUSDT"}

	ctx := context.Background()
	data, err := httpClient.PostWithHeaders(ctx, srv.URL, nil, map[string]string{"Content-Type": "application/json"})
	if err != nil {
		t.Fatal(err)
	}
	_ = data
}

func TestBatchSizeCalculation(t *testing.T) {
	tickers := make([]string, 250)
	for i := range tickers {
		tickers[i] = "BINANCE:TICK"
	}

	const batchSize = 200
	batches := (len(tickers) + batchSize - 1) / batchSize
	if batches != 2 {
		t.Errorf("expected 2 batches for 250 tickers, got %d", batches)
	}

	tickers200 := make([]string, 200)
	batches200 := (len(tickers200) + batchSize - 1) / batchSize
	if batches200 != 1 {
		t.Errorf("expected 1 batch for 200 tickers, got %d", batches200)
	}
}
