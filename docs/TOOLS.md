# Tools Reference

Complete reference for all 25 commands in trading-cli. All commands output structured JSON to stdout. Errors go to stderr with a non-zero exit code.

## Quick Reference

| Category | Command | Description |
|----------|---------|-------------|
| **Screening** | `top-gainers` | Top gaining symbols by exchange |
| | `top-losers` | Top losing symbols by exchange |
| | `bollinger-scan` | Symbols near Bollinger Band extremes |
| | `rating-filter` | Filter by TradingView recommendation rating |
| **Analysis** | `coin-analysis` | Full technical analysis for one symbol |
| | `multi-timeframe-analysis` | Cross-timeframe trend alignment |
| **Patterns** | `consecutive-candles-scan` | Consecutive bullish/bearish candle detection |
| | `advanced-candle-pattern` | Multi-candle pattern scoring (hammer, engulfing, etc.) |
| **Volume** | `volume-breakout-scanner` | Unusual volume with price movement |
| | `volume-confirmation-analysis` | Volume vs price confirmation for one symbol |
| | `smart-volume-scanner` | Volume + RSI + Bollinger combined scan |
| **Sentiment** | `market-sentiment` | Reddit sentiment analysis |
| | `financial-news` | RSS news aggregation |
| | `combined-analysis` | Technical + sentiment + news confluence |
| **Backtesting** | `backtest-strategy` | Single strategy backtest |
| | `compare-strategies` | Compare all 6 strategies |
| | `walk-forward-backtest` | Walk-forward robustness analysis |
| **Yahoo Finance** | `yahoo-price` | Real-time quote from Yahoo Finance |
| | `market-snapshot` | Multi-asset market overview |
| **Planning** | `fibonacci-retracement` | Fibonacci levels with golden pocket |
| | `trade-plan` | 100-point stock scoring and trade setup |
| **System** | `list-exchanges` | List supported exchanges and timeframes |
| | `health` | API connectivity check |
| | `version` | Binary version info |
| | `help` | List commands or show detailed help |

---

## Supported Exchanges

### Crypto
`KUCOIN`, `BINANCE`, `BYBIT`, `OKX`, `BITGET`, `COINBASE`, `GATE`, `MEXC`, `HTX`, `BITFINEX`, `BINGX`, `PHEMEX`, `KRAKEN`

### Stocks
`NASDAQ`, `NYSE`

### Supported Timeframes
`5m`, `15m`, `1h`, `4h`, `1D`, `1W`, `1M`

---

## Futures / Perpetual Support

All screening and scanning commands support an optional `--futures` flag to switch from spot markets to futures/perpetual markets. When enabled, the tool loads symbols from the embedded `{exchange}_futures.txt` lists (populated via `fetch-symbols --futures`).

### Commands supporting `--futures`

- `top-gainers`
- `top-losers`
- `bollinger-scan`
- `rating-filter`
- `consecutive-candles-scan`
- `advanced-candle-pattern`
- `volume-breakout-scanner`
- `smart-volume-scanner`

### Examples

```bash
# Top gainers on Binance Futures
trading-cli top-gainers --exchange BINANCE --timeframe 1D --limit 10 --futures

# Volume breakout scan on Bybit perpetuals
trading-cli volume-breakout-scanner --exchange BYBIT --timeframe 4h --futures

# Bollinger scan on OKX futures
trading-cli bollinger-scan --exchange OKX --timeframe 4h --futures
```

### Symbol notation

Futures symbols from TradingView use the `.P` suffix (e.g. `BINANCE:BTCUSDT.P`). You can also pass `.P` symbols directly to single-symbol commands:

```bash
trading-cli coin-analysis --symbol BTCUSDT.P --exchange BINANCE --timeframe 4h
```

### Updating futures symbol lists

```bash
# Download futures symbols for all exchanges
go run ./cmd/fetch-symbols --futures pkg/tools/screener/data

# Or download spot (default)
go run ./cmd/fetch-symbols pkg/tools/screener/data
```

---

## Screening Tools

### top-gainers

Top gaining symbols sorted by price change percentage.

