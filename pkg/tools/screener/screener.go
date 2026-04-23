package screener

import (
	"strings"

	"github.com/alorse/trading-cli/pkg/client"
)

// timeframeSuffix maps logical timeframes to TradingView column suffixes.
// Empty string means the screener's default timeframe (1D).
var timeframeSuffix = map[string]string{
	"5m":  "|5",
	"15m": "|15",
	"1h":  "|60",
	"4h":  "|240",
	"1D":  "",
	"1W":  "|1W",
	"1M":  "|1M",
}

// ApplyTimeframe appends the TradingView timeframe suffix to each column.
// For the default timeframe (1D) columns are returned unchanged.
func ApplyTimeframe(columns []string, timeframe string) []string {
	suffix := timeframeSuffix[timeframe]
	if suffix == "" {
		return columns
	}
	out := make([]string, len(columns))
	for i, col := range columns {
		out[i] = col + suffix
	}
	return out
}

// NormalizeResults strips timeframe suffixes from column names in the response
// so downstream code can read values using the original unsuffixed keys.
func NormalizeResults(results []client.TVSymbolData, timeframe string) []client.TVSymbolData {
	suffix := timeframeSuffix[timeframe]
	if suffix == "" {
		return results
	}
	for i := range results {
		normalized := make(map[string]interface{}, len(results[i].Values))
		for k, v := range results[i].Values {
			normalized[strings.TrimSuffix(k, suffix)] = v
		}
		results[i].Values = normalized
	}
	return results
}

// GetFloatFromInterface safely extracts a float64 from an interface{} map value
// Returns 0 if the key is missing or the value is not a number
func GetFloatFromInterface(values map[string]interface{}, key string) float64 {
	val, ok := values[key]
	if !ok {
		return 0
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// getFloat is an unexported wrapper for backward compatibility
func getFloat(values map[string]interface{}, key string) float64 {
	return GetFloatFromInterface(values, key)
}

// getInt safely extracts an int64 from an interface{} map value
// Handles both int64 and float64 types
// Returns 0 if the key is missing or the value is not a number
func getInt(values map[string]interface{}, key string) int64 {
	val, ok := values[key]
	if !ok {
		return 0
	}

	switch v := val.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	default:
		return 0
	}
}

// computeBBW computes the Bollinger Band Width
// Formula: (bbUpper - bbLower) / sma20
// Returns 0 if any required value is missing or sma20 is zero
func computeBBW(values map[string]interface{}) float64 {
	bbUpper := getFloat(values, "BB.upper")
	bbLower := getFloat(values, "BB.lower")
	sma20 := getFloat(values, "SMA20")

	if bbUpper == 0 || bbLower == 0 || sma20 == 0 {
		return 0
	}

	return (bbUpper - bbLower) / sma20
}

// computeBBRating computes a rating based on Bollinger Bands and SMA20 middle
// Ratings:
// +3: close > bbUpper
// +2: close > middle + (upper-middle)/2
// +1: close > middle
// -1: close < middle
// -2: close < middle - (middle-lower)/2
// -3: close < bbLower
// 0: otherwise (or missing data)
func computeBBRating(values map[string]interface{}) int {
	close := getFloat(values, "close")
	bbUpper := getFloat(values, "BB.upper")
	bbLower := getFloat(values, "BB.lower")
	sma20 := getFloat(values, "SMA20")

	// If any required value is missing or zero, return 0
	if close == 0 || bbUpper == 0 || bbLower == 0 || sma20 == 0 {
		return 0
	}

	// sma20 is the middle of the Bollinger Bands
	middle := sma20
	upper := bbUpper
	lower := bbLower

	if close > upper {
		return 3
	}
	if close > middle+(upper-middle)/2 {
		return 2
	}
	if close > middle {
		return 1
	}
	if close < lower {
		return -3
	}
	if close < middle-(middle-lower)/2 {
		return -2
	}
	if close < middle {
		return -1
	}

	return 0
}

// ScreenerIndicators holds extracted indicator values
type ScreenerIndicators struct {
	Open    float64 `json:"open"`
	Close   float64 `json:"close"`
	SMA20   float64 `json:"sma20"`
	BBUpper float64 `json:"bbUpper"`
	BBLower float64 `json:"bbLower"`
	EMA50   float64 `json:"ema50"`
	RSI     float64 `json:"rsi"`
	Volume  float64 `json:"volume"`
}

// ScreenerEntry represents a single entry in the screener results
type ScreenerEntry struct {
	Symbol        string             `json:"symbol"`
	ChangePercent float64            `json:"changePercent"`
	Indicators    ScreenerIndicators `json:"indicators"`
}

// buildEntry creates a ScreenerEntry from TVSymbolData
// Returns nil if EMA50 or RSI is 0 (missing/bad data)
func buildEntry(data client.TVSymbolData) *ScreenerEntry {
	values := data.Values

	// Filter out bad data - EMA50 and RSI must be present and non-zero
	ema50 := getFloat(values, "EMA50")
	rsi := getFloat(values, "RSI")

	if ema50 == 0 || rsi == 0 {
		return nil
	}

	indicators := ScreenerIndicators{
		Open:    getFloat(values, "open"),
		Close:   getFloat(values, "close"),
		SMA20:   getFloat(values, "SMA20"),
		BBUpper: getFloat(values, "BB.upper"),
		BBLower: getFloat(values, "BB.lower"),
		EMA50:   ema50,
		RSI:     rsi,
		Volume:  getFloat(values, "volume"),
	}

	return &ScreenerEntry{
		Symbol:        data.Symbol,
		ChangePercent: getFloat(values, "change"),
		Indicators:    indicators,
	}
}
