package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

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

func runVersion() error {
	return system.RunVersion()
}

func runHealth(cfg *config.Config) error {
	return system.RunHealth(cfg)
}

func runHelp() error {
	command := ""
	args := os.Args[2:]
	for i, arg := range args {
		if arg == "--command" && i+1 < len(args) {
			command = args[i+1]
			break
		}
	}
	return system.RunHelp(command)
}

func runListExchanges() error {
	return system.RunListExchanges()
}

// Screener tools

func runTopGainers(cfg *config.Config) error {
	fs := flag.NewFlagSet("top-gainers", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "1D", "Timeframe")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return screener.RunTopGainers(cfg, *exchange, *timeframe, *limit, *futures)
}

func runTopLosers(cfg *config.Config) error {
	fs := flag.NewFlagSet("top-losers", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "1D", "Timeframe")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return screener.RunTopLosers(cfg, *exchange, *timeframe, *limit, *futures)
}

func runBollingerScan(cfg *config.Config) error {
	fs := flag.NewFlagSet("bollinger-scan", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "4h", "Timeframe")
	bbwThreshold := fs.Float64("bbw-threshold", 0.04, "Bollinger Band Width threshold")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return screener.RunBollingerScan(cfg, *exchange, *timeframe, *bbwThreshold, *limit, *futures)
}

func runRatingFilter(cfg *config.Config) error {
	fs := flag.NewFlagSet("rating-filter", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "4h", "Timeframe")
	rating := fs.Int("rating", 2, "Bollinger Band rating (-3 to 3)")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return screener.RunRatingFilter(cfg, *exchange, *timeframe, *rating, *limit, *futures)
}

// Analysis tools

func runCoinAnalysis(cfg *config.Config) error {
	fs := flag.NewFlagSet("coin-analysis", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "4h", "Timeframe")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return analysis.RunCoinAnalysis(cfg, *symbol, *exchange, *timeframe)
}

// Pattern tools

func runConsecutiveCandles(cfg *config.Config) error {
	fs := flag.NewFlagSet("consecutive-candles-scan", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "4h", "Timeframe")
	patternType := fs.String("pattern-type", "bullish", "Pattern type (bullish/bearish)")
	minGrowth := fs.Float64("min-growth", 2.0, "Minimum growth percentage")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return patterns.RunConsecutiveCandles(cfg, *exchange, *timeframe, *patternType, *minGrowth, *limit, *futures)
}