```bash
trading-cli top-gainers --exchange BINANCE --timeframe 1D --limit 10
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `1D` | Candle timeframe |
| `--limit` | `10` | Number of results |
| `--futures` | `false` | Use futures/perpetual symbols instead of spot |

**Example output:**
```json
[
  {
    "symbol": "BINANCE:DENTUSDT",
    "changePercent": 45.6,
    "price": { "open": 0.0012, "close": 0.00175, "high": 0.0018, "low": 0.0011 },
    "volume": 523456789
  }
]
```

---

### top-losers

Top losing symbols sorted by negative price change.

```bash
trading-cli top-losers --exchange BINANCE --timeframe 1h --limit 10
```

**Flags:** Same as `top-gainers`.

---

### bollinger-scan

Find symbols with Bollinger Band width below a threshold (potential breakout candidates).

```bash
trading-cli bollinger-scan --exchange BINANCE --timeframe 4h --bbw-threshold 0.10
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `4h` | Candle timeframe |
| `--bbw-threshold` | `0.05` | Maximum Bollinger Band width (lower = more compressed) |
| `--limit` | `10` | Number of results |
| `--futures` | `false` | Use futures/perpetual symbols instead of spot |

---

### rating-filter

Filter symbols by TradingView aggregate recommendation rating.

```bash
trading-cli rating-filter --exchange BINANCE --timeframe 4h --rating 2
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `4h` | Candle timeframe |
| `--rating` | `2` | Minimum recommendation rating (-3 to 3, positive = bullish) |
| `--limit` | `10` | Number of results |
| `--futures` | `false` | Use futures/perpetual symbols instead of spot |

---

## Analysis Tools

### coin-analysis

Full technical analysis for a single symbol. Returns **23 indicator groups** sourced from the TradingView scanner API, with locally computed derived fields (signals, positions, trend scores).

```bash
trading-cli coin-analysis --symbol BTCUSDT --exchange BINANCE --timeframe 4h
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Symbol, e.g. BTCUSDT, AAPL |
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `4h` | Candle timeframe |

**Indicator groups:**

| Group | Fields | Source | Notes |
|-------|--------|--------|-------|
| Price | open, high, low, close, change%, volume | TradingView | |
| RSI | value, signal, previous | TradingView + calculated | signal computed locally |
| MACD | line, signal, histogram | TradingView + calculated | histogram computed locally; crossover field not present |
| SMA | 10, 20, 50, 100, 200 | TradingView | SMA30 fetched but not output; no cross detection |
| EMA | 9, 20, 50, 100, 200 | TradingView | EMA10/30 not fetched; no cross detection |
| Bollinger Bands | upper, middle, lower, width, position | TradingView + calculated | width and position computed locally; squeeze not implemented |
| ATR | value | TradingView | % of price and volatility label not implemented |
| ADX | value | TradingView | +DI/-DI not implemented |
| Volume | current, ratio | TradingView + calculated | avg20 hardcoded to 0; no signal field |
| Stochastic | %K, %D | TradingView | |
| CCI | value, signal | TradingView + calculated | signal computed locally |
| Williams %R | value | TradingView | |
| Awesome Oscillator | value | TradingView | |
| Momentum | value | TradingView | |
| Parabolic SAR | value | TradingView | |
| Ichimoku | baseLine | TradingView | full cloud not implemented |
| Hull MA | value | TradingView | |
| Stochastic RSI | K | TradingView | D value not implemented |
| Ultimate Oscillator | value | TradingView | |
| VWAP | value | TradingView | |
| VWMA | value | TradingView | |
| Recommendations | all, MA, other | TradingView | |
| Market Structure | trend, trendScore, momentumAlignment | Calculated locally | candle analysis not implemented |

**Not implemented:** OBV, support/resistance pivot levels (fetched but not output), stock score, trade setup, and trade quality (listed in requirements for stock exchanges only).

---

### multi-timeframe-analysis

Analyzes 5 fixed timeframes (1W, 1D, 4h, 1h, 15m) simultaneously to detect trend alignment and divergences. The command uses the same scanner data for all timeframes (TradingView scanner does not differentiate by timeframe).

