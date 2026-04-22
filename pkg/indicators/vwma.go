package indicators

// VWMA calculates the Volume Weighted Moving Average over a given period.
// Formula: sum(close[i] * volume[i]) / sum(volume[i]) for the window.
// Returns a slice of the same length as closes and volumes.
// First (period-1) values are 0.
// If total volume in a window is 0, the result for that bar is 0.
func VWMA(closes []float64, volumes []float64, period int) []float64 {
	if len(closes) == 0 || len(volumes) == 0 || len(closes) != len(volumes) || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(closes))

	for i := period - 1; i < len(closes); i++ {
		var weightedSum, volSum float64
		for j := 0; j < period; j++ {
			idx := i - period + 1 + j
			weightedSum += closes[idx] * volumes[idx]
			volSum += volumes[idx]
		}
		if volSum == 0 {
			result[i] = 0
		} else {
			result[i] = weightedSum / volSum
		}
	}

	return result
}

