package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func buildYahooResponse(timestamps []int64, opens, highs, lows, closes []float64, volumes []int64) map[string]interface{} {
	return map[string]interface{}{
		"chart": map[string]interface{}{
			"result": []map[string]interface{}{
				{
					"timestamp": timestamps,
					"indicators": map[string]interface{}{
						"quote": []map[string]interface{}{
							{
								"open":   opens,
								"high":   highs,
								"low":    lows,
								"close":  closes,
								"volume": volumes,
							},
						},
					},
				},
			},
			"error": nil,
		},
	}
}

func TestParseYahooChart(t *testing.T) {
	timestamps := []int64{1700000000, 1700086400}
	opens := []float64{50000.0, 51000.0}
	highs := []float64{51000.0, 52000.0}
	lows := []float64{49000.0, 50500.0}
	closes := []float64{50500.0, 51500.0}
	volumes := []int64{1000, 2000}

	resp := buildYahooResponse(timestamps, opens, highs, lows, closes, volumes)
	data, _ := json.Marshal(resp)

	candles, err := parseYahooChart(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(candles) != 2 {
		t.Fatalf("expected 2 candles, got %d", len(candles))
	}
	if candles[0].Close != 50500.0 {
		t.Errorf("expected close=50500, got %f", candles[0].Close)
	}
	if candles[0].Volume != 1000 {
		t.Errorf("expected volume=1000, got %d", candles[0].Volume)
	}
	if candles[0].Timestamp != 1700000000 {
		t.Errorf("expected timestamp=1700000000, got %d", candles[0].Timestamp)
	}
}

func TestYahooChartHTTP(t *testing.T) {
	timestamps := []int64{1700000000}
	closes := []float64{50500.0}
	opens := []float64{50000.0}
	highs := []float64{51000.0}
	lows := []float64{49000.0}
	volumes := []int64{5000}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := buildYahooResponse(timestamps, opens, highs, lows, closes, volumes)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	httpClient := testHTTPClient()
	ctx := context.Background()

	data, err := httpClient.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	candles, err := parseYahooChart(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}
	if candles[0].Close != 50500.0 {
		t.Errorf("expected close=50500, got %f", candles[0].Close)
	}
}

func TestParseYahooChartError(t *testing.T) {
	resp := map[string]interface{}{
		"chart": map[string]interface{}{
			"result": nil,
			"error": map[string]interface{}{
				"code":        "Not Found",
				"description": "No fundamentals data found",
			},
		},
	}
	data, _ := json.Marshal(resp)

	_, err := parseYahooChart(data)
	if err == nil {
		t.Error("expected error for yahoo error response")
	}
}

func TestParseYahooChartEmpty(t *testing.T) {
	resp := map[string]interface{}{
		"chart": map[string]interface{}{
			"result": []interface{}{},
			"error":  nil,
		},
	}
	data, _ := json.Marshal(resp)

	_, err := parseYahooChart(data)
	if err == nil {
		t.Error("expected error for empty result")
	}
}

func TestParseYahooSkipsZeroClose(t *testing.T) {
	timestamps := []int64{1700000000, 1700086400, 1700172800}
	opens := []float64{100.0, 0.0, 200.0}
	highs := []float64{110.0, 0.0, 210.0}
	lows := []float64{90.0, 0.0, 190.0}
	closes := []float64{105.0, 0.0, 205.0}
	volumes := []int64{1000, 0, 2000}

	resp := buildYahooResponse(timestamps, opens, highs, lows, closes, volumes)
	data, _ := json.Marshal(resp)

	candles, err := parseYahooChart(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(candles) != 2 {
		t.Errorf("expected 2 candles (skipping zero close), got %d", len(candles))
	}
}
