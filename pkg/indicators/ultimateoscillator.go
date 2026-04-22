package indicators

// UltimateOscillator calculates Larry Williams' Ultimate Oscillator.
// Uses 3 timeframes (7, 14, 28) to reduce false signals.
// Result is in range 0..100. First valid value is at index 27.
// Returns empty slice if fewer than 2 data points.
func UltimateOscillator(highs, lows, closes []float64) []float64 {
	if len(highs) < 2 || len(lows) < 2 || len(closes) < 2 {
		return []float64{}
	}

	n := len(closes)
	result := make([]float64, n)

	// Precompute Buying Pressure and True Range
	bp := make([]float64, n)
	tr := make([]float64, n)

	bp[0] = closes[0] - lows[0]
	tr[0] = highs[0] - lows[0]

	for i := 1; i < n; i++ {
		prevClose := closes[i-1]
		bp[i] = closes[i] - min(lows[i], prevClose)
		tr[i] = max(highs[i], prevClose) - min(lows[i], prevClose)
	}

	for i := 27; i < n; i++ {
		var sumBP7, sumTR7 float64
		for j := i - 6; j <= i; j++ {
			sumBP7 += bp[j]
			sumTR7 += tr[j]
		}

		var sumBP14, sumTR14 float64
		for j := i - 13; j <= i; j++ {
			sumBP14 += bp[j]
			sumTR14 += tr[j]
		}

		var sumBP28, sumTR28 float64
		for j := i - 27; j <= i; j++ {
			sumBP28 += bp[j]
			sumTR28 += tr[j]
		}

		var avg7, avg14, avg28 float64
		if sumTR7 > 0 {
			avg7 = sumBP7 / sumTR7
		}
		if sumTR14 > 0 {
			avg14 = sumBP14 / sumTR14
		}
		if sumTR28 > 0 {
			avg28 = sumBP28 / sumTR28
		}

		result[i] = 100 * ((4 * avg7) + (2 * avg14) + avg28) / 7
	}

	return result
}
