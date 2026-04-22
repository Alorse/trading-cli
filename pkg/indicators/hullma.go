package indicators

import "math"

// HullMA calculates the Hull Moving Average over a given period.
// Returns a slice of the same length as closes.
// Leading values are 0 where insufficient data.
func HullMA(closes []float64, period int) []float64 {
	if len(closes) == 0 || period <= 0 {
		return []float64{}
	}

	halfPeriod := int(math.Round(float64(period) / 2))
	sqrtPeriod := int(math.Round(math.Sqrt(float64(period))))

	wma1 := wma(closes, halfPeriod)
	wma2 := wma(closes, period)

	raw := make([]float64, len(closes))
	for i := 0; i < len(closes); i++ {
		raw[i] = 2*wma1[i] - wma2[i]
	}

	hma := wma(raw, sqrtPeriod)

	firstValid := period + sqrtPeriod - 2
	for i := 0; i < firstValid && i < len(hma); i++ {
		hma[i] = 0
	}

	return hma
}

// wma calculates the Weighted Moving Average over a given period.
// The most recent value gets weight n, the oldest gets weight 1.
// Returns a slice of the same length as closes.
func wma(closes []float64, period int) []float64 {
	if len(closes) == 0 || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(closes))
	denom := float64(period * (period + 1) / 2)

	for i := period - 1; i < len(closes); i++ {
		var sum float64
		for j := 0; j < period; j++ {
			weight := float64(period - j)
			sum += closes[i-j] * weight
		}
		result[i] = sum / denom
	}

	return result
}
