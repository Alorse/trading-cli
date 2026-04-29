package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/tools/analysis"
	"github.com/alorse/trading-cli/pkg/tools/backtest"
	"github.com/alorse/trading-cli/pkg/tools/mtf"
	"github.com/alorse/trading-cli/pkg/tools/patterns"
	"github.com/alorse/trading-cli/pkg/tools/plan"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/tools/sentiment"
	"github.com/alorse/trading-cli/pkg/tools/system"
	"github.com/alorse/trading-cli/pkg/tools/volume"
	"github.com/alorse/trading-cli/pkg/tools/yahoo"
)

var cfg = config.Load()

// captureOutput redirects os.Stdout, runs fn, captures the output, and
// restores os.Stdout. It is safe for concurrent use via stdoutMu.
var stdoutMu sync.Mutex

func captureOutput(fn func() error) (string, error) {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("pipe: %w", err)
	}
	os.Stdout = w

	fnErr := fn()

	if err := w.Close(); err != nil {
		os.Stdout = oldStdout
		return "", fmt.Errorf("close pipe writer: %w", err)
	}
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	if fnErr != nil {
		return "", fnErr
	}
	return string(out), nil
}

// runTool captures stdout from fn and returns annotated content blocks.
// If the captured output is valid JSON, it's parsed and returned as
// structuredContent too.
func runTool(label string, fn func() error) ([]ContentBlock, interface{}, error) {
	out, err := captureOutput(fn)
	if err != nil {
		return nil, nil, toolExecutionError(label, err)
	}

	out = trimTrailingNewline(out)
	if out == "" {
		return []ContentBlock{{Type: "text", Text: label + " completed successfully."}}, nil, nil
	}

	var parsed any
	if json.Unmarshal([]byte(out), &parsed) == nil {
		return buildAnnotatedContentBlocks(label+" results:", parsed)
	}
	return []ContentBlock{{Type: "text", Text: out}}, nil, nil
}

func trimTrailingNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}

// ---------------------------------------------------------------------------
// Screener handlers
// ---------------------------------------------------------------------------

func topGainersHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("top_gainers", func() error {
		return screener.RunTopGainers(cfg,
			argString(args, "exchange"),
			argString(args, "timeframe"),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

func topLosersHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("top_losers", func() error {
		return screener.RunTopLosers(cfg,
			argString(args, "exchange"),
			argString(args, "timeframe"),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

func bollingerScanHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("bollinger_scan", func() error {
		return screener.RunBollingerScan(cfg,
			argString(args, "exchange"),
			argString(args, "timeframe"),
			argFloat(args, "bbw_threshold", 0.04),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

func ratingFilterHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("rating_filter", func() error {
		return screener.RunRatingFilter(cfg,
			argString(args, "exchange"),
			argString(args, "timeframe"),
			argInt(args, "rating", 2),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

// ---------------------------------------------------------------------------
// Analysis handlers
// ---------------------------------------------------------------------------

func coinAnalysisHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("coin_analysis", func() error {
		return analysis.RunCoinAnalysis(cfg,
			symbol,
			argString(args, "exchange"),
			argString(args, "timeframe"),
		)
	})
}

// ---------------------------------------------------------------------------
// Pattern handlers
// ---------------------------------------------------------------------------

func consecutiveCandlesHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("consecutive_candles_scan", func() error {
		return patterns.RunConsecutiveCandles(cfg,
			argString(args, "exchange"),
			argString(args, "timeframe"),
			argString(args, "pattern_type"),
			argFloat(args, "min_growth", 2.0),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

func advancedCandleHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("advanced_candle_pattern", func() error {
		return patterns.RunAdvancedCandle(cfg,
			argString(args, "exchange"),
			argString(args, "base_timeframe"),
			argFloat(args, "min_size_increase", 10.0),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

// ---------------------------------------------------------------------------
// Volume handlers
// ---------------------------------------------------------------------------

func volumeBreakoutHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("volume_breakout_scanner", func() error {
		return volume.RunVolumeBreakout(cfg,
			argString(args, "exchange"),
			argString(args, "timeframe"),
			argFloat(args, "volume_multiplier", 2.0),
			argFloat(args, "price_change_min", 3.0),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

func volumeConfirmationHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("volume_confirmation_analysis", func() error {
		return volume.RunVolumeConfirmation(cfg,
			symbol,
			argString(args, "exchange"),
			argString(args, "timeframe"),
		)
	})
}

func smartVolumeHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("smart_volume_scanner", func() error {
		return volume.RunSmartVolume(cfg,
			argString(args, "exchange"),
			argFloat(args, "min_volume_ratio", 2.0),
			argFloat(args, "min_price_change", 2.0),
			argString(args, "rsi_range"),
			argInt(args, "limit", 10),
			argBool(args, "futures", false),
		)
	})
}

// ---------------------------------------------------------------------------
// Multi-timeframe handler
// ---------------------------------------------------------------------------

func multiTimeframeHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("multi_timeframe_analysis", func() error {
		return mtf.RunMultiTimeframe(cfg,
			symbol,
			argString(args, "exchange"),
		)
	})
}

// ---------------------------------------------------------------------------
// Sentiment handlers
// ---------------------------------------------------------------------------

func marketSentimentHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("market_sentiment", func() error {
		return sentiment.RunMarketSentiment(cfg,
			symbol,
			argString(args, "category"),
			argInt(args, "limit", 50),
		)
	})
}

func financialNewsHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("financial_news", func() error {
		return sentiment.RunFinancialNews(cfg,
			argString(args, "symbol"),
			argString(args, "category"),
			argInt(args, "limit", 10),
		)
	})
}

func combinedAnalysisHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("combined_analysis", func() error {
		return sentiment.RunCombinedAnalysis(cfg,
			symbol,
			argString(args, "exchange"),
			argString(args, "timeframe"),
			argString(args, "category"),
		)
	})
}

// ---------------------------------------------------------------------------
// Backtest handlers
// ---------------------------------------------------------------------------

func backtestStrategyHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	strategy := argString(args, "strategy")
	if symbol == "" || strategy == "" {
		return nil, nil, fmt.Errorf("symbol and strategy are required")
	}
	return runTool("backtest_strategy", func() error {
		return backtest.RunBacktestStrategy(cfg,
			symbol, strategy,
			argString(args, "period"),
			argString(args, "interval"),
			argFloat(args, "initial_capital", 10000),
			argFloat(args, "commission_pct", 0.1),
			argFloat(args, "slippage_pct", 0.05),
			argBool(args, "include_trade_log", false),
			argBool(args, "include_equity_curve", false),
		)
	})
}

func compareStrategiesHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("compare_strategies", func() error {
		return backtest.RunCompareStrategies(cfg,
			symbol,
			argString(args, "period"),
			argString(args, "interval"),
			argFloat(args, "initial_capital", 10000),
		)
	})
}

func walkForwardHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	strategy := argString(args, "strategy")
	if symbol == "" || strategy == "" {
		return nil, nil, fmt.Errorf("symbol and strategy are required")
	}
	return runTool("walk_forward_backtest", func() error {
		return backtest.RunWalkForwardBacktest(cfg,
			symbol, strategy,
			argString(args, "period"),
			argString(args, "interval"),
			argFloat(args, "initial_capital", 10000),
			argFloat(args, "commission_pct", 0.1),
			argFloat(args, "slippage_pct", 0.05),
			argInt(args, "n_splits", 3),
			argFloat(args, "train_ratio", 0.7),
		)
	})
}

// ---------------------------------------------------------------------------
// Yahoo Finance handlers
// ---------------------------------------------------------------------------

func yahooPriceHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("yahoo_price", func() error {
		return yahoo.RunYahooPrice(cfg, symbol)
	})
}

func marketSnapshotHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("market_snapshot", func() error {
		return yahoo.RunMarketSnapshot(cfg)
	})
}

// ---------------------------------------------------------------------------
// Trade Plan handlers
// ---------------------------------------------------------------------------

func tradePlanHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("trade_plan", func() error {
		return plan.RunTradePlan(cfg,
			symbol,
			argString(args, "exchange"),
			argString(args, "timeframe"),
		)
	})
}

func fibonacciHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	symbol := argString(args, "symbol")
	if symbol == "" {
		return nil, nil, fmt.Errorf("symbol is required")
	}
	return runTool("fibonacci_retracement", func() error {
		return plan.RunFibonacci(cfg,
			symbol,
			argString(args, "exchange"),
			argString(args, "lookback"),
			argString(args, "timeframe"),
		)
	})
}

// ---------------------------------------------------------------------------
// System handlers
// ---------------------------------------------------------------------------

func listExchangesHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("list_exchanges", func() error {
		return system.RunListExchanges()
	})
}

func versionHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("version", func() error {
		return system.RunVersion()
	})
}

func healthHandler(ctx context.Context, args map[string]any) ([]ContentBlock, interface{}, error) {
	return runTool("health", func() error {
		return system.RunHealth(cfg)
	})
}
