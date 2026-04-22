# Technical Debt & Implementation Tracker

> Living document. Update after every batch of work.
> Last updated: 2026-04-22

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

**Fix:** Remove the second `/ 100`. Apply per-leg cost correctly.

**Status:** **DONE** (Batch 5)

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

**Status:** **DONE** (Batch 6 — wired CCI, Williams %R, AO, Momentum, Parabolic SAR, Ichimoku, Hull MA, Stochastic RSI, Ultimate Oscillator, VWAP, VWMA into output. OBV remains unavailable from TV scanner.)

### 4. `multi-timeframe-analysis` — Oversimplified Logic

| Timeframe | Spec Requirement | Current Implementation |
|-----------|-----------------|----------------------|
| 1W | EMA200/100 trend, MACD momentum, RSI > 50 | `close > EMA200 && RSI > 50` only |
| 1D | EMA50/200 golden/death cross, RSI 40-60 zone, volume ratio, MACD | `close > EMA50 && close > EMA200` only |
| 4h | EMA20 > EMA50 alignment, MACD crossover | `EMA20 > EMA50 && change > 0` only |
| 1h | EMA20 dynamic S/R, volume spikes, VWAP | `close > VWAP && change > 0` only |
| 15m | EMA9/20 fast crossover, VWAP institutional | `EMA9 > EMA20 && change > 0` only |

**Status:** **DONE** (Batch 6 — expanded to full scoring logic per spec)

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
