package indicators

// SMA calculates the Simple Moving Average over a given period.
// Returns a slice of the same length as prices.
// First (period-1) values are 0.
func SMA(prices []float64, period int) []float64 {
	if len(prices) == 0 || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(prices))

	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += prices[i-period+1+j]
		}
		result[i] = sum / float64(period)
	}

	return result
}
