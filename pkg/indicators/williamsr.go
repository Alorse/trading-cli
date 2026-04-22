package indicators

// WilliamsR calculates the Williams %R oscillator over a given period.
// %R = (HighestHigh - Close) / (HighestHigh - LowestLow) * -100
// Standard period is 14. Values range from -100 (oversold) to 0 (overbought).
// If the range is 0 (flat market), returns -50 as neutral.
// The result slice has the same length as the input slices; leading values
// are 0 where there is insufficient data.
func WilliamsR(highs, lows, closes []float64, period int) []float64 {
	if len(highs) == 0 || len(lows) == 0 || len(closes) == 0 || period <= 0 {
		return []float64{}
	}

	n := len(closes)
	result := make([]float64, n)

	if period > n {
		return result
	}

	for i := period - 1; i < n; i++ {
		lowestLow := lows[i]
		highestHigh := highs[i]

		for j := i - period + 1; j <= i; j++ {
			if lows[j] < lowestLow {
				lowestLow = lows[j]
			}
			if highs[j] > highestHigh {
				highestHigh = highs[j]
			}
		}

		rng := highestHigh - lowestLow
		if rng > 0 {
			result[i] = (highestHigh - closes[i]) / rng * -100
		} else {
			result[i] = -50
		}
	}

	return result
}
