package models

// ScreenerResult represents a single symbol from a screener scan.
type ScreenerResult struct {
	Symbol        string             `json:"symbol"`
	ChangePercent float64            `json:"changePercent"`
	Indicators    ScreenerIndicators `json:"indicators"`
}

// ScreenerIndicators holds the technical indicators for a screener result.
type ScreenerIndicators struct {
	Open    float64 `json:"open"`
	Close   float64 `json:"close"`
	High    float64 `json:"high"`
	Low     float64 `json:"low"`
	SMA20   float64 `json:"sma20"`
	BBUpper float64 `json:"bbUpper"`
	BBLower float64 `json:"bbLower"`
	EMA50   float64 `json:"ema50"`
	RSI     float64 `json:"rsi"`
	Volume  float64 `json:"volume"`
}
