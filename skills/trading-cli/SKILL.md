---
name: trading-cli
description: Crypto technical analysis, PHEMEX perpetual futures scanning, and live order execution via the `trading-cli` binary. Use when the user asks to scan markets, find scalp setups, check funding rates / open interest, view PHEMEX account balance / positions / open orders, place or cancel orders, or backtest strategies. Defaults to TESTNET for safety.
metadata: {"nanobot":{"emoji":"📈","requires":{"bins":["trading-cli"]}}}
---

# trading-cli

A Go CLI for crypto technical analysis and PHEMEX perpetual-futures trading. Binary at `/usr/local/bin/trading-cli`. All commands print JSON to stdout. Stderr is for diagnostics (e.g. `# base=...` lines).

## When to use this skill

Trigger on any of these intents:
- "scan", "find me", "what's moving" — public market scanners
- "funding", "open interest", "OI" — perpetual contract metrics
- "scalp", "quick trade", "entry now" — short-term setup discovery
- "balance", "my positions", "my orders" — account state (PHEMEX)
- "open a long/short", "place order", "buy/sell" — order execution
- "cancel order", "close position" — order management
- "backtest", "compare strategies" — historical strategy testing

## Configuration

Authenticated commands (`phemex *`) require credentials. They are read from environment variables, which picoclaw should source from `~/.picoclaw/.env`:

```
PHEMEX_API_ID=<your-api-id>
PHEMEX_API_SECRET=<your-api-secret>
PHEMEX_API_URL=https://testnet-api.phemex.com   # default; switch to https://api.phemex.com for prod
```

**Safety rule:** if `PHEMEX_API_URL` contains `api.phemex.com` (without `testnet-`), warn the user explicitly that they are about to act on REAL money before placing or canceling any order.

## Commands at a glance

```text
Market data (no auth):
  funding-rates [--mode extreme|positive|negative] [--limit 25] [--min-turnover 1000000]
  open-interest [--limit 25] [--min-turnover 1000000]
  scalp-scan    [--limit 15] [--min-score 50] [--min-turnover 1000000]
  top-gainers   --exchange PHEMEX --timeframe 1h --limit 15 --futures
  top-losers    --exchange PHEMEX --timeframe 1h --limit 15 --futures
  bollinger-scan --exchange PHEMEX --timeframe 4h --futures
  smart-volume-scanner --exchange PHEMEX --min-volume-ratio 1.5 --min-price-change 2

Per-symbol analysis (no auth):
  coin-analysis --symbol BTCUSDT.P --exchange PHEMEX --timeframe 4h
  multi-timeframe-analysis --symbol BTCUSDT.P --exchange PHEMEX
  trade-plan --symbol BTCUSDT.P --exchange PHEMEX --timeframe 4h
  combined-analysis --symbol BTCUSDT.P --exchange PHEMEX --timeframe 1h

PHEMEX trading (auth required):
  phemex account                                    # balance + all positions
  phemex positions                                  # open positions only (filter sizeRq != 0)
  phemex orders --symbol BTCUSDT                    # active orders for a symbol
  phemex place-order --symbol BTCUSDT --side Buy --pos-side Long --ord-type Market --qty 0.001
                                                    # optional: --price (Limit), --sl, --tp, --reduce-only, --cl-ord-id, --tif
  phemex cancel --symbol BTCUSDT --order-id <id> --pos-side Long

Backtesting (Yahoo Finance):
  compare-strategies --symbol BTC-USD --interval 1h --period 6mo
  backtest-strategy --symbol BTC-USD --strategy wvf --interval 1d --period 1y --include-trade-log
```

Note PHEMEX **symbol formats differ**:
- TradingView / scanners use `BTCUSDT.P` (perpetual suffix)
- PHEMEX private API uses `BTCUSDT` (no suffix)
- The skill must translate when chaining commands

## The scalp workflow (the main loop)

For "find me something to trade" requests:

