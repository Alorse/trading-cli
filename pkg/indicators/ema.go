package indicators

// EMA calculates the Exponential Moving Average over a given period.
// Seeds with SMA of first `period` values, then applies multiplier = 2/(period+1).
// Returns a slice of the same length as prices.
// First (period-1) values are 0.
func EMA(prices []float64, period int) []float64 {
	if len(prices) == 0 || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	// Seed with SMA of first period values
	if len(prices) >= period {
		sum := 0.0
		for i := 0; i < period; i++ {
			sum += prices[i]
		}
		result[period-1] = sum / float64(period)
	}

	// Apply EMA formula
	for i := period; i < len(prices); i++ {
		result[i] = prices[i]*multiplier + result[i-1]*(1-multiplier)
	}

	return result
}
