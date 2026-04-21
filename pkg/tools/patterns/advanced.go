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

// AdvancedCandleEntry represents an advanced candle pattern entry
type AdvancedCandleEntry struct {
	Symbol       string  `json:"symbol"`
	Score        int     `json:"score"`
	BodyRatio    float64 `json:"bodyRatio"`
	Change       float64 `json:"change"`
	Volume       float64 `json:"volume"`
	RSI          float64 `json:"rsi"`
	Direction    string  `json:"direction"`
}

// AdvancedCandleResult represents the complete scan result
type AdvancedCandleResult struct {
	Exchange          string                   `json:"exchange"`
	BaseTimeframe     string                   `json:"baseTimeframe"`
	MinSizeIncrease   float64                  `json:"minSizeIncrease"`
	TotalFound        int                      `json:"totalFound"`
	Data              []AdvancedCandleEntry    `json:"data"`
}

// getFloatAdv safely extracts a float64 value
func getFloatAdv(values map[string]interface{}, key string) float64 {
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

// absFloatAdv returns absolute value of a float
func absFloatAdv(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// scoreAdvancedCandleEntry scores a candle on a 0-7 scale
func scoreAdvancedCandleEntry(bodyRatio, change, volume, rsi, close, ema50 float64, minSizeIncrease float64) int {
	score := 0

	// bodyRatio > 0.7: +2, > 0.5: +1
	if bodyRatio > 0.7 {
		score += 2
	} else if bodyRatio > 0.5 {
		score += 1
	}

	// abs(change) >= minSizeIncrease: +2, >= minSizeIncrease/2: +1
	absChange := absFloatAdv(change)
	if absChange >= minSizeIncrease {
		score += 2
	} else if absChange >= minSizeIncrease/2 {
		score += 1
	}

	// volume > 5000: +1
	if volume > 5000 {
		score += 1
	}

	// (change > 0 AND RSI > 50) OR (change < 0 AND RSI < 50): +1
	if (change > 0 && rsi > 50) || (change < 0 && rsi < 50) {
		score += 1
	}

	// (change > 0 AND close > EMA50) OR (change < 0 AND close < EMA50): +1
	if (change > 0 && close > ema50) || (change < 0 && close < ema50) {
		score += 1
	}

	return score
}

// RunAdvancedCandle scans for advanced candle patterns
func RunAdvancedCandle(cfg *config.Config, exchange, baseTimeframe string, minSizeIncrease float64, limit int) error {
	// Validate inputs
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	if minSizeIncrease <= 0 {
		return fmt.Errorf("minSizeIncrease must be positive")
	}

	// Load symbols
	symbols, err := screener.LoadSymbols(exchange)
	if err != nil {
		return fmt.Errorf("load symbols: %w", err)
	}

	// Calculate how many to fetch (no explicit limit mentioned, fetch more)
	fetchCount := limit * 3
	if fetchCount > 500 {
		fetchCount = 500
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

	ctx := context.Background()
	allResults, err := tvClient.GetMultipleAnalysis(ctx, screenName, tickers, client.DefaultColumns)
	if err != nil {
		return fmt.Errorf("fetch analysis: %w", err)
	}

	entries := make([]AdvancedCandleEntry, 0)

	for _, result := range allResults {
		close := getFloatAdv(result.Values, "close")
		open := getFloatAdv(result.Values, "open")
		high := getFloatAdv(result.Values, "high")
		low := getFloatAdv(result.Values, "low")
		change := getFloatAdv(result.Values, "change")
		volume := getFloatAdv(result.Values, "volume")
		rsi := getFloatAdv(result.Values, "RSI")
		ema50 := getFloatAdv(result.Values, "EMA50")

		bodyRatio := computeCandleBodyRatio(close, open, high, low)

		score := scoreAdvancedCandleEntry(bodyRatio, change, volume, rsi, close, ema50, minSizeIncrease)

		// Filter: include if score >= 3
		if score < 3 {
			continue
		}

		// Determine direction
		direction := "bullish"
		if change < 0 {
			direction = "bearish"
		}

		entry := AdvancedCandleEntry{
			Symbol:    result.Symbol,
			Score:     score,
			BodyRatio: bodyRatio,
			Change:    change,
			Volume:    volume,
			RSI:       rsi,
			Direction: direction,
		}

		entries = append(entries, entry)
	}

	// Sort by score desc
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	// Return top limit
	if len(entries) > limit {
		entries = entries[:limit]
	}

	result := AdvancedCandleResult{
		Exchange:        exchange,
		BaseTimeframe:   baseTimeframe,
		MinSizeIncrease: minSizeIncrease,
		TotalFound:      len(entries),
		Data:            entries,
	}

	return utils.PrintJSON(result)
}