### Step 1 — Scan for confluence
```bash
trading-cli scalp-scan --limit 10 --min-score 60 --min-turnover 2000000
```
This returns top candidates ranked by a confluence score combining: fresh 1h momentum, volume surge vs 10d avg, RSI in bullish-not-cooked zone, BB position with room to run, healthy funding, and 24h day-bias. Each entry includes a pre-built `trade` block with entry/SL/TP1/TP2 for 5x leverage.

### Step 2 — Cross-check funding edge
```bash
trading-cli funding-rates --mode negative --limit 10 --min-turnover 1000000
```
Symbols with **negative funding while pumping** = shorts trapped, paying longs. Strong continuation bias. If a scalp-scan candidate also appears here, score doubles in conviction.

### Step 3 — Confirm with per-symbol analysis (only top 2-3)
```bash
trading-cli multi-timeframe-analysis --symbol <SYM>USDT.P --exchange PHEMEX
```
Look for `recommendation: BUY` and `alignment: MOSTLY BULLISH` or better. Skip the symbol if MTF says NEUTRAL or BEARISH.

### Step 4 — Verify account capacity
```bash
trading-cli phemex account --currency USDT
```
Read `data.account.accountBalanceRv`. Compute position sizing: at 5x leverage with -1% SL, you risk 5% of position size = 1% of account if position size = 20% account. Never use more than 60% of account across simultaneous open positions.

### Step 5 — Place the order
Use the `trade` block from scalp-scan output as the source of truth for SL/TP/leverage. Build the position size in base-asset units (PHEMEX API uses `orderQtyRq` as base asset, not USDT).

Example: BTCUSDT at $80,000, 5x leverage, want to deploy 20% of a $1000 account = $200 position. Base qty = 200 / 80000 = 0.0025 BTC.

```bash
trading-cli phemex place-order \
  --symbol BTCUSDT \
  --side Buy --pos-side Long --ord-type Market \
  --qty 0.0025 \
  --sl 79200 --tp 81600 \
  --cl-ord-id scalp-<timestamp>
```

### Step 6 — Track and exit
After placement:
- Save the returned `orderID` so you can `phemex cancel` if needed.
- Periodically check `phemex positions` to see unrealized PnL on the slot matching `symbol` and `posSide`.
- The SL/TP are server-side (passed to PHEMEX with `--sl`/`--tp`). If not set, you must monitor manually.

## Risk management — HARD rules

These are non-negotiable. **Refuse to execute orders that violate them.**

1. **Default to testnet.** If `PHEMEX_API_URL` is testnet, proceed freely. If it's prod, REQUIRE explicit user confirmation in the same message before placing any order. Say something like: "This will execute on REAL money. Confirm by repeating the symbol and qty."
2. **Max risk per trade = 1% of account.** Compute: `(entry - stopLoss) / entry * positionSize <= account * 0.01`. If the requested size exceeds this, reduce it automatically and tell the user.
3. **Max simultaneous open positions = 3.** Before opening a new one, check `phemex positions` and count where `sizeRq != 0`.
4. **No SL = no trade.** If the user wants a scalp without a stop loss, refuse and propose a sensible one (entry - 1% for 5x).
5. **Leverage cap.** Default 5x. Never go above 10x on suggestions, never above 20x even if user insists — explain liquidation math instead.
6. **Reduce-only on closes.** When closing an existing position with an opposite-side order, ALWAYS pass `--reduce-only` to prevent accidentally flipping the position.
7. **Symbol must exist.** Before placing, run `trading-cli funding-rates --limit 200` or check `scalp-scan` output to confirm the symbol is a tradable PHEMEX perpetual.

## Reading the output

