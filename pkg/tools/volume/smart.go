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

// SmartVolumeEntry represents a smart volume scan entry with RSI filtering
type SmartVolumeEntry struct {
	Symbol                string             `json:"symbol"`
	ChangePercent         float64            `json:"changePercent"`
	VolumeRatio           float64            `json:"volumeRatio"`
	VolumeStrength        float64            `json:"volumeStrength"`
	CurrentVolume         float64            `json:"currentVolume"`
	BreakoutType          string             `json:"breakoutType"`
	RSI                   float64            `json:"rsi"`
	TradingRecommendation string             `json:"tradingRecommendation"`
	Indicators            map[string]float64 `json:"indicators"`
}

// SmartVolumeResult is the complete smart volume scan result
type SmartVolumeResult struct {
	Exchange      string             `json:"exchange"`
	MinVolumeRatio float64            `json:"minVolumeRatio"`
	MinPriceChange float64            `json:"minPriceChange"`
	RSIRange      string             `json:"rsiRange"`
	TotalScanned  int                `json:"totalScanned"`
	TotalFound    int                `json:"totalFound"`
	Data          []SmartVolumeEntry `json:"data"`
}

// shouldIncludeByRSI checks if RSI is within the specified range
func shouldIncludeByRSI(rsi float64, rsiRange string) bool {
	switch rsiRange {
	case "oversold":
		return rsi < 30
	case "overbought":
		return rsi > 70
	case "neutral":
		return rsi >= 30 && rsi <= 70
	case "any":
		return true
	default:
		return true
	}
}

// computeTradingRecommendation determines the trading recommendation
func computeTradingRecommendation(change, ratio, rsi float64) string {
	if change > 0 && ratio >= 2 {
		if rsi < 70 {
			return "STRONG BUY"
		}
		return "OVERBOUGHT - CAUTION"
	}

	if change < 0 && ratio >= 2 {
		if rsi > 30 {
			return "STRONG SELL"
		}
		return "OVERSOLD - OPPORTUNITY?"
	}

	return "NEUTRAL"
}

// RunSmartVolume performs smart volume scanning with RSI filtering
func RunSmartVolume(cfg *config.Config, exchange string, minVolumeRatio, minPriceChange float64, rsiRange string, limit int) error {
	// Validate inputs
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	if minVolumeRatio <= 0 {
		return fmt.Errorf("minVolumeRatio must be positive")
	}

	// Load symbols
	symbols, err := screener.LoadSymbols(exchange)
	if err != nil {
		return fmt.Errorf("load symbols: %w", err)
	}

	// Calculate how many to fetch (limit*2 internally)
	fetchCount := limit * 2
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

	// Process results and filter
	entries := make([]SmartVolumeEntry, 0)

	for _, result := range allResults {
		change := getFloat(result.Values, "change")
		volume := getFloat(result.Values, "volume")
		volumeAvg20 := getFloat(result.Values, "average_volume_10d_calc")
		rsi := getFloat(result.Values, "RSI")

		ratio := computeVolumeRatio(volume, volumeAvg20)

		// Filter: abs(change) >= minPriceChange AND volumeRatio >= minVolumeRatio
		if absFloat(change) < minPriceChange || ratio < minVolumeRatio {
			continue
		}

		// Filter by RSI range
		if !shouldIncludeByRSI(rsi, rsiRange) {
			continue
		}

		strength := computeVolumeStrength(ratio)
		bType := computeBreakoutType(change)
		rec := computeTradingRecommendation(change, ratio, rsi)

		entry := SmartVolumeEntry{
			Symbol:                result.Symbol,
			ChangePercent:         change,
			VolumeRatio:           ratio,
			VolumeStrength:        strength,
			CurrentVolume:         volume,
			BreakoutType:          bType,
			RSI:                   rsi,
			TradingRecommendation: rec,
			Indicators: map[string]float64{
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

	result := SmartVolumeResult{
		Exchange:       exchange,
		MinVolumeRatio: minVolumeRatio,
		MinPriceChange: minPriceChange,
		RSIRange:       rsiRange,
		TotalScanned:   len(allResults),
		TotalFound:     len(entries),
		Data:           entries,
	}

	return utils.PrintJSON(result)
}
