package screener

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// RunBollingerScan retrieves symbols with Bollinger Band Width below the specified threshold
func RunBollingerScan(cfg *config.Config, exchange, timeframe string, bbwThreshold float64, limit int) error {
	// Validate inputs
	if err := utils.ValidateTimeframe(timeframe); err != nil {
		return err
	}

	if err := utils.ValidateIntRange("limit", limit, 1, 1000); err != nil {
		return err
	}

	if err := utils.ValidateRange("bbwThreshold", bbwThreshold, 0, 1000); err != nil {
		return err
	}

	// Load symbols
	symbols, err := LoadSymbols(exchange)
	if err != nil {
		return fmt.Errorf("failed to load symbols: %w", err)
	}

	if len(symbols) == 0 {
		return fmt.Errorf("no symbols loaded for exchange %s", exchange)
	}

	// Get screener for exchange
	screener, err := client.ScreenerForExchange(exchange)
	if err != nil {
		return err
	}

	// Create HTTP client and TradingView client
	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)

	// Define columns to fetch
	columns := []string{
		"open", "high", "low", "close", "volume",
		"change", "SMA20", "BB.upper", "BB.lower", "EMA50", "RSI",
	}

	// Fetch up to limit*2 symbols (capped at total symbols)
	fetchCount := limit * 2
	if fetchCount > len(symbols) {
		fetchCount = len(symbols)
	}
	fetchSymbols := symbols[:fetchCount]

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch analysis data
	results, err := tvClient.GetMultipleAnalysis(ctx, screener, fetchSymbols, columns)
	if err != nil {
		return fmt.Errorf("failed to fetch analysis data: %w", err)
	}

	// Build entries and filter
	var entries []*ScreenerEntry
	for _, result := range results {
		entry := buildEntry(result)
		if entry == nil {
			continue
		}

		// Filter: BBW < threshold AND BBW > 0 AND EMA50 != 0 AND RSI != 0
		bbw := computeBBW(result.Values)
		if bbw > 0 && bbw < bbwThreshold {
			entries = append(entries, entry)
		}
	}

	// Sort by ChangePercent descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ChangePercent > entries[j].ChangePercent
	})

	// Limit results
	if len(entries) > limit {
		entries = entries[:limit]
	}

	// Output as JSON
	return utils.PrintJSON(entries)
}
