# trading-cli
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Alorse/trading-cli)

A standalone Go binary for technical analysis, screening, backtesting, sentiment analysis, and news aggregation for financial markets.

Designed to be used as a CLI tool or wrapped as an MCP (Model Context Protocol) server for AI agents. **No TradingView account or authentication required** — all data sourced from public APIs.

## Features

- **Screener Tools** — Top gainers/losers, Bollinger Band scans, rating filters
- **Single Asset Analysis** — Full technical analysis with 23 indicator groups including RSI, MACD, Bollinger Bands, and more (see docs/TOOLS.md for full list)
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

# Top gainers on Binance Futures
trading-cli top-gainers --exchange BINANCE --timeframe 15m --limit 10 --futures

# Full analysis of Bitcoin
trading-cli coin-analysis --symbol BTCUSDT --exchange KUCOIN --timeframe 1h

# Analysis of a perpetual futures symbol
trading-cli coin-analysis --symbol BTCUSDT.P --exchange BINANCE --timeframe 1h

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

## MCP Server

trading-cli can run as an MCP server over stdio, exposing all 25 tools to AI agents like Claude Code, Cursor, or any MCP-compatible client.

### Quick install

```bash
trading-cli mcp install                        # Claude Desktop (default)
trading-cli mcp install --client claude-code   # Claude Code
trading-cli mcp install --client cursor        # Cursor
trading-cli mcp install --client windsurf      # Windsurf
trading-cli mcp install --client codex         # OpenAI Codex CLI
trading-cli mcp install --client vscode        # VS Code Copilot
trading-cli mcp install --client gemini        # Gemini CLI
trading-cli mcp install --client amazon-q      # Amazon Q Developer
trading-cli mcp install --client zed           # Zed
trading-cli mcp install --client lm-studio     # LM Studio
trading-cli mcp install --client --list        # show all supported clients
```

This command writes the MCP config into your client's settings file automatically. No manual JSON editing needed.

Use `--dry-run` to preview the change or `--force` to overwrite an existing entry.

### Start the server manually

```bash
trading-cli mcp
```

The server speaks JSON-RPC 2.0 over stdin/stdout. No flags or configuration needed — just pipe stdin and read stdout.

### Manual config

If you prefer to edit the config yourself, add this to your client's MCP settings:

```json
{
  "mcpServers": {
    "trading-cli": {
      "command": "trading-cli",
      "args": ["mcp"]
    }
  }
}
```

If trading-cli is installed via Homebrew or `go install`, the binary is on `$PATH`. Otherwise, use the full path to the binary.

### Available tools

All 25 CLI commands are exposed as MCP tools with the same parameters (using `snake_case` names). See [docs/TOOLS.md](docs/TOOLS.md) for parameter details.

| Category | Tools |
|----------|-------|
| Screening | `top_gainers`, `top_losers`, `bollinger_scan`, `rating_filter` |
| Analysis | `coin_analysis`, `multi_timeframe_analysis` |
| Patterns | `consecutive_candles_scan`, `advanced_candle_pattern` |
| Volume | `volume_breakout_scanner`, `volume_confirmation_analysis`, `smart_volume_scanner` |
| Sentiment | `market_sentiment`, `financial_news`, `combined_analysis` |
| Backtesting | `backtest_strategy`, `compare_strategies`, `walk_forward_backtest` |
| Yahoo Finance | `yahoo_price`, `market_snapshot` |
| Planning | `trade_plan`, `fibonacci_retracement` |
| System | `list_exchanges`, `version`, `health` |

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
