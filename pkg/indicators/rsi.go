package indicators

// RSI calculates the Relative Strength Index over a given period.
// Uses standard Wilder smoothing. First `period` values are 0.
// Result is in range 0..100.
func RSI(prices []float64, period int) []float64 {
	if len(prices) == 0 || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(prices))

	if len(prices) <= period {
		return result
	}

	// Calculate gains and losses for the first period
	var sumGains, sumLosses float64
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			sumGains += change
		} else {
			sumLosses += -change
		}
	}

	// Initialize averages using Wilder's smoothing
	avgGain := sumGains / float64(period)
	avgLoss := sumLosses / float64(period)

	// Set RSI at period index
	if avgLoss == 0 {
		if avgGain > 0 {
			result[period] = 100
		} else {
			result[period] = 0
		}
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - (100 / (1 + rs))
	}

	// Calculate RSI for remaining periods using Wilder smoothing
	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		var gain, loss float64
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}

		// Wilder's smoothing
		avgGain = (avgGain*(float64(period)-1) + gain) / float64(period)
		avgLoss = (avgLoss*(float64(period)-1) + loss) / float64(period)

		if avgLoss == 0 {
			if avgGain > 0 {
				result[i] = 100
			} else {
				result[i] = 0
			}
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}

	return result
}
