package backtest

import (
	"context"
	"fmt"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// RunBacktestStrategy executes a single strategy backtest.
// Validates inputs, fetches candles, selects strategy, and prints JSON results.
func RunBacktestStrategy(cfg *config.Config, symbol, strategy, period, interval string, initialCapital, commissionPct, slippagePct float64, includeTradeLog, includeEquityCurve bool) error {
	// Validate inputs
	if err := utils.ValidateStrategy(strategy); err != nil {
		return err
	}
	if err := utils.ValidatePeriod(period); err != nil {
		return err
	}
	if err := utils.ValidateInterval(interval); err != nil {
		return err
	}

	// Create HTTP and Yahoo clients
	httpClient := client.NewHTTPClient(cfg)
	yahooClient := client.NewYahooClient(httpClient)

	// Fetch candles
	ctx := context.Background()
	candles, err := yahooClient.GetChart(ctx, symbol, interval, period)
	if err != nil {
		return fmt.Errorf("fetch chart: %w", err)
	}

	if len(candles) == 0 {
		return fmt.Errorf("no candles retrieved for %s", symbol)
	}

	// Get strategy
	strat := GetStrategy(strategy)
	if strat == nil {
		return fmt.Errorf("unknown strategy: %s", strategy)
	}

	// Run backtest
	result := RunBacktest(symbol, strategy, candles, strat, initialCapital, commissionPct, slippagePct, includeTradeLog, includeEquityCurve)
	result.Period = period

	// Print result as JSON
	if err := utils.PrintJSON(result); err != nil {
		return fmt.Errorf("print result: %w", err)
	}

	return nil
}