```bash
trading-cli multi-timeframe-analysis --symbol ETHUSDT --exchange BINANCE
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Symbol |
| `--exchange` | `BINANCE` | Exchange name |

**Per-timeframe evaluation:**

| Timeframe | Signals Evaluated | Bias Logic |
|-----------|-------------------|------------|
| **1W** | EMA100/200 trend, MACD momentum, RSI | Bullish: 2+ of EMA100>EMA200, MACD line>signal, RSI>50 |
| **1D** | Golden/death cross, RSI zone, volume ratio, MACD | Bullish: score >= 2 from close>EMA50>EMA200, RSI>60, relVolume>1.0, MACD bullish |
| **4h** | EMA20/50 alignment, MACD crossover | Bullish: EMA20>EMA50 and MACD line>signal |
| **1h** | EMA20 support/resistance, volume spikes, VWAP | Bullish: 2+ of close>EMA20, relVolume>1.5, close>VWAP |
| **15m** | EMA9/20 alignment, VWAP | Bullish: EMA9>EMA20 and close>VWAP |

**Key output fields:**
- `overallSignal` — LEAN BULLISH / LEAN BEARISH / NEUTRAL
- `action` — STRONG BUY / CAUTIOUS BUY / HOLD / CAUTIOUS SELL / STRONG SELL
- `alignedTimeframes` — Timeframes sharing the same trend
- `divergentTimeframes` — Timeframes with conflicting signals

---

## Pattern Tools

### consecutive-candles-scan

Scans all symbols on an exchange for consecutive bullish or bearish candles with pattern strength scoring.

```bash
trading-cli consecutive-candles-scan --exchange BINANCE --timeframe 4h --pattern-type bullish --min-growth 2.0
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `4h` | Candle timeframe |
| `--pattern-type` | `bullish` | Pattern type: `bullish` or `bearish` |
| `--min-growth` | `2.0` | Minimum growth percentage to qualify |
| `--limit` | `10` | Number of results |
| `--futures` | `false` | Use futures/perpetual symbols instead of spot |

---

### advanced-candle-pattern

Detects advanced multi-candle patterns (hammer, engulfing, doji, etc.) with a 0-7 scoring system. Patterns with score >= 3 qualify.

```bash
trading-cli advanced-candle-pattern --exchange BINANCE --base-timeframe 4h --min-size-increase 5.0
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--exchange` | `BINANCE` | Exchange name |
| `--base-timeframe` | `4h` | Candle timeframe |
| `--min-size-increase` | `10.0` | Minimum size increase percentage |
| `--limit` | `10` | Number of results |
| `--futures` | `false` | Use futures/perpetual symbols instead of spot |

---

## Volume Tools

### volume-breakout-scanner

Scans 500 symbols per exchange for unusual volume combined with significant price movement.

```bash
trading-cli volume-breakout-scanner --exchange BINANCE --timeframe 1h --volume-multiplier 2.0 --price-change-min 3.0
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `4h` | Candle timeframe |
| `--volume-multiplier` | `2.0` | Minimum volume ratio vs 10-day average |
| `--price-change-min` | `3.0` | Minimum price change percentage |
| `--limit` | `10` | Number of results |
| `--futures` | `false` | Use futures/perpetual symbols instead of spot |

---

### volume-confirmation-analysis

Analyzes whether volume confirms or diverges from price action for a single symbol.

```bash
trading-cli volume-confirmation-analysis --symbol BTCUSDT --exchange BINANCE --timeframe 4h
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Symbol |
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `4h` | Candle timeframe |

**Key output fields:**
- `volume.ratio` — Current volume / 10-day average
- `volume.strength` — STRONG / MODERATE / WEAK
- `signals` — List of detected signals
- `assessment` — BULLISH / BEARISH / NEUTRAL

---

### smart-volume-scanner

Combines volume breakout detection with RSI filtering and Bollinger Band position for high-confidence signals.

```bash
trading-cli smart-volume-scanner --exchange BINANCE --min-volume-ratio 2.0 --min-price-change 3.0 --rsi-range oversold
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--exchange` | `BINANCE` | Exchange name |
| `--min-volume-ratio` | `2.0` | Minimum volume ratio |
| `--min-price-change` | `2.0` | Minimum price change percentage |
| `--rsi-range` | `any` | RSI filter: `any`, `oversold`, `neutral`, `overbought` |
| `--limit` | `10` | Number of results |
| `--futures` | `false` | Use futures/perpetual symbols instead of spot |

