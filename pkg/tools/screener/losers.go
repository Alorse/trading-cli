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

// RunTopLosers retrieves and displays the top losing symbols for a given exchange and timeframe
func RunTopLosers(cfg *config.Config, exchange, timeframe string, limit int, futures bool) error {
	// Validate inputs
	if err := utils.ValidateTimeframe(timeframe); err != nil {
		return err
	}

	if err := utils.ValidateIntRange("limit", limit, 1, 1000); err != nil {
		return err
	}

	// Load symbols
	symbols, err := LoadSymbols(exchange, futures)
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

	// Apply timeframe suffix to columns
	columns = ApplyTimeframe(columns, timeframe)

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch analysis data
	results, err := tvClient.GetMultipleAnalysis(ctx, screener, symbols, columns)
	if err != nil {
		return fmt.Errorf("failed to fetch analysis data: %w", err)
	}

	// Normalize result keys back to unsuffixed names
	results = NormalizeResults(results, timeframe)

	// Build entries and filter
	entries := make([]*ScreenerEntry, 0)
	for _, result := range results {
		entry := buildEntry(result)
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	// Sort by ChangePercent ascending (most negative first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ChangePercent < entries[j].ChangePercent
	})

	// Limit results
	if len(entries) > limit {
		entries = entries[:limit]
	}

	// Output as JSON
	return utils.PrintJSON(entries)
}
