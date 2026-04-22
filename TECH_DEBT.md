# Technical Debt & Implementation Tracker

> Living document. Update after every batch of work.
> Last updated: 2026-04-22 (post-audit fixes applied)

---

## P0 — Critical (Blocking full spec compliance)

### 1. Missing Pure-Go Technical Indicators

The REQUIREMENT.md mandates *"All indicator calculations (RSI, MACD, Bollinger Bands, EMA, SMA, ATR, Supertrend, Donchian Channel) must be implemented in pure Go using only the standard library."*

Currently implemented (8/23+):
- [x] RSI
- [x] MACD
- [x] Bollinger Bands
- [x] EMA
- [x] SMA
- [x] ATR
- [x] Supertrend
- [x] Donchian Channel

**Missing (need pure-Go implementation in `pkg/indicators/`):**

| # | Indicator | Status | Batch |
|---|-----------|--------|-------|
| 1 | Stochastic (%K, %D) | **DONE** | Batch 1 |
| 2 | ADX (+DI, -DI) | **DONE** | Batch 1 |
| 3 | OBV (On-Balance Volume) | **DONE** | Batch 1 |
| 4 | CCI (Commodity Channel Index) | **DONE** | Batch 2 |
| 5 | Williams %R | **DONE** | Batch 2 |
| 6 | Awesome Oscillator | **DONE** | Batch 2 |
| 7 | Momentum | **DONE** | Batch 3 |
| 8 | Parabolic SAR | **DONE** | Batch 3 |
| 9 | Ichimoku Cloud | **DONE** | Batch 3 |
| 10 | Hull MA | **DONE** | Batch 4 |
| 11 | Stochastic RSI | **DONE** | Batch 4 |
| 12 | Ultimate Oscillator | **DONE** | Batch 4 |
| 13 | VWMA (Volume Weighted Moving Average) | **DONE** | Batch 5 |
| 14 | VWAP (Volume Weighted Average Price) | **DONE** | Batch 5 |

**Acceptance criteria per indicator:**
- `pkg/indicators/<name>.go` with pure-Go calculation
- `pkg/indicators/<name>_test.go` with table-driven tests (happy path + edge cases)
- Follow existing patterns (see `rsi.go`, `macd.go`, `bollinger.go`)
- No external dependencies
- Run `go test ./pkg/indicators/ -cover` and ensure tests pass

### 2. Backtest Transaction Cost Bug

**Location:** `pkg/tools/backtest/engine.go:74-104`

**Problem:** Double division by 100.
```go
costPerTrade := (commissionPct + slippagePct) / 100 * 2  // e.g. 0.15 -> 0.003
transactionCost := costPerTrade / 100                     // 0.003 -> 0.00003
```
Expected: `0.0015` (0.15%). Actual: `0.00003` (0.003%). Costs are 100x smaller.

**Fix:** Remove the second `/ 100` in BOTH trade exit paths (sell signal + position close at last candle).

**Status:** **DONE** (Batch 5 + post-audit fix)

---

## Post-Audit Critical Fixes (2026-04-22)

### A1. MaxDrawdown always reports 0 — FIXED

**File:** `pkg/tools/backtest/engine.go`
**Problem:** `equityCurve` only populated when `includeEquityCurve=true`, but `calculateMaxDrawdown()` always ran against it. With flag=false (default in compare-strategies), curve was `[10000]` → MaxDrawdown=0 always.
**Fix:** Always append to `equityCurve` on trade exits. Flag now only controls JSON serialization.
**Status:** **DONE** — MaxDrawdown now returns -28.69 for BTC-USD 1y RSI backtest.

### A2. Supertrend produces 0 trades — FIXED

**File:** `pkg/indicators/supertrend.go`
**Problem:** Band continuity logic was too aggressive — upper band never dropped, lower band never rose. With multiplier=3.0, bands became permanently detached from price, causing zero direction flips.
**Fix:** Applied standard Supertrend band continuity: bands only stay at previous level if the previous close respected that level. Otherwise, bands adjust freely.
**Status:** **DONE** — Supertrend now produces 7 trades for BTC-USD 1y (was 0). Donchian unchanged at 1 trade (matches reference).

---

## Divergence Analysis (2026-04-22)

| # | Divergence | Verdict | Action |
|---|-----------|---------|--------|
| 3 | Change% formula: Go uses `(close-prevClose)/prevClose`, Python uses `(close-open)/open` | **Go is correct** — industry standard is prevClose-based. Python computes "change from open," a niche metric. | No action needed for Go. Python reference has a misnamed function. |
| 4 | Volume ratio: Go uses TV `relative_volume_10d_calc`, Python uses `volume/SMA20` | **Both valid** — no single industry standard. TV uses 10d, breakout scanners often use 20d, StockCharts uses 50d. | Consider self-computing ratio with configurable lookback (default 20). Also fix naming: `Avg20` fields actually use 10d data. |
| 5 | EMA cross periods: Go uses 20/50, Python uses 9/21 | **Go is spec-compliant** — REQUIREMENT.md specifies 20/50. Both pairs are widely used; 20/50 is more conservative and appropriate for a general CLI. | Fix docs: `docs/TOOLS.md` incorrectly documents "EMA 9/21" while code implements 20/50. Consider adding `--ema-fast` / `--ema-slow` flags. |

