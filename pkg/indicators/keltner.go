package indicators

// KeltnerResult holds Keltner Channel bands.
type KeltnerResult struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
	Width  []float64
}

// KeltnerChannel computes Keltner Channels.
// Middle = EMA(closes, period). Upper = Middle + mult*ATR(period). Lower = Middle - mult*ATR(period).
// Width = (Upper - Lower) / Middle.
func KeltnerChannel(highs, lows, closes []float64, period int, mult float64) KeltnerResult {
	n := len(closes)
	result := KeltnerResult{
		Upper:  make([]float64, n),
		Middle: make([]float64, n),
		Lower:  make([]float64, n),
		Width:  make([]float64, n),
	}
	if n == 0 || period <= 0 {
		return result
	}

	middle := EMA(closes, period)
	atr := ATR(highs, lows, closes, period)

	for i := 0; i < n; i++ {
		result.Middle[i] = middle[i]
		if middle[i] == 0 || atr[i] == 0 {
			continue
		}
		result.Upper[i] = middle[i] + mult*atr[i]
		result.Lower[i] = middle[i] - mult*atr[i]
		if middle[i] != 0 {
			result.Width[i] = (result.Upper[i] - result.Lower[i]) / middle[i]
		}
	}

	return result
}
