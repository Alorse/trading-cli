package indicators

// AwesomeOscillator calculates the Awesome Oscillator (AO).
// AO = SMA5(MedianPrice) - SMA34(MedianPrice)
// MedianPrice = (High + Low) / 2
// Returns a slice of the same length as inputs; leading zeros where
// insufficient data. First valid value is at index 33.
func AwesomeOscillator(highs, lows []float64) []float64 {
	if len(highs) == 0 || len(lows) == 0 || len(highs) != len(lows) {
		return []float64{}
	}

	median := make([]float64, len(highs))
	for i := range highs {
		median[i] = (highs[i] + lows[i]) / 2
	}

	sma5 := SMA(median, 5)
	sma34 := SMA(median, 34)

	result := make([]float64, len(highs))
	for i := 33; i < len(highs); i++ {
		result[i] = sma5[i] - sma34[i]
	}

	return result
}