func runAdvancedCandle(cfg *config.Config) error {
	fs := flag.NewFlagSet("advanced-candle-pattern", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	baseTimeframe := fs.String("base-timeframe", "4h", "Base timeframe")
	minSizeIncrease := fs.Float64("min-size-increase", 10.0, "Minimum size increase percentage")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return patterns.RunAdvancedCandle(cfg, *exchange, *baseTimeframe, *minSizeIncrease, *limit, *futures)
}

// Volume tools

func runVolumeBreakout(cfg *config.Config) error {
	fs := flag.NewFlagSet("volume-breakout-scanner", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "4h", "Timeframe")
	volumeMultiplier := fs.Float64("volume-multiplier", 2.0, "Volume multiplier")
	priceChangeMin := fs.Float64("price-change-min", 3.0, "Minimum price change percentage")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return volume.RunVolumeBreakout(cfg, *exchange, *timeframe, *volumeMultiplier, *priceChangeMin, *limit, *futures)
}

func runVolumeConfirmation(cfg *config.Config) error {
	fs := flag.NewFlagSet("volume-confirmation-analysis", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "4h", "Timeframe")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return volume.RunVolumeConfirmation(cfg, *symbol, *exchange, *timeframe)
}

func runSmartVolume(cfg *config.Config) error {
	fs := flag.NewFlagSet("smart-volume-scanner", flag.ContinueOnError)
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	minVolumeRatio := fs.Float64("min-volume-ratio", 2.0, "Minimum volume ratio")
	minPriceChange := fs.Float64("min-price-change", 2.0, "Minimum price change percentage")
	rsiRange := fs.String("rsi-range", "any", "RSI range (any/oversold/neutral/overbought)")
	limit := fs.Int("limit", 10, "Number of results")
	futures := fs.Bool("futures", false, "Use futures/perpetual symbols")
	fs.Parse(os.Args[2:])

	return volume.RunSmartVolume(cfg, *exchange, *minVolumeRatio, *minPriceChange, *rsiRange, *limit, *futures)
}

// Multi-timeframe

func runMultiTimeframe(cfg *config.Config) error {
	fs := flag.NewFlagSet("multi-timeframe-analysis", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return mtf.RunMultiTimeframe(cfg, *symbol, *exchange)
}

// Sentiment and News

func runMarketSentiment(cfg *config.Config) error {
	fs := flag.NewFlagSet("market-sentiment", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	category := fs.String("category", "all", "Category (all/crypto/stocks)")
	limit := fs.Int("limit", 50, "Number of posts to analyze")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return sentiment.RunMarketSentiment(cfg, *symbol, *category, *limit)
}

func runFinancialNews(cfg *config.Config) error {
	fs := flag.NewFlagSet("financial-news", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol")
	category := fs.String("category", "all", "Category (all/crypto/stocks)")
	limit := fs.Int("limit", 10, "Number of articles")
	fs.Parse(os.Args[2:])

	return sentiment.RunFinancialNews(cfg, *symbol, *category, *limit)
}

func runCombinedAnalysis(cfg *config.Config) error {
	fs := flag.NewFlagSet("combined-analysis", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	timeframe := fs.String("timeframe", "4h", "Timeframe")
	category := fs.String("category", "all", "Category (all/crypto/stocks)")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return sentiment.RunCombinedAnalysis(cfg, *symbol, *exchange, *timeframe, *category)
}

// Backtesting

func runBacktest(cfg *config.Config) error {
	fs := flag.NewFlagSet("backtest-strategy", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	strategy := fs.String("strategy", "", "Strategy (required)")
	period := fs.String("period", "1y", "Period (1y, 6mo, 3mo, 1mo, etc)")
	interval := fs.String("interval", "1d", "Interval (1d, 1h, 15m, etc)")
	initialCapital := fs.Float64("initial-capital", 10000, "Initial capital")
	commissionPct := fs.Float64("commission-pct", 0.1, "Commission percentage")
	slippagePct := fs.Float64("slippage-pct", 0.05, "Slippage percentage")
	includeTradeLog := fs.Bool("include-trade-log", false, "Include trade log")
	includeEquityCurve := fs.Bool("include-equity-curve", false, "Include equity curve")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}
	if *strategy == "" {
		return fmt.Errorf("--strategy is required")
	}

	return backtest.RunBacktestStrategy(cfg, *symbol, *strategy, *period, *interval,
		*initialCapital, *commissionPct, *slippagePct, *includeTradeLog, *includeEquityCurve)
}

func runCompareStrategies(cfg *config.Config) error {
	fs := flag.NewFlagSet("compare-strategies", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	period := fs.String("period", "1y", "Period (1y, 6mo, 3mo, 1mo, etc)")
	interval := fs.String("interval", "1d", "Interval (1d, 1h, 15m, etc)")
	initialCapital := fs.Float64("initial-capital", 10000, "Initial capital")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return backtest.RunCompareStrategies(cfg, *symbol, *period, *interval, *initialCapital)
}

func runWalkForward(cfg *config.Config) error {
	fs := flag.NewFlagSet("walk-forward-backtest", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	strategy := fs.String("strategy", "", "Strategy (required)")
	period := fs.String("period", "2y", "Period (1y, 6mo, 3mo, 1mo, etc)")
	interval := fs.String("interval", "1d", "Interval (1d, 1h, 15m, etc)")
	initialCapital := fs.Float64("initial-capital", 10000, "Initial capital")
	commissionPct := fs.Float64("commission-pct", 0.1, "Commission percentage")
	slippagePct := fs.Float64("slippage-pct", 0.05, "Slippage percentage")
	nSplits := fs.Int("n-splits", 3, "Number of splits")
	trainRatio := fs.Float64("train-ratio", 0.7, "Training ratio")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}
	if *strategy == "" {
		return fmt.Errorf("--strategy is required")
	}

	return backtest.RunWalkForwardBacktest(cfg, *symbol, *strategy, *period, *interval,
		*initialCapital, *commissionPct, *slippagePct, *nSplits, *trainRatio)
}

// Yahoo Finance

func runYahooPrice(cfg *config.Config) error {
	fs := flag.NewFlagSet("yahoo-price", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return yahoo.RunYahooPrice(cfg, *symbol)
}

func runMarketSnapshot(cfg *config.Config) error {
	return yahoo.RunMarketSnapshot(cfg)
}

// Trade Planning

func runTradePlan(cfg *config.Config) error {
	fs := flag.NewFlagSet("trade-plan", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	exchange := fs.String("exchange", "NASDAQ", "Exchange name")
	timeframe := fs.String("timeframe", "1D", "Timeframe")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return plan.RunTradePlan(cfg, *symbol, *exchange, *timeframe)
}

func runFibonacci(cfg *config.Config) error {
	fs := flag.NewFlagSet("fibonacci-retracement", flag.ContinueOnError)
	symbol := fs.String("symbol", "", "Symbol (required)")
	exchange := fs.String("exchange", "BINANCE", "Exchange name")
	lookback := fs.String("lookback", "52W", "Lookback period (1M, 3M, 6M, 52W, ALL)")
	timeframe := fs.String("timeframe", "1D", "Timeframe")
	fs.Parse(os.Args[2:])

	if *symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	return plan.RunFibonacci(cfg, *symbol, *exchange, *lookback, *timeframe)
}

func notImplemented(cmd string) error {
	return fmt.Errorf("command %q is not yet implemented (coming soon)", cmd)
}

// Helper function to parse float from string
func parseFloat(s string, defaultVal float64) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return defaultVal
}
