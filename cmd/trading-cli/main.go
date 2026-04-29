package main

import (
	"fmt"
	"os"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/mcp"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	cfg := config.Load()

	var err error
	switch cmd {
	case "version":
		err = runVersion()
	case "health":
		err = runHealth(cfg)
	case "help":
		err = runHelp()
	case "top-gainers":
		err = runTopGainers(cfg)
	case "top-losers":
		err = runTopLosers(cfg)
	case "bollinger-scan":
		err = runBollingerScan(cfg)
	case "rating-filter":
		err = runRatingFilter(cfg)
	case "coin-analysis":
		err = runCoinAnalysis(cfg)
	case "consecutive-candles-scan":
		err = runConsecutiveCandles(cfg)
	case "advanced-candle-pattern":
		err = runAdvancedCandle(cfg)
	case "volume-breakout-scanner":
		err = runVolumeBreakout(cfg)
	case "volume-confirmation-analysis":
		err = runVolumeConfirmation(cfg)
	case "smart-volume-scanner":
		err = runSmartVolume(cfg)
	case "multi-timeframe-analysis":
		err = runMultiTimeframe(cfg)
	case "market-sentiment":
		err = runMarketSentiment(cfg)
	case "financial-news":
		err = runFinancialNews(cfg)
	case "combined-analysis":
		err = runCombinedAnalysis(cfg)
	case "backtest-strategy":
		err = runBacktest(cfg)
	case "compare-strategies":
		err = runCompareStrategies(cfg)
	case "walk-forward-backtest":
		err = runWalkForward(cfg)
	case "yahoo-price":
		err = runYahooPrice(cfg)
	case "market-snapshot":
		err = runMarketSnapshot(cfg)
	case "trade-plan":
		err = runTradePlan(cfg)
	case "fibonacci-retracement":
		err = runFibonacci(cfg)
	case "list-exchanges":
		err = runListExchanges()
	case "mcp":
		if err := mcp.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
			os.Exit(1)
		}
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `trading-cli — Technical analysis and trading tools

Usage: trading-cli <command> [flags]

Commands:
  Screeners:
    top-gainers              Top performing assets by price change
    top-losers               Worst performing assets by price change
    bollinger-scan           Bollinger Band squeeze detection
    rating-filter            Filter by Bollinger Band rating

  Analysis:
    coin-analysis            Full technical analysis for a symbol

  Patterns:
    consecutive-candles-scan Detect consecutive candle patterns
    advanced-candle-pattern  Multi-timeframe candle pattern scoring

  Volume:
    volume-breakout-scanner  Volume + price breakout detection
    volume-confirmation-analysis  Deep volume analysis for a symbol
    smart-volume-scanner     Volume scanner with RSI and recommendations

  Multi-Timeframe:
    multi-timeframe-analysis Cross-timeframe alignment analysis

  Sentiment & News:
    market-sentiment         Reddit sentiment analysis
    financial-news           RSS financial news headlines
    combined-analysis        Technical + sentiment + news confluence

  Backtesting:
    backtest-strategy        Backtest a single strategy
    compare-strategies       Compare all strategies
    walk-forward-backtest    Walk-forward overfitting detection

  Yahoo Finance:
    yahoo-price              Real-time quote
    market-snapshot          Global market overview

  Trade Planning:
    trade-plan               Generate a complete trade plan
    fibonacci-retracement    Fibonacci analysis

  System:
    list-exchanges           List available exchanges
    version                  Show version
    health                   Check API connectivity
    help                     Show this help

Use "trading-cli help --command <name>" for detailed help.
`)
}
