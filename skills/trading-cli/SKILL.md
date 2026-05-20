---
name: trading-cli
description: Crypto technical analysis, PHEMEX perpetual futures scanning, and live order execution via the `trading-cli` binary. Use when the user asks to scan markets, find scalp setups, check funding rates / open interest, view PHEMEX account balance / positions / open orders, place or cancel orders, or backtest strategies. Defaults to TESTNET for safety.
metadata: {"nanobot":{"emoji":"📈","requires":{"bins":["trading-cli"]}}}
---

# trading-cli

Go CLI at `/usr/local/bin/trading-cli`. All commands print JSON to stdout; stderr is for diagnostics like `# base=...` lines.

## When to invoke

- "scan", "find me", "what's moving", "scalp", "entry now" → scanners
- "funding", "open interest", "OI" → perp metrics
- "balance", "my positions", "my orders" → account state
- "open / place / buy / sell / long / short" → order execution
- "cancel", "close position" → order management
- "backtest", "compare strategies" → historical testing

## Configuration

Credentials live in `~/.picoclaw/.env` (already set):
```
PHEMEX_API_ID, PHEMEX_API_SECRET, PHEMEX_API_URL
```
Default URL is testnet. If `PHEMEX_API_URL` contains `api.phemex.com` (no `testnet-` prefix), require explicit user confirmation by repeating symbol and qty in the same message before any order — that's real money.

Symbol format:
- Scanners / TradingView: `BTCUSDT.P`
- PHEMEX auth API: `BTCUSDT` (no `.P`)
- Translate when chaining commands.

## Hard rules — non-negotiable

Refuse to execute orders that violate any of these.

| # | Rule | Value |
|---|---|---|
| 1 | Operating capital | **1000 USDT fixed** (ignore excess balance) |
| 2 | Leverage | **5x always**, no exceptions |
| 3 | Per-trade margin | **100 USDT** (10% of operating cap) |
| 4 | Per-trade notional | **500 USDT** = `100 × 5` |
| 5 | Max simultaneous positions | **5** |
| 6 | Max margin in use total | **500 USDT** (50% of cap) |
| 7 | Max loss cap per trade | **50 USDT**. If `sl_distance > 10%`, shrink notional to `50 / sl_distance` |
| 8 | SL is mandatory | every order needs `--sl` |
| 9 | Closes use `--reduce-only` | always, no exception |
| 10 | Symbol must be a PHEMEX perp | confirm in `scalp-scan` or `funding-rates` output |
| 11 | Don't change capital mid-session | if user asks, explain rule + require explicit re-baseline |

### Pre-trade validation (canonical flow)

```bash
trading-cli phemex account --currency USDT
```
Then parse:
- `data.account.accountBalanceRv` → balance
- `data.account.totalUsedBalanceRv` → margin in use
- `data.positions[]` where `sizeRq != "0"` → open positions count

Gates:
- `open_count ≥ 5` → REFUSE, ask user to close one first
- `used_margin + 100 > 500` → REFUSE, would exceed cap
- `balance < 100` → REFUSE, insufficient balance

Then:
```
qty = 500 / entry_price       # base asset units
if sl_distance > 0.10:        # rare
    notional = 50 / sl_distance
    qty = notional / entry_price
```

**Leverage must be set explicitly to 5x cross on every symbol before the first order on it.** PHEMEX defaults new symbols to 10x cross. Run once per symbol:
```bash
trading-cli phemex set-leverage --symbol BTCUSDT --leverage 5 --margin cross
```
After this, all future positions on that symbol will be 5x cross until reset. Verify with `phemex positions`: `leverageRr` should be `"-5"` (negative = cross, absolute = 5x). If `-10` is shown, re-run `set-leverage` before placing the order.

## Commands

```text
Market data (no auth):
  funding-rates [--mode extreme|positive|negative] [--limit 25] [--min-turnover N]
  open-interest [--limit 25] [--min-turnover N]
  scalp-scan    [--limit 15] [--min-score 50] [--min-turnover N]
  top-gainers   --exchange PHEMEX --timeframe 1h --futures
  bollinger-scan --exchange PHEMEX --timeframe 4h --futures
  smart-volume-scanner --exchange PHEMEX --min-volume-ratio 1.5 --min-price-change 2

Per-symbol analysis (no auth):
  coin-analysis            --symbol BTCUSDT.P --exchange PHEMEX --timeframe 4h
  multi-timeframe-analysis --symbol BTCUSDT.P --exchange PHEMEX
  trade-plan               --symbol BTCUSDT.P --exchange PHEMEX --timeframe 4h
  combined-analysis        --symbol BTCUSDT.P --exchange PHEMEX --timeframe 1h

PHEMEX trading (auth):
  phemex account                                  # balance + positions snapshot
  phemex positions                                # filter sizeRq != 0 for open
  phemex orders --symbol BTCUSDT                  # active orders for a symbol
  phemex place-order --symbol BTCUSDT --side Buy --pos-side Long --ord-type Market \
                     --qty 0.001 --sl 76000 --tp 79000 [--ord-type Limit --price X]
                     [--reduce-only] [--cl-ord-id ID] [--tif GoodTillCancel]
  phemex cancel --symbol BTCUSDT --order-id <id> --pos-side Long
  phemex set-leverage --symbol BTCUSDT --leverage 5 --margin cross  # run once per new symbol

Backtest (Yahoo, only for reference / strategy R&D):
  compare-strategies --symbol BTC-USD --interval 1h --period 6mo
  backtest-strategy  --symbol BTC-USD --strategy wvf --interval 1d --period 1y --include-trade-log
```

