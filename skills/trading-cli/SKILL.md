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
| 2 | Leverage (perp only) | **5x always**, no exceptions. Spot has no leverage. |
| 3 | Per-trade capital | **100 USDT** = perp margin OR spot exposure (NOT both for same trade) |
| 4 | Per-trade notional | Perp: **500 USDT** (`100 × 5`). Spot: **100 USDT** (no leverage). |
| 5 | Max simultaneous positions | **5 combined** across perp + spot |
| 6 | Max capital deployed | **500 USDT total** (perp margins + spot asset values, combined) |
| 7 | Max loss cap per trade | **50 USDT**. If `sl_distance > 10%`, shrink notional to `50 / sl_distance` |
| 8 | SL is mandatory | every order needs `--sl` (perp inline, spot via auto-Stop) |
| 9 | Closes use `--reduce-only` (perp) | always. Spot closes via opposite-side Market with `--base-qty`. |
| 10 | Symbol must exist | confirm in `scalp-scan` / `funding-rates` (perp) or `spot-orders` (spot) |
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

PHEMEX spot (fallback when perp fails — see "Spot fallback" section):
  phemex spot-balance [--currency USDT|BTC|...]            # wallet (USDT or specific asset)
  phemex spot-orders --symbol BTCUSDT                       # active spot orders (CLI adds 's')
  phemex spot-place --symbol BTCUSDT --side Buy --ord-type Market \
                    --quote-qty 100 [--sl <price>] [--tp <price>]
                                                            # CLI auto-places conditional Stop+Limit
  phemex spot-place --symbol BTCUSDT --side Sell --ord-type Market --base-qty 0.001
                                                            # to close: pass base-qty of held asset
  phemex spot-cancel --symbol BTCUSDT --order-id <id>

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
   Only if open_count < 5 AND used_margin < 500.
   See "Stale-market protocol" below for the adaptive thresholds.

   1. Read MEMORY.md → ## State (consecutive empty hours, regime flag)
                    → ## Recent rejections (last 3h, with reason)
   2. If regime flag is active AND still valid → SKIP scanning, only manage positions, log.
   3. Compute current scan parameters per the protocol (min-score, min-turnover, extra lens).
   4. trading-cli scalp-scan --limit 10 --min-score <S> --min-turnover <T>
   5. trading-cli funding-rates --mode negative --limit 10 --min-turnover 1000000
      (cross-check: a symbol in BOTH lists doubles conviction)
   5b. STANDALONE FUNDING EDGE: scan the funding-rates output for symbols with
       fundingRatePct <= -0.10 (extreme short squeeze) AND changePct24h > 0
       (avoid catching a falling knife). For up to 1 such symbol per session,
       run multi-timeframe-analysis. If MTF is BUY or MOSTLY BULLISH, this
       is a valid entry path REGARDLESS of scalp-scan score (the funding
       extreme is the edge). Use Limit entry at last_price * 0.995 to demand
       a small pullback. Log: "funding-edge entry on SYM, funding=N% MTF=BUY".
   6. For each candidate, BEFORE deep analysis:
      - If symbol is in ## Recent rejections (last 3h) with the SAME blocking reason
        AND its current score did NOT improve by ≥5 points → skip it.
        Log: "skipped re-analysis of SYM, still blocked by X"
   7. For remaining candidates, run multi-timeframe-analysis. Require BUY or lean BUY.
   8. Choose entry type per setup quality:
      - MTF override: if MTF is 5/5 STRONG BUY (perfect alignment all timeframes)
        AND score ≥55, proceed as Limit entry. Cap to 1 such override per session.
        Log: "MTF=5/5 override allowed score=N entry".
      - Market: score ≥75 AND 1h change in 1-3% range AND momentum healthy
      - Limit:  score 65-75 → place below current at 1h EMA20 / BB middle, 4h expiry
      - SKIP:   anything else; record rejection reason in MEMORY.md ## Recent rejections
   9. BEFORE placing the first order on a NEW symbol (one without leverageRr=-5
      in current `phemex positions`), run set-leverage:
        trading-cli phemex set-leverage --symbol <SYM> --leverage 5 --margin cross
   9b. SPOT FALLBACK: if set-leverage errors with TE_SYMBOL_NOT_IN_RANGE OR
       a subsequent place-order fails with TE_SYMBOL_NOT_IN_RANGE /
       TE_CANNOT_COVER_ESTIMATE_ORDER_LOSS / HTTP 5xx, try spot instead:
         a) Verify spot pair exists: `phemex spot-orders --symbol <SYM>` returns
            code 0 (even with empty rows). If it errors, the spot pair doesn't
            exist either — SKIP and log rejection.
         b) Require score ≥70 for spot entries (higher bar — spot lacks the
            funding edge that perp can use).
         c) Place via spot (NO set-leverage step on spot):
              trading-cli phemex spot-place --symbol <SYM> --side Buy
                --ord-type Market --quote-qty 100 --sl <price> --tp <price>
            The CLI auto-places the conditional Stop Sell (SL) and Limit Sell (TP).
         d) Spot sizing: quote-qty = 100 USDT direct exposure (NO leverage).
            Expected loss at SL 1% = $1, at SL 5% = $5. Naturally below the $50 cap.
         e) Log: "fallback to spot for SYM, perp error: <X>"
  10. If a trade was opened: reset ## State counter to 0; clear regime flag if set.
      If no trade opened: increment ## State counter by 1.

C. HOUSEKEEPING
   Cancel any Limit order older than 6 hours unfilled — setup expired.

