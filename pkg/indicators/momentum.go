package indicators

// Momentum calculates the Momentum indicator over a given period.
// Formula: Close[t] - Close[t-period]
// First `period` values are 0 (insufficient data).
// Returns a slice of the same length as closes.
func Momentum(closes []float64, period int) []float64 {
	if len(closes) == 0 || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(closes))

	for i := period; i < len(closes); i++ {
		result[i] = closes[i] - closes[i-period]
	}

	return result
}
