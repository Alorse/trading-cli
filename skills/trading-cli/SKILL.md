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

### Capital and sizing (the master rules)

1. **Operating capital is FIXED at 1000 USDT.** Even if `phemex account` reports more (e.g., 7000+), treat anything beyond 1000 as untouchable reserve. Never deploy more than 1000 USDT of capital across simultaneous trades.
2. **Leverage is ALWAYS 5x.** No exceptions. Do not raise it for high-conviction setups, do not lower it for caution. Risk is managed via position sizing, not leverage.
3. **Per-trade margin = 100 USDT (10% of operating capital).** Position notional at 5x = **500 USDT**. Qty in base asset = `500 / entry_price`.
4. **Max 5 simultaneous open positions.** Total margin in use ≤ 500 USDT (50% of operating capital). The remaining 500 is mandatory buffer against adverse moves and liquidation cushion.
5. **Dynamic sizing override**: if a setup's SL distance is unusually wide (>10% from entry), reduce notional so the expected loss at SL stays ≤ 50 USDT (5% of operating capital). Compute:
   - `expected_loss = notional × sl_distance_pct`
   - If `expected_loss > 50`, set `notional = 50 / sl_distance_pct` (and `qty = notional / entry`).
   - This rarely triggers for normal scalps (SL 1-3% gives loss 5-15 USDT, well within budget).

### Mandatory pre-trade validation flow

Before placing ANY order, run these checks in order:

```bash
# 1. Get balance + count open positions
trading-cli phemex account --currency USDT > /tmp/acct.json

# Parse: total_balance = data.account.accountBalanceRv
#        used_margin = data.account.totalUsedBalanceRv
#        open_positions = count of data.positions where sizeRq != "0"

# 2. Enforce the gates
#    If open_positions >= 5: REFUSE → "Already at max 5 positions, close one first"
#    If used_margin + 100 > 500: REFUSE → "Would exceed 500 USDT margin cap"
#    If total_balance < 100: REFUSE → "Insufficient balance for 100 USDT margin"

# 3. Compute qty
qty = 500 / entry_price   # default 500 USDT notional / current price
# Apply dynamic sizing override if sl_distance > 10%

# 4. Set leverage explicitly to 5x via the position endpoint if available,
#    OR pass it inline (PHEMEX cross-margin will use up to its default — verify with `phemex positions` after).
```

### Other hard rules