---

## Sentiment & News

### market-sentiment

Analyzes Reddit posts for sentiment around a symbol. Returns bullish/bearish/neutral counts and overall score.

```bash
trading-cli market-sentiment --symbol BTC
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Symbol to analyze (without USDT suffix) |

**Key output fields:**
- `sentimentScore` — -1.0 (bearish) to 1.0 (bullish)
- `sentimentLabel` — Strongly Bullish / Bullish / Neutral / Bearish / Strongly Bearish
- `postsAnalyzed` — Number of posts processed
- `bullishCount` / `bearishCount` / `neutralCount`

---

### financial-news

Fetches financial news from RSS feeds filtered by category.

```bash
trading-cli financial-news --category crypto
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--category` | `general` | News category: `general`, `crypto`, `forex`, `business` |

---

### combined-analysis

Combines technical analysis, sentiment, and news into a single confluence report with confidence rating.

```bash
trading-cli combined-analysis --symbol BTCUSDT --exchange BINANCE --timeframe 4h --category crypto
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Symbol |
| `--exchange` | `BINANCE` | Exchange name |
| `--timeframe` | `4h` | Candle timeframe |
| `--category` | `general` | News category |

**Key output fields:**
- `confluence.techBullish` — Whether technicals are bullish
- `confluence.sentBullish` — Whether sentiment is bullish
- `confluence.confidence` — STRONG BUY / BUY / MIXED / SELL / STRONG SELL
- `confluence.recommendation` — Human-readable summary

---

## Backtesting

Backtesting uses Yahoo Finance historical data. Transaction costs (commission and slippage percentages) are applied correctly per trade. Use Yahoo-compatible symbols:
- **Crypto**: `BTC-USD`, `ETH-USD`
- **Stocks**: `AAPL`, `GOOGL`, `TSLA`

### Available Strategies

| Strategy | Description |
|----------|-------------|
| `rsi` | RSI oversold/overbought reversals |
| `bollinger` | Bollinger Band mean reversion |
| `macd` | MACD line/signal crossovers |
| `ema-cross` | EMA 20/50 crossover |
| `supertrend` | Supertrend direction changes |
| `donchian` | Donchian Channel breakout |

### backtest-strategy

Run a single strategy backtest with detailed metrics.

```bash
trading-cli backtest-strategy --symbol ETH-USD --strategy rsi --period 1y --interval 1d
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Yahoo Finance symbol |
| `--strategy` | (required) | Strategy name |
| `--period` | `1y` | Lookback period: `1mo`, `3mo`, `6mo`, `1y`, `2y` |
| `--interval` | `1d` | Candle interval |
| `--initial-capital` | `10000` | Starting capital |
| `--commission-pct` | `0.1` | Commission percentage |
| `--slippage-pct` | `0.05` | Slippage percentage |
| `--include-trade-log` | `false` | Include trade log in output |
| `--include-equity-curve` | `false` | Include equity curve in output |

**Key output fields:**
- `totalReturn` — Net return percentage
- `winRate` — Percentage of profitable trades
- `sharpeRatio` — Risk-adjusted return
- `maxDrawdown` — Largest peak-to-trough decline
- `profitFactor` — Gross profit / gross loss
- `totalTrades` — Number of executed trades

---

### compare-strategies

Runs all 6 strategies on the same data and ranks them.

```bash
trading-cli compare-strategies --symbol ETH-USD --period 1y
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Yahoo Finance symbol |
| `--period` | `1y` | Lookback period |
| `--interval` | `1d` | Candle interval |
| `--initial-capital` | `10000` | Starting capital |

**Key output fields:**
- `winner` — Best performing strategy name
- `rankings` — All strategies sorted by return

---

### walk-forward-backtest

Walk-forward analysis splits data into folds, trains on a portion, and tests on the remainder to detect overfitting.

