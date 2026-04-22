# trading-cli
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Alorse/trading-cli)

A standalone Go binary for technical analysis, screening, backtesting, sentiment analysis, and news aggregation for financial markets.

Designed to be used as a CLI tool or wrapped as an MCP (Model Context Protocol) server for AI agents. **No TradingView account or authentication required** — all data sourced from public APIs.

## Features

- **Screener Tools** — Top gainers/losers, Bollinger Band scans, rating filters
- **Single Asset Analysis** — Full technical analysis with 34+ indicator groups
- **Candle Pattern Detection** — Consecutive candles, advanced multi-candle patterns
- **Volume Scanners** — Breakout detection, confirmation analysis, smart volume with RSI
- **Multi-Timeframe Analysis** — Cross-timeframe alignment and divergence detection
- **Sentiment & News** — Reddit sentiment analysis, RSS financial news, combined confluence
- **Backtesting Engine** — 6 strategies (RSI, Bollinger, MACD, EMA Cross, Supertrend, Donchian) with walk-forward analysis
- **Yahoo Finance** — Real-time quotes and global market snapshots
- **Trade Planning** — Stock scoring, trade setup, Fibonacci retracement

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap alorse/homebrew-tap
brew install trading-cli
```

### Go install

```bash
go install github.com/alorse/trading-cli/cmd/trading-cli@latest
```

### Binary download

Download pre-built binaries for your platform from the [releases page](https://github.com/alorse/trading-cli/releases/latest). Available for Linux, macOS, and Windows (amd64 and arm64).

## Documentation

- **[Tools Reference](docs/TOOLS.md)** — Complete documentation for all 25 commands with flags, examples, and output formats

## Usage

```bash
# Top gainers on KuCoin
trading-cli top-gainers --exchange KUCOIN --timeframe 15m --limit 10

# Full analysis of Bitcoin
trading-cli coin-analysis --symbol BTCUSDT --exchange KUCOIN --timeframe 1h

# Backtest RSI strategy on Apple
trading-cli backtest-strategy --symbol AAPL --strategy rsi --period 1y

# Compare all strategies on Ethereum
trading-cli compare-strategies --symbol ETH-USD --period 2y

# Reddit sentiment for Bitcoin
trading-cli market-sentiment --symbol BTC --category crypto

# Market snapshot
trading-cli market-snapshot

# Health check
trading-cli health
```

All commands output structured JSON to stdout. Errors go to stderr with a non-zero exit code.

## Development

```bash
# Build
go build -o trading-cli ./cmd/trading-cli/

# Test
go test ./... -race -cover

# Lint
golangci-lint run
```

## License

MIT
