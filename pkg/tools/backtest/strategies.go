package backtest

import (
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/indicators"
)

// RSIStrategy implements a simple RSI-based strategy.
// Buy when RSI(14) < 40, Sell when RSI(14) > 60.
type RSIStrategy struct{}

func (s *RSIStrategy) Name() string {
	return "rsi"
}

func (s *RSIStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	// Extract closes
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	// Calculate RSI(14)
	rsiValues := indicators.RSI(closes, 14)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 0; i < len(candles); i++ {
		if rsiValues[i] == 0 {
			continue
		}

		if !inPosition && rsiValues[i] < 40 {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && rsiValues[i] > 60 {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// BollingerStrategy implements a Bollinger Bands strategy.
// Buy when close < BB(20,2).Lower, Sell when close > BB(20,2).Middle.
type BollingerStrategy struct{}

func (s *BollingerStrategy) Name() string {
	return "bollinger"
}

func (s *BollingerStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	bb := indicators.BollingerBands(closes, 20, 2.0)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 0; i < len(candles); i++ {
		if bb.Lower[i] == 0 {
			continue
		}

		if !inPosition && closes[i] < bb.Lower[i] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && closes[i] > bb.Middle[i] {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// MACDStrategy implements a MACD crossover strategy.
// Buy when MACD crosses above Signal, Sell when MACD crosses below Signal.
type MACDStrategy struct{}

func (s *MACDStrategy) Name() string {
	return "macd"
}

func (s *MACDStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	macdResult := indicators.MACD(closes, 12, 26, 9)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 1; i < len(candles); i++ {
		// Ensure we have valid MACD and Signal values
		if macdResult.MACD[i] == 0 && macdResult.Signal[i] == 0 {
			continue
		}

		// Check for crossover: MACD crosses above Signal
		if !inPosition && macdResult.MACD[i] > macdResult.Signal[i] && macdResult.MACD[i-1] <= macdResult.Signal[i-1] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && macdResult.MACD[i] < macdResult.Signal[i] && macdResult.MACD[i-1] >= macdResult.Signal[i-1] {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// EMACrossStrategy implements an EMA crossover strategy.
// Buy when EMA(20) crosses above EMA(50), Sell when EMA(20) crosses below EMA(50).
type EMACrossStrategy struct{}

func (s *EMACrossStrategy) Name() string {
	return "ema-cross"
}

func (s *EMACrossStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	ema20 := indicators.EMA(closes, 20)
	ema50 := indicators.EMA(closes, 50)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 1; i < len(candles); i++ {
		if ema20[i] == 0 || ema50[i] == 0 {
			continue
		}

		// EMA(20) crosses above EMA(50)
		if !inPosition && ema20[i] > ema50[i] && ema20[i-1] <= ema50[i-1] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && ema20[i] < ema50[i] && ema20[i-1] >= ema50[i-1] {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// SupertrendStrategy implements a Supertrend strategy.
// Buy when direction flips to +1, Sell when direction flips to -1.
// Uses period=10, multiplier=3.0.
type SupertrendStrategy struct{}

func (s *SupertrendStrategy) Name() string {
	return "supertrend"
}

func (s *SupertrendStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))

	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}

	st := indicators.Supertrend(highs, lows, closes, 10, 3.0)

	signals := make([]int, len(candles))
	var lastDirection int

	for i := 0; i < len(candles); i++ {
		currentDirection := st.Direction[i]

		if currentDirection == 0 {
			continue
		}

		// Buy when direction flips to +1
		if currentDirection == 1 && lastDirection != 1 {
			signals[i] = 1
		} else if currentDirection == -1 && lastDirection != -1 {
			// Sell when direction flips to -1
			signals[i] = -1
		}

		lastDirection = currentDirection
	}

	return signals
}

// DonchianStrategy implements a Donchian breakout strategy.
// Buy when high > previous period upper, Sell when close < current lower.
// Uses period=20.
type DonchianStrategy struct{}

func (s *DonchianStrategy) Name() string {
	return "donchian"
}

func (s *DonchianStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))

	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
	}

	dc := indicators.DonchianChannel(highs, lows, 20)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 1; i < len(candles); i++ {
		if dc.Upper[i] == 0 || dc.Lower[i] == 0 {
			continue
		}

		// Buy when high breaks above previous upper band
		if !inPosition && highs[i] > dc.Upper[i-1] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && candles[i].Close < dc.Lower[i] {
			// Sell when close falls below current lower band
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// --- StochRSI + Supertrend + EMA200 strategy ---
// Long-only. Enters on StochRSI oversold cross-up while above EMA200 and Supertrend bullish.
// Provides ATR-based SL/TP via StrategyWithStops.
type StochRSISupertrendStrategy struct{}

func (s *StochRSISupertrendStrategy) Name() string { return "stochrsi-supertrend" }

func (s *StochRSISupertrendStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	n := len(candles)
	if n == 0 {
		return []int{}
	}
	closes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
		highs[i] = c.High
		lows[i] = c.Low
	}

	ema200 := indicators.EMA(closes, 200)
	st := indicators.Supertrend(highs, lows, closes, 11, 2.0)
	srsi := indicators.StochRSI(closes, 14, 14, 3, 3)

	signals := make([]int, n)
	inPos := false
	for i := 1; i < n; i++ {
		if ema200[i] == 0 || srsi.K[i] == 0 && srsi.D[i] == 0 {
			continue
		}
		if !inPos {
			trendOK := closes[i] > ema200[i]
			supertrendOK := st.Direction[i] == 1
			crossUp := srsi.K[i-1] <= srsi.D[i-1] && srsi.K[i] > srsi.D[i]
			oversold := srsi.K[i] < 30
			if trendOK && supertrendOK && crossUp && oversold {
				signals[i] = 1
				inPos = true
			}
		} else {
			// Exit when Supertrend flips bearish (dynamic trailing stop)
			// or when StochRSI hits extreme overbought zone.
			supertrendFlip := st.Direction[i] == -1
			extremeOB := srsi.K[i] > 80 && srsi.K[i-1] >= srsi.D[i-1] && srsi.K[i] < srsi.D[i]
			if supertrendFlip || extremeOB {
				signals[i] = -1
				inPos = false
			}
		}
	}
	return signals
}

func (s *StochRSISupertrendStrategy) EntryStops(candles []client.YahooOHLCV, entryIdx int) (float64, float64) {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	atr := indicators.ATR(highs, lows, closes, 14)
	if entryIdx >= len(atr) || atr[entryIdx] == 0 {
		return 0, 0
	}
	entry := candles[entryIdx].Close
	// Wider hard SL (catastrophic), wider TP (let runners). Supertrend flip handles primary exit.
	return entry - 3.0*atr[entryIdx], entry + 6.0*atr[entryIdx]
}

// --- DMI + ADX>23 + EMA55 strategy ---
type DMIADXTrendStrategy struct{}

func (s *DMIADXTrendStrategy) Name() string { return "dmi-adx-trend" }

func (s *DMIADXTrendStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	n := len(candles)
	if n == 0 {
		return []int{}
	}
	closes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	opens := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
		highs[i] = c.High
		lows[i] = c.Low
		opens[i] = c.Open
	}

	ema55 := indicators.EMA(closes, 55)
	dmi := indicators.ADX(highs, lows, closes, 14)

	signals := make([]int, n)
	inPos := false
	for i := 3; i < n; i++ {
		if ema55[i] == 0 || dmi.ADX[i] == 0 {
			continue
		}
		if !inPos {
			diDom := dmi.PlusDI[i] > dmi.MinusDI[i] &&
				dmi.PlusDI[i-1] > dmi.MinusDI[i-1] &&
				dmi.PlusDI[i-2] > dmi.MinusDI[i-2]
			adxOK := dmi.ADX[i] > 23
			trendOK := closes[i] > ema55[i]
			bullCandle := closes[i] > opens[i]
			if diDom && adxOK && trendOK && bullCandle {
				signals[i] = 1
				inPos = true
			}
		} else {
			if dmi.PlusDI[i] < dmi.MinusDI[i] {
				signals[i] = -1
				inPos = false
			}
		}
	}
	return signals
}

func (s *DMIADXTrendStrategy) EntryStops(candles []client.YahooOHLCV, entryIdx int) (float64, float64) {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	atr := indicators.ATR(highs, lows, closes, 14)
	if entryIdx >= len(atr) || atr[entryIdx] == 0 {
		return 0, 0
	}
	entry := candles[entryIdx].Close
	return entry - 1.5*atr[entryIdx], entry + 3.0*atr[entryIdx]
}

// --- TTM Squeeze + linreg momentum strategy ---
type TTMSqueezeStrategy struct{}

func (s *TTMSqueezeStrategy) Name() string { return "ttm-squeeze" }

// linreg computes a simple linear regression value at index i over the last `length` values.
// Returns the regression line's last point (equivalent to TradingView's linreg(source, length, 0)).
func linregAt(values []float64, i, length int) float64 {
	if i < length-1 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(length)
	for k := 0; k < length; k++ {
		x := float64(k)
		y := values[i-length+1+k]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	slope := (n*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / n
	return intercept + slope*(n-1) // value at last x = length-1
}

func (s *TTMSqueezeStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	n := len(candles)
	if n == 0 {
		return []int{}
	}
	closes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
		highs[i] = c.High
		lows[i] = c.Low
	}

	const length = 20
	bb := indicators.BollingerBands(closes, length, 2.0)
	kc := indicators.KeltnerChannel(highs, lows, closes, length, 1.5)

	// TTM momentum oscillator: linreg of (close - average(highest, lowest, sma) of length) over length.
	// Source per Carter: val = close - avg( avg(highest(high,length), lowest(low,length)), sma(close,length) )
	src := make([]float64, n)
	sma := indicators.SMA(closes, length)
	for i := length - 1; i < n; i++ {
		hh := highs[i]
		ll := lows[i]
		for k := i - length + 1; k <= i; k++ {
			if highs[k] > hh {
				hh = highs[k]
			}
			if lows[k] < ll {
				ll = lows[k]
			}
		}
		mid := (hh + ll) / 2.0
		avg := (mid + sma[i]) / 2.0
		src[i] = closes[i] - avg
	}
	osc := make([]float64, n)
	for i := length - 1; i < n; i++ {
		osc[i] = linregAt(src, i, length)
	}

	signals := make([]int, n)
	inPos := false
	for i := length + 1; i < n; i++ {
		if bb.Middle[i] == 0 || kc.Middle[i] == 0 {
			continue
		}
		squeezePrev := bb.Upper[i-1] < kc.Upper[i-1] && bb.Lower[i-1] > kc.Lower[i-1]
		squeezeNow := bb.Upper[i] < kc.Upper[i] && bb.Lower[i] > kc.Lower[i]
		release := squeezePrev && !squeezeNow

		if !inPos {
			if release && osc[i] > 0 && osc[i] > osc[i-1] {
				signals[i] = 1
				inPos = true
			}
		} else {
			// Exit when momentum crosses below zero or starts falling sharply
			if osc[i] < 0 || (osc[i] < osc[i-1] && osc[i-1] < osc[i-2]) {
				signals[i] = -1
				inPos = false
			}
		}
	}
	return signals
}

func (s *TTMSqueezeStrategy) EntryStops(candles []client.YahooOHLCV, entryIdx int) (float64, float64) {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	atr := indicators.ATR(highs, lows, closes, 14)
	if entryIdx >= len(atr) || atr[entryIdx] == 0 {
		return 0, 0
	}
	entry := candles[entryIdx].Close
	return entry - 2.0*atr[entryIdx], entry + 4.0*atr[entryIdx]
}

// --- Williams Vix Fix strategy ---
// Buy on capitulation: WVF crosses above its own BB upper or 85th percentile of prior 50 bars,
// while price is below BB(close, 20, 2.0) lower (deeply oversold).
// Exit on Supertrend flip down or BB upper touch.
type WilliamsVixFixStrategy struct{}

func (s *WilliamsVixFixStrategy) Name() string { return "wvf" }

func (s *WilliamsVixFixStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	n := len(candles)
	if n == 0 {
		return []int{}
	}
	closes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
		highs[i] = c.High
		lows[i] = c.Low
	}

	const pd = 22
	const lb = 50
	const bbl = 20
	const mult = 2.0
	const ph = 0.85

	// WVF series
	wvf := make([]float64, n)
	for i := pd; i < n; i++ {
		hc := closes[i-pd+1]
		for k := i - pd + 1; k <= i; k++ {
			if closes[k] > hc {
				hc = closes[k]
			}
		}
		if hc > 0 {
			wvf[i] = ((hc - lows[i]) / hc) * 100.0
		}
	}

	// SMA + StDev of WVF
	midLine := indicators.SMA(wvf, bbl)
	sDev := indicators.StdDev(wvf, bbl)
	upperBand := make([]float64, n)
	for i := 0; i < n; i++ {
		upperBand[i] = midLine[i] + mult*sDev[i]
	}

	// Range high (percentile-ish proxy: 85% of highest WVF over lb bars)
	rangeHigh := make([]float64, n)
	for i := lb; i < n; i++ {
		hh := wvf[i-lb+1]
		for k := i - lb + 1; k <= i; k++ {
			if wvf[k] > hh {
				hh = wvf[k]
			}
		}
		rangeHigh[i] = hh * ph
	}

	bb := indicators.BollingerBands(closes, bbl, mult)
	st := indicators.Supertrend(highs, lows, closes, 10, 3.0)

	signals := make([]int, n)
	inPos := false
	for i := lb + 1; i < n; i++ {
		if midLine[i] == 0 || bb.Lower[i] == 0 {
			continue
		}
		if !inPos {
			cap := (wvf[i] >= upperBand[i] || wvf[i] >= rangeHigh[i]) &&
				(wvf[i-1] >= upperBand[i-1] || wvf[i-1] >= rangeHigh[i-1]) &&
				closes[i] < bb.Lower[i]
			if cap {
				signals[i] = 1
				inPos = true
			}
		} else {
			if st.Direction[i] == -1 || closes[i] > bb.Upper[i] {
				signals[i] = -1
				inPos = false
			}
		}
	}
	return signals
}

func (s *WilliamsVixFixStrategy) EntryStops(candles []client.YahooOHLCV, entryIdx int) (float64, float64) {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	atr := indicators.ATR(highs, lows, closes, 14)
	if entryIdx >= len(atr) || atr[entryIdx] == 0 {
		return 0, 0
	}
	entry := candles[entryIdx].Close
	// Capitulation entries: wide SL (price is already extended low), let target run.
	return entry - 2.5*atr[entryIdx], entry + 5.0*atr[entryIdx]
}

// --- Range Filter strategy ---
// Computes an EMA-smoothed range filter and trades only with its direction.
// Long when filter is rising for N bars + close above filter.
type RangeFilterStrategy struct{}

func (s *RangeFilterStrategy) Name() string { return "range-filter" }

func (s *RangeFilterStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	n := len(candles)
	if n == 0 {
		return []int{}
	}
	closes := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
		highs[i] = c.High
		lows[i] = c.Low
	}

	const per = 100
	const mult = 3.0

	// Average range = EMA(|close - prev_close|, per)
	absChange := make([]float64, n)
	for i := 1; i < n; i++ {
		diff := closes[i] - closes[i-1]
		if diff < 0 {
			diff = -diff
		}
		absChange[i] = diff
	}
	avrng := indicators.EMA(absChange, per)
	// Smoothed range = EMA(avrng, per*2 - 1) * mult
	smrng := indicators.EMA(avrng, per*2-1)
	for i := range smrng {
		smrng[i] *= mult
	}

	// Filter: trailing range filter
	filt := make([]float64, n)
	filt[0] = closes[0]
	for i := 1; i < n; i++ {
		prev := filt[i-1]
		r := smrng[i]
		if r == 0 {
			filt[i] = prev
			continue
		}
		if closes[i] > prev {
			if closes[i]-r < prev {
				filt[i] = prev
			} else {
				filt[i] = closes[i] - r
			}
		} else {
			if closes[i]+r > prev {
				filt[i] = prev
			} else {
				filt[i] = closes[i] + r
			}
		}
	}

	// Direction counters
	upward := make([]int, n)
	for i := 1; i < n; i++ {
		if filt[i] > filt[i-1] {
			upward[i] = upward[i-1] + 1
		} else if filt[i] < filt[i-1] {
			upward[i] = 0
		} else {
			upward[i] = upward[i-1]
		}
	}

	signals := make([]int, n)
	inPos := false
	for i := per + 1; i < n; i++ {
		if filt[i] == 0 {
			continue
		}
		if !inPos {
			// Entry: filter has been rising 2+ bars, close above filter, and a fresh leg.
			if upward[i] >= 2 && upward[i-1] >= 1 && closes[i] > filt[i] && closes[i] > closes[i-1] {
				signals[i] = 1
				inPos = true
			}
		} else {
			// Exit: filter starts falling or close drops below filter.
			if upward[i] == 0 || closes[i] < filt[i] {
				signals[i] = -1
				inPos = false
			}
		}
	}
	return signals
}

func (s *RangeFilterStrategy) EntryStops(candles []client.YahooOHLCV, entryIdx int) (float64, float64) {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	atr := indicators.ATR(highs, lows, closes, 14)
	if entryIdx >= len(atr) || atr[entryIdx] == 0 {
		return 0, 0
	}
	entry := candles[entryIdx].Close
	return entry - 2.5*atr[entryIdx], entry + 5.0*atr[entryIdx]
}

// GetStrategy returns the appropriate strategy for the given name.
func GetStrategy(name string) Strategy {
	switch name {
	case "rsi":
		return &RSIStrategy{}
	case "bollinger":
		return &BollingerStrategy{}
	case "macd":
		return &MACDStrategy{}
	case "ema-cross":
		return &EMACrossStrategy{}
	case "supertrend":
		return &SupertrendStrategy{}
	case "donchian":
		return &DonchianStrategy{}
	case "stochrsi-supertrend":
		return &StochRSISupertrendStrategy{}
	case "dmi-adx-trend":
		return &DMIADXTrendStrategy{}
	case "ttm-squeeze":
		return &TTMSqueezeStrategy{}
	case "wvf":
		return &WilliamsVixFixStrategy{}
	case "range-filter":
		return &RangeFilterStrategy{}
	default:
		return nil
	}
}
