package backtest

import (
	"context"
	"fmt"
	"sort"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// ComparisonResult holds the comparison of all strategies.
type ComparisonResult struct {
	Symbol  string          `json:"symbol"`
	Period  string          `json:"period"`
	Winner  string          `json:"winner"`
	Ranking []BacktestResult `json:"ranking"`
}

// RunCompareStrategies runs all 6 strategies on the same data and compares them.
// Sorts results by TotalReturn descending and outputs JSON.
func RunCompareStrategies(cfg *config.Config, symbol, period, interval string, initialCapital float64) error {
	// Validate inputs
	if err := utils.ValidatePeriod(period); err != nil {
		return err
	}
	if err := utils.ValidateInterval(interval); err != nil {
		return err
	}

	// Create HTTP and Yahoo clients
	httpClient := client.NewHTTPClient(cfg)
	yahooClient := client.NewYahooClient(httpClient)

	// Fetch candles once
	ctx := context.Background()
	candles, err := yahooClient.GetChart(ctx, symbol, interval, period)
	if err != nil {
		return fmt.Errorf("fetch chart: %w", err)
	}

	if len(candles) == 0 {
		return fmt.Errorf("no candles retrieved for %s", symbol)
	}

	// Define all strategies
	strategyNames := []string{"rsi", "bollinger", "macd", "ema-cross", "supertrend", "donchian"}
	var results []BacktestResult

	// Run all strategies
	for _, strategyName := range strategyNames {
		strat := GetStrategy(strategyName)
		if strat == nil {
			continue
		}

		result := RunBacktest(symbol, strategyName, candles, strat, initialCapital, 0.1, 0.1, false, false)
		result.Period = period
		results = append(results, result)
	}

	// Sort by TotalReturn descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalReturn > results[j].TotalReturn
	})

	// Determine winner
	winner := ""
	if len(results) > 0 {
		winner = results[0].Strategy
	}

	// Create comparison result
	comparison := ComparisonResult{
		Symbol:  symbol,
		Period:  period,
		Winner:  winner,
		Ranking: results,
	}

	// Print result as JSON
	if err := utils.PrintJSON(comparison); err != nil {
		return fmt.Errorf("print result: %w", err)
	}

	return nil
}
