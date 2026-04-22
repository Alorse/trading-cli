package indicators

// CCI calculates the Commodity Channel Index.
// Typical Price = (High + Low + Close) / 3
// CCI = (TP - SMA(TP, period)) / (0.015 * Mean Deviation)
// Mean Deviation = average of absolute differences between each TP and SMA(TP)
// over the period.
// Returns a slice of the same length as inputs; leading zeros where
// insufficient data.
func CCI(highs, lows, closes []float64, period int) []float64 {
	if len(highs) == 0 || len(lows) == 0 || len(closes) == 0 || period <= 0 {
		return []float64{}
	}

	n := len(highs)
	result := make([]float64, n)

	if n <= period {
		return result
	}

	// Calculate Typical Prices
	tp := make([]float64, n)
	for i := 0; i < n; i++ {
		tp[i] = (highs[i] + lows[i] + closes[i]) / 3
	}

	// Calculate SMA of Typical Prices
	tpSMA := SMA(tp, period)

	// Calculate CCI for each valid window
	for i := period - 1; i < n; i++ {
		// Mean Deviation: average of absolute differences
		meanDev := 0.0
		for j := 0; j < period; j++ {
			idx := i - period + 1 + j
			meanDev += abs(tp[idx] - tpSMA[i])
		}
		meanDev /= float64(period)

		if meanDev == 0 {
			result[i] = 0
			continue
		}

		result[i] = (tp[i] - tpSMA[i]) / (0.015 * meanDev)
	}

	return result
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
