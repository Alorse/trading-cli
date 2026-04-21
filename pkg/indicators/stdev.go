package indicators

import (
	"math"
)

// StdDev calculates population standard deviation over a rolling window.
// Returns a slice of the same length as values.
// First (period-1) values are 0.
func StdDev(values []float64, period int) []float64 {
	if len(values) == 0 || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(values))

	for i := period - 1; i < len(values); i++ {
		// Calculate mean of window
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += values[i-period+1+j]
		}
		mean := sum / float64(period)

		// Calculate variance
		variance := 0.0
		for j := 0; j < period; j++ {
			diff := values[i-period+1+j] - mean
			variance += diff * diff
		}
		variance /= float64(period)

		// Standard deviation is sqrt of variance
		result[i] = math.Sqrt(variance)
	}

	return result
}