```bash
trading-cli walk-forward-backtest --symbol ETH-USD --strategy rsi --period 2y --n-splits 3
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Yahoo Finance symbol |
| `--strategy` | (required) | Strategy name |
| `--period` | `2y` | Lookback period |
| `--interval` | `1d` | Candle interval |
| `--initial-capital` | `10000` | Starting capital |
| `--commission-pct` | `0.1` | Commission percentage |
| `--slippage-pct` | `0.05` | Slippage percentage |
| `--n-splits` | `3` | Number of folds |
| `--train-ratio` | `0.7` | Fraction of each fold used for training |

**Key output fields:**
- `verdict` — `ROBUST` / `MODERATE` / `WEAK` / `OVERFITTED`
- `avgRobustness` — Average test return / train return across folds
- `folds` — Per-fold train/test returns and robustness

---

## Yahoo Finance

### yahoo-price

Real-time quote from Yahoo Finance for any global asset.

```bash
trading-cli yahoo-price --symbol AAPL
trading-cli yahoo-price --symbol BTC-USD
trading-cli yahoo-price --symbol EURUSD=X
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Yahoo Finance symbol |

**Yahoo symbol formats:**
- Stocks: `AAPL`, `GOOGL`, `TSLA`
- Crypto: `BTC-USD`, `ETH-USD`
- Forex: `EURUSD=X`, `GBPUSD=X`
- Indices: `^GSPC` (S&P 500), `^DJI` (Dow Jones)

---

### market-snapshot

Global market overview with major indices, crypto, and forex.

```bash
trading-cli market-snapshot
```

No flags required.

---

## Planning Tools

### fibonacci-retracement

Computes Fibonacci retracement and extension levels with golden pocket detection. Uses Yahoo Finance historical data for swing detection.

```bash
trading-cli fibonacci-retracement --symbol ETH-USD --exchange BINANCE --lookback 6mo --timeframe 1D
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Symbol |
| `--exchange` | `BINANCE` | Exchange name |
| `--lookback` | `52W` | Lookback period (`1M`, `3M`, `6M`, `52W`, `ALL`) |
| `--timeframe` | `1D` | Candle timeframe |

**Key output fields:**
- `trend` — uptrend / downtrend
- `swingHigh` / `swingLow` — Detected swing points
- `retracementLevels` — 0.236, 0.382, 0.5, 0.618, 0.786, 0.886, 0.945
- `extensionLevels` — 1.272, 1.414, 1.618
- `goldenPocket` — 0.618-0.786 zone with `inZone` flag

---

### trade-plan

Generates a 100-point stock score, ATR-based trade setup (entry, stop-loss, targets), and quality assessment.

```bash
trading-cli trade-plan --symbol AAPL --exchange NASDAQ --timeframe 1D
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--symbol` | (required) | Symbol |
| `--exchange` | `NASDAQ` | Exchange name |
| `--timeframe` | `1D` | Candle timeframe |

**Key output fields:**
- `score` — 0-100 score
- `grade` — A / B / C / D / F
- `verdict` — QUALIFIED / REVIEW / AVOID
- `setup.entry` / `setup.stopLoss` / `setup.targets` — ATR-based levels

---

## System Tools

### list-exchanges

List all supported exchanges, their type (crypto/stock), and available timeframes.

```bash
trading-cli list-exchanges
```

---

### health

Check connectivity to TradingView and Yahoo Finance APIs.

```bash
trading-cli health
```

---

### version

Display binary version and Go version.

```bash
trading-cli version
```

---

### help

List all commands or show detailed help for a specific command.

```bash
trading-cli help
trading-cli help --command coin-analysis
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--command` | — | Command name for detailed help |

---

## Output Format

All commands output JSON to stdout:

- **Successful responses**: JSON object or array
- **Errors**: Error message to stderr with non-zero exit code
- **Empty results**: `[]` for arrays (never `null`)

## Data Sources

| Source | Used by | Notes |
|--------|---------|-------|
| TradingView Scanner API | Screener, analysis, pattern, volume tools | Public API, no auth required |
| Yahoo Finance | Backtesting, price, snapshot, fibonacci, trade-plan | Public API, no auth required |
| Reddit (public JSON) | market-sentiment | Public RSS/JSON, no auth required |
| RSS Feeds | financial-news | CoinDesk, etc. |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error (invalid flags, API failure, no data) |
