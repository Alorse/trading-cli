package volume

import (
	"context"
	"fmt"
	"sort"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

// VolumeBreakoutEntry represents a single volume breakout entry
type VolumeBreakoutEntry struct {
	Symbol         string             `json:"symbol"`
	ChangePercent  float64            `json:"changePercent"`
	VolumeRatio    float64            `json:"volumeRatio"`
	VolumeStrength float64            `json:"volumeStrength"`
	CurrentVolume  float64            `json:"currentVolume"`
	BreakoutType   string             `json:"breakoutType"`
	Indicators     map[string]float64 `json:"indicators"`
}

// VolumeBreakoutResult is the complete breakout scan result
type VolumeBreakoutResult struct {
	Exchange         string                `json:"exchange"`
	Timeframe        string                `json:"timeframe"`
	VolumeMultiplier float64               `json:"volumeMultiplier"`
	PriceChangeMin   float64               `json:"priceChangeMin"`
	TotalScanned     int                   `json:"totalScanned"`
	TotalFound       int                   `json:"totalFound"`
	Data             []VolumeBreakoutEntry `json:"data"`
}

// getFloat safely extracts a float64 value
func getFloat(values map[string]interface{}, key string) float64 {
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

// computeVolumeRatio calculates the volume ratio
func computeVolumeRatio(current, avg20 float64) float64 {
	if avg20 == 0 {
		return 0
	}
	return current / avg20
}

// computeVolumeStrength calculates volume strength capped at 10.0
func computeVolumeStrength(ratio float64) float64 {
	if ratio > 10.0 {
		return 10.0
	}
	return ratio
}

// computeBreakoutType determines if breakout is bullish or bearish
func computeBreakoutType(change float64) string {
	if change > 0 {
		return "bullish"
	}
	return "bearish"
}

// RunVolumeBreakout scans for volume breakouts
func RunVolumeBreakout(cfg *config.Config, exchange, timeframe string, volumeMultiplier, priceChangeMin float64, limit int, futures bool) error {
	// Validate inputs
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	if volumeMultiplier <= 0 {
		return fmt.Errorf("volumeMultiplier must be positive")
	}

	// Load symbols
	symbols, err := screener.LoadSymbols(exchange, futures)
	if err != nil {
		return fmt.Errorf("load symbols: %w", err)
	}

	// Scan up to 500 symbols to find movers anywhere in the list
	fetchCount := 500
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

	// Fetch analysis data in batches
	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)

	ctx := context.Background()
	allResults, err := tvClient.GetMultipleAnalysis(ctx, screenName, tickers, client.DefaultColumns)
	if err != nil {
		return fmt.Errorf("fetch analysis: %w", err)
	}

	// Process results and filter
	entries := make([]VolumeBreakoutEntry, 0)

	for _, result := range allResults {
		change := getFloat(result.Values, "change")
		volume := getFloat(result.Values, "volume")
		relVol := getFloat(result.Values, "relative_volume_10d_calc")

		// Use relative volume as volume ratio; fall back to 1.0 if unavailable
		ratio := relVol
		if ratio == 0 {
			ratio = 1.0
		}

		// Filter: abs(change) >= priceChangeMin AND volumeRatio >= volumeMultiplier
		if absFloat(change) < priceChangeMin || ratio < volumeMultiplier {
			continue
		}

		strength := computeVolumeStrength(ratio)
		bType := computeBreakoutType(change)

		entry := VolumeBreakoutEntry{
			Symbol:         result.Symbol,
			ChangePercent:  change,
			VolumeRatio:    ratio,
			VolumeStrength: strength,
			CurrentVolume:  volume,
			BreakoutType:   bType,
			Indicators: map[string]float64{
				"RSI": getFloat(result.Values, "RSI"),
				"ADX": getFloat(result.Values, "ADX"),
				"ATR": getFloat(result.Values, "ATR"),
			},
		}

		entries = append(entries, entry)
	}

	// Sort by volumeStrength desc, then abs(change) desc
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].VolumeStrength != entries[j].VolumeStrength {
			return entries[i].VolumeStrength > entries[j].VolumeStrength
		}
		return absFloat(entries[i].ChangePercent) > absFloat(entries[j].ChangePercent)
	})

	// Return top limit
	if len(entries) > limit {
		entries = entries[:limit]
	}

	result := VolumeBreakoutResult{
		Exchange:         exchange,
		Timeframe:        timeframe,
		VolumeMultiplier: volumeMultiplier,
		PriceChangeMin:   priceChangeMin,
		TotalScanned:     len(allResults),
		TotalFound:       len(entries),
		Data:             entries,
	}

	return utils.PrintJSON(result)
}

// absFloat returns the absolute value of a float64
func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
