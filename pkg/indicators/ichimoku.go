package indicators

// IchimokuResult holds the five Ichimoku Cloud lines.
type IchimokuResult struct {
	Tenkan  []float64
	Kijun   []float64
	SenkouA []float64
	SenkouB []float64
	Chikou  []float64
}

// Ichimoku calculates the Ichimoku Cloud indicator.
// Standard periods: Tenkan=9, Kijun=26, SenkouB=52, displacement=26.
// Tenkan-sen    = (Highest High + Lowest Low) / 2 over period 9
// Kijun-sen     = (Highest High + Lowest Low) / 2 over period 26
// Senkou Span A = (Tenkan + Kijun) / 2, shifted 26 periods ahead
// Senkou Span B = (Highest High + Lowest Low) / 2 over period 52, shifted 26 periods ahead
// Chikou Span   = Close price, shifted 26 periods behind
// All slices have same length as input; leading zeros where not enough data.
func Ichimoku(highs, lows, closes []float64) IchimokuResult {
	n := len(highs)
	result := IchimokuResult{
		Tenkan:  make([]float64, n),
		Kijun:   make([]float64, n),
		SenkouA: make([]float64, n),
		SenkouB: make([]float64, n),
		Chikou:  make([]float64, n),
	}

	if n == 0 {
		return result
	}

	const (
		tenkanPeriod  = 9
		kijunPeriod   = 26
		senkouBPeriod = 52
		displacement  = 26
	)

	// Helper: compute (HH + LL) / 2 for a given period at each index.
	hl2 := func(period int) []float64 {
		out := make([]float64, n)
		for i := period - 1; i < n; i++ {
			maxHigh := highs[i-period+1]
			minLow := lows[i-period+1]
			for j := i - period + 2; j <= i; j++ {
				if highs[j] > maxHigh {
					maxHigh = highs[j]
				}
				if lows[j] < minLow {
					minLow = lows[j]
				}
			}
			out[i] = (maxHigh + minLow) / 2.0
		}
		return out
	}

	tenkan := hl2(tenkanPeriod)
	kijun := hl2(kijunPeriod)
	senkouBRaw := hl2(senkouBPeriod)

	copy(result.Tenkan, tenkan)
	copy(result.Kijun, kijun)

	// Senkou Span A: (Tenkan + Kijun) / 2, shifted forward by displacement.
	// Only valid once Kijun has enough data (i >= kijunPeriod-1).
	for i := kijunPeriod - 1; i < n; i++ {
		if i+displacement < n {
			result.SenkouA[i+displacement] = (tenkan[i] + kijun[i]) / 2.0
		}
	}

	// Senkou Span B: HL2 over 52, shifted forward by displacement.
	for i := 0; i < n; i++ {
		if i+displacement < n {
			result.SenkouB[i+displacement] = senkouBRaw[i]
		}
	}

	// Chikou Span: close price shifted backward by displacement.
	for i := 0; i < n; i++ {
		if i-displacement >= 0 {
			result.Chikou[i-displacement] = closes[i]
		}
	}

	return result
}