6. **Default to testnet.** If `PHEMEX_API_URL` is `testnet-api.phemex.com`, proceed. If it's `api.phemex.com` (prod), REQUIRE explicit user confirmation by name in the same message before any order. Format: "About to execute on REAL money — confirm by repeating symbol and qty."
7. **No SL = no trade.** Every order must include `--sl`. If the user doesn't specify, derive from the scalp-scan `trade.stopLoss` field (default entry × 0.99 = -1%).
8. **Reduce-only on closes.** When closing an existing position with an opposite-side order, ALWAYS pass `--reduce-only` to prevent accidentally flipping the position from Long to Short or vice versa.
9. **Symbol must exist.** Before placing, confirm the symbol is in `scalp-scan` or `funding-rates --limit 200` output (it's a real PHEMEX perpetual).
10. **Never modify operating capital mid-session.** If the user asks to "use more" or "go bigger", explain the rule and request explicit re-baselining.

### Example sizing computation

User says "open the top scalp-scan candidate":
```
scalp-scan returns: symbol=BTCUSDT.P, entry=80000, stopLoss=79200 (-1%)
sl_distance_pct = (80000 - 79200) / 80000 = 0.01
expected_loss = 500 * 0.01 = 5 USDT  (within budget)
qty = 500 / 80000 = 0.00625 BTC
margin needed = 500 / 5 = 100 USDT
```

Then: pre-trade check → place at qty=0.00625, sl=79200, tp=takeProfit1 from scan.

If scalp-scan returns SL at -12% (unusual, e.g., wide setup):
```
sl_distance_pct = 0.12
expected_loss at 500 notional = 60 USDT  (exceeds 50 cap)
→ reduce notional: new_notional = 50 / 0.12 = 416.67
→ qty = 416.67 / entry
```

## Acting like a trader — not a robot

You will be invoked roughly hourly to evaluate the market and the portfolio. **Doing nothing is a valid output.** Most hours, the right move is to wait.

### Decision tree per hourly check

```
1. Read open positions (phemex positions).
   For each open position:
     a) Check current unrealized PnL (markPriceRp vs avgEntryPriceRp).
     b) Re-run multi-timeframe-analysis on its symbol.
     c) DECIDE:
        - If position is in PROFIT (PnL > 0) AND MTF turned NEUTRAL/BEARISH
          OR the 1h candle structure is reversing (e.g. lower high after impulse):
            → CLOSE NOW with reduce-only order. Don't wait for TP. Lock the profit.
        - If position is in PROFIT > +1% AND momentum still strong (RSI 60-70 rising,
          volume holding): trail by tightening SL to entry (move to break-even).
        - If position is in LOSS but not at SL yet and MTF still bullish:
          hold. Don't panic-close.
        - If MTF flipped strongly bearish on a Long position regardless of PnL:
          consider closing to free margin for a better setup. Take the small loss.

2. Read account state (phemex account → balance, used margin, open count).

3. If open positions < 5 AND used margin < 500:
   a) Run scalp-scan with min-score 65 (higher bar when account already has positions).
   b) Cross-check with funding-rates negative mode.
   c) For each candidate, run multi-timeframe-analysis. Require BUY or BUY-leaning.
   d) If a clean setup emerges, decide ENTRY TYPE:
      - Market: only if score >=75 AND 1h change is fresh (1-3%) AND momentum healthy.
      - Limit (BELOW current price): if score 65-75, place limit at BB middle or 1h EMA20.
        Set the limit valid for ~4 hours then expire. Don't chase price.
      - SKIP: anything else. Wait for next hour.

4. Cancel stale limit orders.
   Run phemex orders --symbol <sym> for each symbol you placed limits on.
   If any limit is older than 6 hours and unfilled, cancel it (the setup expired).

5. Log decision and rationale. Even "did nothing because no setup ≥65" is valuable info.
```

### Heuristics for "trend reversal while in profit"

You don't have a magic indicator for "is this reversing right now". Use a layered check:

- **Strong reversal signal (close immediately)**:
  - Position Long: 1h candle closes below entry's BB middle.
  - Volume ratio dropping below 0.8 after entry's spike.
  - RSI(1h) crossed below 50 from above on a Long.
  - Multi-timeframe-analysis flips to NEUTRAL or BEARISH bias.
- **Weak reversal signal (tighten SL to break-even)**:
  - 1h RSI starts diverging from price (price higher, RSI lower).
  - 1h volume ratio < 1.0 but no candle close yet.
  - Funding rate flipped from negative to positive while position is open.

### When to use Limit orders instead of Market

Market is "I want in NOW even if I pay spread + slippage." Use Market for:
- Tier-1 setups (scalp-scan score ≥ 75)
- Setups with negative funding + fresh momentum
- When current price is already in the recommended entry zone

Use Limit (with `--ord-type Limit --price <X>`) for:
- Tier-2 setups (score 60-75) where you want a better entry
- Symbol where BB position > 0.7 (price near upper band — wait for pullback)
- Symbol with low 1h volume ratio (let the entry come to you)

For a Long Limit, place the limit at the 1h EMA20 estimate or BB middle of the symbol's 4h chart (use `coin-analysis --timeframe 4h`).

### What NOT to do

- Don't chase a candle that already moved +5% in 1h. Late FOMO entries are scalp killers.
- Don't open a new position just because there's "available margin." The right answer most hours is "no setup, skip."
- Don't average down. If a position is at -SL, take the loss and move on.
- Don't move SL away from price to avoid a loss. Move it CLOSER to lock profits, never further.
- Don't override the leverage / capital rules even if the user phrases it as "just this once."

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
