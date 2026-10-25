package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

const phemexBase = "https://api.phemex.com"

// PhemexTicker holds the relevant 24h ticker fields for a perpetual contract.
type PhemexTicker struct {
	Symbol          string  `json:"symbol"`
	LastPrice       float64 `json:"lastPrice"`
	Open            float64 `json:"open"`
	High            float64 `json:"high"`
	Low             float64 `json:"low"`
	Volume          float64 `json:"volume"`      // base asset volume (24h)
	Turnover        float64 `json:"turnover"`    // quote asset turnover (24h, USDT for USDT perps)
	FundingRate     float64 `json:"fundingRate"` // current funding rate (decimal, e.g. 0.0001 = 0.01%)
	PredFundingRate float64 `json:"predFundingRate"`
	OpenInterest    float64 `json:"openInterest"` // in base asset units
	IndexPrice      float64 `json:"indexPrice"`
	MarkPrice       float64 `json:"markPrice"`
	ChangePct       float64 `json:"changePct"` // 24h price change %
}

// PhemexClient hits the public PHEMEX API.
type PhemexClient struct {
	http *HTTPClient
}

func NewPhemexClient(http *HTTPClient) *PhemexClient {
	return &PhemexClient{http: http}
}

// GetAllTickers returns the 24hr ticker for every PHEMEX perpetual.
// Only USDT-margined perpetuals are typically interesting for the scalp strategy,
// so this returns them all and the caller filters by symbol suffix if desired.
func (c *PhemexClient) GetAllTickers(ctx context.Context) ([]PhemexTicker, error) {
	url := phemexBase + "/md/v3/ticker/24hr/all?id=10001"
	body, err := c.http.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("phemex get: %w", err)
	}

	var raw struct {
		Result []map[string]interface{} `json:"result"`
		Error  interface{}              `json:"error"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("phemex parse: %w", err)
	}

	parseF := func(v interface{}) float64 {
		switch x := v.(type) {
		case string:
			f, _ := strconv.ParseFloat(x, 64)
			return f
		case float64:
			return x
		case json.Number:
			f, _ := x.Float64()
			return f
		}
		return 0
	}

	out := make([]PhemexTicker, 0, len(raw.Result))
	for _, row := range raw.Result {
		t := PhemexTicker{
			Symbol:          fmt.Sprintf("%v", row["symbol"]),
			LastPrice:       parseF(row["lastRp"]),
			Open:            parseF(row["openRp"]),
			High:            parseF(row["highRp"]),
			Low:             parseF(row["lowRp"]),
			Volume:          parseF(row["volumeRq"]),
			Turnover:        parseF(row["turnoverRv"]),
			FundingRate:     parseF(row["fundingRateRr"]),
			PredFundingRate: parseF(row["predFundingRateRr"]),
			OpenInterest:    parseF(row["openInterestRv"]),
			IndexPrice:      parseF(row["indexRp"]),
			MarkPrice:       parseF(row["markRp"]),
		}
		if t.Open > 0 {
			t.ChangePct = (t.LastPrice - t.Open) / t.Open * 100.0
		}
		if t.Symbol != "" && t.Symbol != "<nil>" {
			out = append(out, t)
		}
	}
	return out, nil
}