## Hourly decision tree

You'll be invoked roughly hourly. **Doing nothing is a valid output.** Most hours, no action is the right call.

```
A. MANAGE OPEN POSITIONS
   For each position where sizeRq != 0:
     - Compute current PnL: (markPriceRp - avgEntryPriceRp) * sizeRq
     - Run: multi-timeframe-analysis --symbol <SYM>USDT.P --exchange PHEMEX
     - DECIDE:
       * In PROFIT + MTF flipped NEUTRAL/BEARISH         → CLOSE NOW (don't wait for TP)
       * In PROFIT > +1% + momentum still strong         → Tighten SL to entry (break-even)
       * In LOSS, not at SL, MTF still bullish           → Hold
       * MTF strongly BEARISH regardless of PnL          → Close, free margin for better setup

B. EVALUATE NEW ENTRIES
   Only if open_count < 5 AND used_margin < 500:
   1. trading-cli scalp-scan --limit 10 --min-score 65 --min-turnover 2000000
   2. trading-cli funding-rates --mode negative --limit 10 --min-turnover 1000000
      (cross-check: a symbol in BOTH lists doubles conviction)
   3. For top 1-3 candidates: multi-timeframe-analysis. Require BUY or lean BUY.
   4. Choose entry type per setup quality:
      - Market:  score ≥75 AND 1h change in 1-3% range AND momentum healthy
      - Limit:   score 65-75 → place below current at 1h EMA20 / BB middle, 4h expiry
      - SKIP:    anything else; wait next hour
   5. BEFORE placing the first order on a NEW symbol (one you haven't traded yet
      in this session), run set-leverage to enforce 5x cross:
        trading-cli phemex set-leverage --symbol <SYM> --leverage 5 --margin cross
      Skip this step if `phemex positions` already shows leverageRr=-5 for the symbol.

C. HOUSEKEEPING
   Cancel any Limit order older than 6 hours unfilled — setup expired.

D. LOG
   Append rationale to MEMORY.md (see section below). Even "no action; no setup ≥65" is valuable.
```

### Reversal signals (for early close in profit)

**Strong (close immediately):**
- Long position: 1h candle closes below its BB middle
- Volume ratio drops below 0.8 after entry's spike
- RSI(1h) crosses below 50 from above
- MTF bias flips to NEUTRAL/BEARISH

**Weak (move SL to break-even):**
- Bearish divergence: price higher, RSI(1h) lower
- Funding rate flips positive while in a long position
- 1h volume ratio < 1.0 but candle hasn't closed

### What NOT to do

- Don't chase a candle already +5% in 1h (late FOMO)
- Don't open a trade just because margin is available
- Don't average down on losers
- Don't widen SL to avoid loss — only ever tighten
- Don't override the hard rules even if user says "just this once"

## Output cheat sheet

**`scalp-scan` candidate (key fields):**
```
score       0-100 confluence; ≥60 actionable, ≥75 tier-1
changePct1h sweet spot 1-5% (fresh, not exhausted)
rsi1h       ideal 50-68
volumeRatio ≥1.5 ideal; <0.8 weak
fundingRatePct -0.02..0.05 healthy; <-0.05 short-squeeze bias
bbPosition  0=lower band, 1=upper; prefer <0.6
trade.{entry,stopLoss,takeProfit1,takeProfit2}  ← source of truth for SL/TP
```

**`phemex positions`:**
- Open ↔ `sizeRq != "0"` (PHEMEX pre-creates empty slots per traded symbol)
- Live PnL: `markPriceRp` vs `avgEntryPriceRp`, signed by `posSide`

**`phemex account`:**
- `data.account.accountBalanceRv` total USDT
- `data.account.totalUsedBalanceRv` margin currently locked
- Available = balance − used

**Error envelope:**
```
{"code": <N>, "msg": "<reason>", "data": null}
```
`code != 0` = logical error. The `orders` subcommand translates `code 10002 OM_ORDER_NOT_FOUND` to an empty rows list automatically; other endpoints surface it.

## Memory — MEMORY.md

You have a writable memory file at `/home/alorse/.picoclaw/workspace/skills/trading-cli/MEMORY.md`. It already exists with a starter structure. **Use it freely** to:
- Log every decision per hourly run (even "no action")
- Track each opened trade with rationale, entry conditions, and outcome on close
- Note recurring market regimes (e.g., "BTC stuck in range past 3 days", "alts pumping on Asia open")
- Record setups you considered and rejected, with the reason — patterns emerge over time
- Remember user preferences or one-off instructions ("user said skip MEME tokens")

Read it at the start of every hourly run. Append at the end. Trim entries older than 30 days when the file grows past ~200 lines.

Edit it as a regular file (e.g., with `cat >>`, `sed`, or your editor of choice). It's plain markdown.

## Out of scope

- Automatic execution loops (you're invoked per hour, not continuously)
- Portfolio rebalancing
- Spot trading, options, inverse contracts (only USDT perps via `/g-*` endpoints)
- Price alerts / push notifications

## Troubleshooting

- `PHEMEX_API_ID/PHEMEX_API_SECRET not set` → `~/.picoclaw/.env` not sourced
- `phemex error code=10002` → no active orders, gracefully handled in `orders` command
- `phemex http 401` / signature errors → IP whitelist on PHEMEX API key may need update
- `no data returned for symbol` → that perpetual doesn't exist; check `funding-rates --limit 200`
- `place-order` fails with symbol → drop `.P` suffix; PHEMEX auth API doesn't use it
