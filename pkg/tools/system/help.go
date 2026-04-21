package system

import "fmt"

// CommandHelp contains detailed help for each command.
var CommandHelp = map[string]string{
	"top-gainers": `top-gainers — Top performing assets by price change percentage

Flags:
  --exchange string   Exchange (default: KUCOIN)
  --timeframe string  Timeframe: 5m, 15m, 1h, 4h, 1D, 1W, 1M (default: 15m)
  --limit int         Number of results (default: 25, range: 1-50)`,
	"top-losers": `top-losers — Worst performing assets by price change percentage

Flags:
  --exchange string   Exchange (default: KUCOIN)
  --timeframe string  Timeframe (default: 15m)
  --limit int         Number of results (default: 25, range: 1-50)`,
	"bollinger-scan": `bollinger-scan — Detect Bollinger Band squeeze conditions

Flags:
  --exchange string      Exchange (default: KUCOIN)
  --timeframe string     Timeframe (default: 4h)
  --bbw-threshold float  Maximum BB width (default: 0.04)
  --limit int            Number of results (default: 50, range: 1-100)`,
	"rating-filter": `rating-filter — Filter symbols by Bollinger Band rating

Flags:
  --exchange string   Exchange (default: KUCOIN)
  --timeframe string  Timeframe (default: 5m)
  --rating int        BB rating -3 to +3 (default: 2)
  --limit int         Number of results (default: 25, range: 1-50)`,
	"coin-analysis": `coin-analysis — Full technical analysis for a single symbol

Flags:
  --symbol string    Symbol, e.g. BTCUSDT, AAPL (required)
  --exchange string  Exchange (default: KUCOIN)
  --timeframe string Timeframe (default: 15m)`,
	"consecutive-candles-scan": `consecutive-candles-scan — Detect consecutive bullish/bearish candle patterns

Flags:
  --exchange string     Exchange (default: KUCOIN)
  --timeframe string    Timeframe (default: 15m)
  --pattern-type string bullish or bearish (default: bullish)
  --candle-count int    Consecutive candles (default: 3, range: 2-5)
  --min-growth float    Minimum price change %% (default: 2.0)
  --limit int           Number of results (default: 20, range: 1-50)`,
	"advanced-candle-pattern": `advanced-candle-pattern — Multi-timeframe candle pattern scoring

Flags:
  --exchange string        Exchange (default: KUCOIN)
  --base-timeframe string  Base timeframe (default: 15m)
  --pattern-length int     Pattern length (default: 3, range: 2-4)
  --min-size-increase float Minimum size increase %% (default: 10.0)
  --limit int              Number of results (default: 15, range: 1-30)`,
	"volume-breakout-scanner": `volume-breakout-scanner — Volume + price breakout detection

Flags:
  --exchange string        Exchange (default: KUCOIN)
  --timeframe string       Timeframe (default: 15m)
  --volume-multiplier float Volume ratio threshold (default: 2.0)
  --price-change-min float  Minimum price change %% (default: 3.0)
  --limit int              Number of results (default: 25, range: 1-50)`,
	"volume-confirmation-analysis": `volume-confirmation-analysis — Deep volume analysis for a symbol

Flags:
  --symbol string    Symbol (required)
  --exchange string  Exchange (default: KUCOIN)
  --timeframe string Timeframe (default: 15m)`,
	"smart-volume-scanner": `smart-volume-scanner — Volume scanner with RSI and recommendations

Flags:
  --exchange string       Exchange (default: KUCOIN)
  --min-volume-ratio float Minimum volume ratio (default: 2.0)
  --min-price-change float Minimum price change %% (default: 2.0)
  --rsi-range string      oversold, overbought, neutral, any (default: any)
  --limit int             Number of results (default: 20, range: 1-30)`,
	"multi-timeframe-analysis": `multi-timeframe-analysis — Cross-timeframe alignment analysis

Flags:
  --symbol string    Symbol (required)
  --exchange string  Exchange (default: KUCOIN)`,
	"market-sentiment": `market-sentiment — Reddit sentiment analysis

Flags:
  --symbol string   Symbol (required)
  --category string crypto, stocks, all (default: all)
  --limit int       Posts to analyze (default: 50, range: 1-100)`,
	"financial-news": `financial-news — RSS financial news headlines

Flags:
  --symbol string   Filter by symbol (optional)
  --category string crypto, stocks, all (default: all)
  --limit int       Number of results (default: 10, range: 1-50)`,
	"combined-analysis": `combined-analysis — Technical + sentiment + news confluence

Flags:
  --symbol string    Symbol (required)
  --exchange string  Exchange (default: KUCOIN)
  --timeframe string Timeframe (default: 15m)
  --category string  crypto, stocks, all (default: all)`,
	"backtest-strategy": `backtest-strategy — Backtest a trading strategy

Flags:
  --symbol string         Yahoo Finance ticker (required)
  --strategy string       rsi, bollinger, macd, ema-cross, supertrend, donchian (required)
  --period string         1mo, 3mo, 6mo, 1y, 2y (default: 1y)
  --initial-capital float Starting capital (default: 10000)
  --commission-pct float  Commission %% (default: 0.1)
  --slippage-pct float    Slippage %% (default: 0.05)
  --interval string       1d, 1h (default: 1d)
  --include-trade-log     Include trade log
  --include-equity-curve  Include equity curve`,
	"compare-strategies": `compare-strategies — Compare all 6 strategies

Flags:
  --symbol string         Yahoo Finance ticker (required)
  --period string         1mo, 3mo, 6mo, 1y, 2y (default: 1y)
  --initial-capital float Starting capital (default: 10000)
  --interval string       1d, 1h (default: 1d)`,
	"walk-forward-backtest": `walk-forward-backtest — Walk-forward overfitting detection

Flags:
  --symbol string         Yahoo Finance ticker (required)
  --strategy string       Strategy name (required)
  --period string         1mo, 3mo, 6mo, 1y, 2y (default: 2y)
  --initial-capital float Starting capital (default: 10000)
  --commission-pct float  Commission %% (default: 0.1)
  --slippage-pct float    Slippage %% (default: 0.05)
  --n-splits int          Number of folds (default: 3, range: 2-10)
  --train-ratio float     Train/test split (default: 0.7)
  --interval string       1d, 1h (default: 1d)`,
	"yahoo-price": `yahoo-price — Real-time quote from Yahoo Finance

Flags:
  --symbol string  Symbol, e.g. AAPL, BTC-USD, ^GSPC (required)`,
	"market-snapshot": `market-snapshot — Global market overview

No flags required.`,
	"trade-plan": `trade-plan — Generate a complete trade plan

Flags:
  --symbol string    Symbol (required)
  --exchange string  Exchange (default: NASDAQ)
  --timeframe string Timeframe (default: 1D)`,
	"fibonacci-retracement": `fibonacci-retracement — Fibonacci analysis

Flags:
  --symbol string    Symbol (required)
  --lookback string  1M, 3M, 6M, 52W, ALL (default: 52W)
  --timeframe string Timeframe (default: 1D)`,
	"list-exchanges": `list-exchanges — List available exchanges and timeframes

No flags required.`,
	"version": `version — Show version information

No flags required.`,
	"health": `health — Check API connectivity

No flags required.`,
}

// RunHelp prints help for a specific command or all commands.
func RunHelp(command string) error {
	if command == "" {
		fmt.Println("Use 'trading-cli help --command <name>' for detailed help.")
		fmt.Println("\nAvailable commands:")
		for name := range CommandHelp {
			fmt.Printf("  %s\n", name)
		}
		return nil
	}

	help, ok := CommandHelp[command]
	if !ok {
		return fmt.Errorf("no help available for command: %s", command)
	}
	fmt.Println(help)
	return nil
}