### `scalp-scan` candidate
```json
{
  "symbol": "HUSDT.P",
  "score": 55,                    // 0-100 confluence; >=60 actionable
  "changePct1h": 3.08,            // fresh momentum, sweet spot 1-5%
  "rsi1h": 55.75,                 // ideal 50-68
  "volumeRatio": 0.84,            // >=1.5 ideal; <0.8 weak
  "fundingRatePct": 0.005,        // -0.02 to 0.05 = healthy
  "bbPosition": 0.72,             // 0=lower band, 1=upper; prefer <0.6
  "reasons": ["fresh-momentum", "rsi-bullish", "funding-healthy"],
  "trade": { "entry":..., "stopLoss":..., "takeProfit1":..., "takeProfit2":..., "suggestedLeverage":5 }
}
```

### `phemex positions` — the position is OPEN when:
```json
{ "symbol": "BTCUSDT", "posSide": "Long", "sizeRq": "0.001", ... }
```
`sizeRq == "0"` means the slot is empty (PHEMEX pre-creates a slot per traded symbol). Filter on `sizeRq != "0"`.

### `phemex account` — read these fields:
- `data.account.accountBalanceRv` — total USDT balance
- `data.account.totalUsedBalanceRv` — margin currently locked in positions
- Available = balance - usedBalance

### Error envelope
```json
{ "code": 10002, "msg": "OM_ORDER_NOT_FOUND", "data": null }
```
`code != 0` is a logical error. The `orders` subcommand specifically translates `10002` to an empty rows array; other endpoints return the error verbatim.

## Examples of complete sessions

### "Find me a scalp right now"
```bash
trading-cli scalp-scan --limit 5 --min-score 65 --min-turnover 3000000
# Pick top 1-2 by score
trading-cli funding-rates --mode negative --limit 10 --min-turnover 1000000
# Cross-check: any scalp-scan symbol also in negative-funding list? Higher conviction.
trading-cli multi-timeframe-analysis --symbol <SYM>USDT.P --exchange PHEMEX
# Confirm BUY/MOSTLY BULLISH
trading-cli phemex account
# Compute size from accountBalanceRv * 0.20 / lastPrice
trading-cli phemex place-order --symbol <SYM>USDT --side Buy --pos-side Long --ord-type Market --qty <computed> --sl <stopLoss from trade block> --tp <takeProfit1>
```

### "What's the state of my account?"
```bash
trading-cli phemex account --currency USDT
# Report: balance, used margin, open positions count + symbols + unrealized PnL
```

### "Cancel everything on BTCUSDT"
```bash
trading-cli phemex orders --symbol BTCUSDT
# For each row, extract orderID and posSide, then:
trading-cli phemex cancel --symbol BTCUSDT --order-id <id> --pos-side <side>
```

### "Backtest WVF on BTC"
```bash
trading-cli backtest-strategy --symbol BTC-USD --strategy wvf --interval 1d --period 1y --include-trade-log
# Available strategies: rsi, bollinger, macd, ema-cross, supertrend, donchian,
#                       stochrsi-supertrend, dmi-adx-trend, ttm-squeeze, wvf, range-filter
```

## What this skill does NOT do

- **No automatic execution loops.** This skill executes single orders or scans on demand. If the user wants a recurring monitor, they need to set up cron / a wrapper script separately.
- **No portfolio rebalancing.** Each command is independent.
- **No options or spot trading.** Only USDT-margined perpetuals via `/g-*` endpoints.
- **No inverse contracts (BTCUSD, ETHUSD).** Symbol format must be USDT-quoted.
- **No price alerts / push notifications.** The user must poll or wrap the CLI themselves.

## Troubleshooting

- `phemex: PHEMEX_API_ID/PHEMEX_API_SECRET not set` → env vars missing from picoclaw's .env
- `phemex error code=10002` on `orders` → no active orders (handled gracefully, returns empty)
- `phemex http 401` or signature errors → check IP whitelist on PHEMEX API key settings
- `no data returned for symbol` → that perpetual doesn't exist on PHEMEX (check with `funding-rates --limit 200`)
- Symbol shows in TradingView with `.P` but `place-order` fails → drop the `.P` suffix for `phemex` commands