---

## P1 — High (Spec mismatch, user-facing)

### 3. `coin-analysis` Output Gaps

The REQUIREMENT specifies 23+ indicator groups. Current implementation extracts ~12.

**Missing fields in `CoinAnalysisOutput`:**
- OBV
- CCI
- Williams %R
- Awesome Oscillator
- Momentum
- Parabolic SAR
- Ichimoku
- Hull MA
- Stochastic RSI
- Ultimate Oscillator
- VWMA
- VWAP

**Also missing for stock exchanges:**
- Stock score (100-point composite)
- Trade setup (entry, stop-loss, targets, R:R)
- Trade quality (100-point assessment)

**Dependencies:** Blocked until all P0 indicators are implemented.

**Status:** **DONE** (Batch 6 — wired CCI, Williams %R, AO, Momentum, Parabolic SAR, Ichimoku, Hull MA, Stochastic RSI, Ultimate Oscillator, VWAP, VWMA into output. OBV remains unavailable from TV scanner. Post-audit fix: Avg20 now uses `average_volume_10d_calc` from TV instead of hardcoded 0.)

### 4. `multi-timeframe-analysis` — Oversimplified Logic

| Timeframe | Spec Requirement | Current Implementation |
|-----------|-----------------|----------------------|
| 1W | EMA200/100 trend, MACD momentum, RSI > 50 | `close > EMA200 && RSI > 50` only |
| 1D | EMA50/200 golden/death cross, RSI 40-60 zone, volume ratio, MACD | `close > EMA50 && close > EMA200` only |
| 4h | EMA20 > EMA50 alignment, MACD crossover | `EMA20 > EMA50 && change > 0` only |
| 1h | EMA20 dynamic S/R, volume spikes, VWAP | `close > VWAP && change > 0` only |
| 15m | EMA9/20 fast crossover, VWAP institutional | `EMA9 > EMA20 && change > 0` only |

**Status:** **DONE** (Batch 6 — expanded scoring logic; post-audit fix: now fetches real per-timeframe data using TV column suffixes |1W, |240, |60, |15)

### 5. `rating-filter` — Wrong Default Timeframe

| Flag | Spec | Actual |
|------|------|--------|
| `--timeframe` | `5m` | `15m` |

**Status:** PENDING (one-line fix)

---

## P2 — Medium (Polish & Infrastructure)

### 6. `combined-analysis` — Technical Logic Mismatch

Spec says: *"Technical bullish if change > 0"*
Actual: Uses `MarketStructure.Trend == "bullish"` (requires `close > EMA50 && EMA50 > EMA200`).

**Status:** PENDING

### 7. `volume-breakout-scanner` — Missing Batch Logic

Spec says: *"Fetch symbols in batches of 100 (up to 500)"*
Actual: Loads 500 symbols and sends in a single request.

**Status:** PENDING

### 8. CI/CD — GitHub Actions Missing

Spec says: *"CI should run tests on every commit (GitHub Actions recommended)"*
Actual: No `.github/workflows/` directory.

**Status:** PENDING

---

## Implementation Batches

| Batch | Scope | Indicators/Features |
|-------|-------|---------------------|
| 1 | Pure-Go indicators | Stochastic, ADX, OBV |
| 2 | Pure-Go indicators | CCI, Williams %R, Awesome Oscillator |
| 3 | Pure-Go indicators | Momentum, Parabolic SAR, Ichimoku |
| 4 | Pure-Go indicators | Hull MA, Stochastic RSI, Ultimate Oscillator |
| 5 | Pure-Go indicators + bugfix | VWMA, VWAP, Fix backtest transaction cost |
| 6 | Integration | Wire all new indicators into `coin-analysis`, update `multi-timeframe-analysis` logic |
| 7 | Polish | Fix defaults, `combined-analysis` logic, `volume-breakout-scanner` batches, CI/CD |

---

## Reference Patterns

### Indicator Implementation Pattern
```
pkg/indicators/
  <name>.go       — Pure Go calculation function(s)
  <name>_test.go  — Table-driven tests with known inputs/outputs
```

See existing files for style:
- `rsi.go` + `rsi_test.go` — Single-value oscillator with period parameter
- `macd.go` + `macd_test.go` — Multi-output struct (MACD, Signal, Histogram)
- `bollinger.go` + `bollinger_test.go` — Multi-output with upper/middle/lower bands
- `supertrend.go` + `supertrend_test.go` — Complex indicator using high/low/close

### Test Pattern
```go
func TestIndicatorName(t *testing.T) {
    tests := []struct {
        name     string
        input    []float64
        period   int
        expected float64 // or struct
    }{
        {"basic case", []float64{...}, 14, 65.5},
        {"edge empty", []float64{}, 14, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := IndicatorName(tt.input, tt.period)
            if result != tt.expected {
                t.Errorf(...)
            }
        })
    }
}
```

### Commit Pattern
```
Add <indicator-name> pure-Go indicator with tests

- Implements <formula/description>
- Covers happy path and edge cases
- Verified with go test ./pkg/indicators/

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
```
