package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const yahooChartBase = "https://query1.finance.yahoo.com/v8/finance/chart"

type YahooOHLCV struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

type YahooClient struct {
	http *HTTPClient
}

func NewYahooClient(http *HTTPClient) *YahooClient {
	return &YahooClient{http: http}
}

func (c *YahooClient) GetChart(ctx context.Context, symbol, interval, rangeStr string) ([]YahooOHLCV, error) {
	url := fmt.Sprintf("%s/%s?interval=%s&range=%s", yahooChartBase, symbol, interval, rangeStr)

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
	}

	data, err := c.http.GetWithHeaders(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("yahoo chart fetch: %w", err)
	}

	return parseYahooChart(data)
}

type yahooResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func parseYahooChart(data []byte) ([]YahooOHLCV, error) {
	var resp yahooResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse yahoo response: %w", err)
	}

	if resp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo error %s: %s", resp.Chart.Error.Code, resp.Chart.Error.Description)
	}

	if len(resp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart data returned")
	}

	result := resp.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data in chart")
	}

	quote := result.Indicators.Quote[0]
	n := len(result.Timestamp)
	candles := make([]YahooOHLCV, 0, n)

	for i := 0; i < n; i++ {
		if i >= len(quote.Close) || quote.Close[i] == 0 {
			continue
		}
		candle := YahooOHLCV{
			Timestamp: result.Timestamp[i],
			Close:     quote.Close[i],
		}
		if i < len(quote.Open) {
			candle.Open = quote.Open[i]
		}
		if i < len(quote.High) {
			candle.High = quote.High[i]
		}
		if i < len(quote.Low) {
			candle.Low = quote.Low[i]
		}
		if i < len(quote.Volume) {
			candle.Volume = quote.Volume[i]
		}
		candles = append(candles, candle)
	}

	return candles, nil
}