D. LOG
   Append rationale to MEMORY.md (see section below). Even "no action; no setup ≥65" is valuable.
```

### Stale-market protocol — when nothing qualifies hour after hour

Read `## State` from MEMORY.md at the start of each run. The counter tracks consecutive hours without opening a trade. Adapt scan parameters according to this table:

| Consecutive empty hours | min-score | min-turnover | Extra scan |
|---|---|---|---|
| 0 (just acted) — 1 | 65 | 2,000,000 | none |
| 2 — 3 | **60** | 2,000,000 | none |
| 4 — 5 | **55** | 2,000,000 | also run `funding-rates --mode negative` as standalone lens |
| 6 — 7 | **50** (floor) | **500,000** (small caps) | funding lens + flag potential dead regime |
| ≥ 8 | **50** | 500,000 | set regime flag (see below) |

**Floor is 50. Never lower it.** Better to stay in cash 12 hours than take a 45-score trade.

#### Regime flag (capa 3)

When consecutive empty hours reaches 8, write into MEMORY.md `## State`:
```
regimeFlag: dead
regimeFlagSince: 2026-05-20 23:00 UTC
regimeFlagReason: 8h consecutive empty + BTC ranging tight + funding broadly neutral
```

While the flag is set:
- Skip the scan portion of the hourly run entirely (saves tokens)
- Still manage open positions (Section A of decision tree)
- Still cancel stale limit orders (Section C)
- Re-evaluate the flag once per run via:
  ```bash
  trading-cli coin-analysis --symbol BTCUSDT.P --exchange PHEMEX --timeframe 4h
  ```
  Clear the flag (and reset counter to 0) if:
  - BTC 4h breaks out of its prior range (price > prior high or < prior low by >1%), OR
  - BTC 4h ATR doubles vs the value when flag was set, OR
  - 6 hours have passed since `regimeFlagSince` (forced re-scan to test conditions)

#### Memory dedup rules (capa 1)

When a candidate is rejected in step 8 of the decision tree, append to `## Recent rejections`:
```
- SYMBOL @ YYYY-MM-DD HH:MM UTC — score N — reason: <short text>
```

Each hour, prune entries older than 3 hours from this list.

When evaluating new candidates next hour, the dedup check is:
- Same symbol, last 3h: skip deep analysis unless current score improved by ≥5 points
- Otherwise just log: "skipped re-analysis of SYM (still N, still blocked by X)"

This prevents burning analysis cycles on the same 2-3 obvious-no candidates.

### Spot fallback — when perpetual fails or symbol doesn't exist as perp

PHEMEX testnet perpetuals have shown unreliability: `TE_SYMBOL_NOT_IN_RANGE` (e.g. TIAUSDT, INJUSDT, AKTUSDT not listed as perp testnet) and `TE_CANNOT_COVER_ESTIMATE_ORDER_LOSS` (e.g. SOLUSDT testnet bug). Spot is the plan B.

**When fallback triggers (step 9b of decision tree):**
- `set-leverage` returns error with `TE_SYMBOL_NOT_IN_RANGE`
- `place-order` returns `TE_SYMBOL_NOT_IN_RANGE`, `TE_CANNOT_COVER_ESTIMATE_ORDER_LOSS`, or HTTP 5xx

**Differences from perpetual flow:**

| Aspect | Perp | Spot |
|---|---|---|
| Symbol API format | `BTCUSDT` | `sBTCUSDT` (CLI adds `s` prefix automatically) |
| Leverage | 5x cross (set explicitly) | None — no `set-leverage` step on spot |
| Sizing | $100 margin × 5x = $500 notional | $100 direct exposure (no leverage) |
| Expected loss at 1% SL | $5 | $1 |
| Expected loss at 5% SL | $25 | $5 |
| SL/TP placement | Inline in order params | CLI orchestrates 3 separate orders (Buy + Stop Sell + Limit Sell) |
| Shorting | Yes (`--pos-side Short`) | NO — spot is Buy then Sell only |
| Funding rate edge | Available | Not available — score floor raised to ≥70 for spot |
| Capital tracking | `positionMarginRv` per position | USDT-equivalent of held asset (price × qty in wallet) |

**Combined capital rules (perp + spot):**
- Max 5 simultaneous positions across BOTH markets
- Total deployed capital ≤ $500 (perp margins + spot asset values)
- Per-trade $100: that's $100 perp margin OR $100 spot exposure (NOT both for the same logical trade)

**Closing a spot position:** sell the held base asset back:
```bash
# 1. Check what you hold
trading-cli phemex spot-balance --currency BTC  # for sBTCUSDT
# 2. Sell that amount
trading-cli phemex spot-place --symbol BTCUSDT --side Sell \
  --ord-type Market --base-qty <held>
# 3. Cancel any pending SL/TP orders that no longer apply
trading-cli phemex spot-orders --symbol BTCUSDT      # list active
trading-cli phemex spot-cancel --symbol BTCUSDT --order-id <id>
```

**Important caveat about spot pricing on testnet:** PHEMEX testnet spot and perp prices DIVERGE significantly (observed: perp BTCUSDT ~$77.5k while spot sBTCUSDT ~$70.5k). **Always derive SL/TP from spot's own price**, not from perp's `trade.stopLoss` block — use `coin-analysis --symbol BTCUSDT.P` for technicals but recompute the entry/SL/TP from current spot last price. A stop above spot's current price will fail with `TE_FALLING_TRIGGER_DIRECTLY`.

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
