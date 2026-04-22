package patterns

import (
	"context"
	"fmt"
	"sort"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

// ConsecutiveEntry represents a candle pattern entry
type ConsecutiveEntry struct {
	Symbol          string  `json:"symbol"`
	Price           float64 `json:"price"`
	CurrentChange   float64 `json:"currentChange"`
	CandleBodyRatio float64 `json:"candleBodyRatio"`
	PatternStrength int     `json:"patternStrength"`
	Volume          float64 `json:"volume"`
	BollingerRating int     `json:"bollingerRating"`
	RSI             float64 `json:"rsi"`
}

// ConsecutiveResult represents the complete scan result
type ConsecutiveResult struct {
	Exchange    string             `json:"exchange"`
	Timeframe   string             `json:"timeframe"`
	PatternType string             `json:"patternType"`
	TotalFound  int                `json:"totalFound"`
	Data        []ConsecutiveEntry `json:"data"`
}

// getFloatConsec safely extracts a float64 value
func getFloatConsec(values map[string]interface{}, key string) float64 {
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

// RunConsecutiveCandles scans for consecutive candle patterns
func RunConsecutiveCandles(cfg *config.Config, exchange, timeframe, patternType string, minGrowth float64, limit int) error {
	// Validate inputs
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if patternType != "bullish" && patternType != "bearish" {
		return fmt.Errorf("patternType must be 'bullish' or 'bearish'")
	}
	if limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}

	// Load symbols
	symbols, err := screener.LoadSymbols(exchange)
	if err != nil {
		return fmt.Errorf("load symbols: %w", err)
	}

	// Calculate how many to fetch
	fetchCount := limit * 3
	if fetchCount > 200 {
		fetchCount = 200
	}
	if fetchCount > len(symbols) {
		fetchCount = len(symbols)
	}

	symbols = symbols[:fetchCount]

	// Format tickers
	tickers := make([]string, len(symbols))
	for i, sym := range symbols {
		tickers[i] = screener.FormatTicker(exchange, sym)
	}

	// Get screener for exchange
	screenName, err := client.ScreenerForExchange(exchange)
	if err != nil {
		return fmt.Errorf("invalid exchange: %w", err)
	}

	// Fetch analysis data
	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	allResults, err := tvClient.GetMultipleAnalysis(ctx, screenName, tickers, client.DefaultColumns)
	if err != nil {
		return fmt.Errorf("fetch analysis: %w", err)
	}

	// Extract screener's BB rating function
	entries := make([]ConsecutiveEntry, 0)

	for _, result := range allResults {
		close := getFloatConsec(result.Values, "close")
		open := getFloatConsec(result.Values, "open")
		high := getFloatConsec(result.Values, "high")
		low := getFloatConsec(result.Values, "low")
		change := getFloatConsec(result.Values, "change")
		volume := getFloatConsec(result.Values, "volume")
		rsi := getFloatConsec(result.Values, "RSI")
		sma20 := getFloatConsec(result.Values, "SMA20")
		bbUpper := getFloatConsec(result.Values, "BB.upper")
		bbLower := getFloatConsec(result.Values, "BB.lower")

		bodyRatio := computeCandleBodyRatio(close, open, high, low)

		var strength int
		if patternType == "bullish" {
			strength = scoreBullishCandle(change, bodyRatio, close, sma20, rsi, volume)
		} else {
			strength = scoreBearishCandle(change, bodyRatio, close, sma20, rsi, volume)
		}

		// Filter: include if strength >= 3
		if strength < 3 {
			continue
		}

		// Calculate Bollinger rating (from screener package)
		bbRating := computeBBRating(close, bbUpper, bbLower, sma20)

		entry := ConsecutiveEntry{
			Symbol:          result.Symbol,
			Price:           close,
			CurrentChange:   change,
			CandleBodyRatio: bodyRatio,
			PatternStrength: strength,
			Volume:          volume,
			BollingerRating: bbRating,
			RSI:             rsi,
		}

		entries = append(entries, entry)
	}

	// Sort by patternStrength desc
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PatternStrength > entries[j].PatternStrength
	})

	// Return top limit
	if len(entries) > limit {
		entries = entries[:limit]
	}

	result := ConsecutiveResult{
		Exchange:    exchange,
		Timeframe:   timeframe,
		PatternType: patternType,
		TotalFound:  len(entries),
		Data:        entries,
	}

	return utils.PrintJSON(result)
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
func computeBBRating(close, bbUpper, bbLower, sma20 float64) int {
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
