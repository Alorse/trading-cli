package indicators

// StochasticResult holds the %K and %D values from the Stochastic Oscillator.
type StochasticResult struct {
	K []float64
	D []float64
}

// Stochastic calculates the Stochastic Oscillator.
// %K measures the current close relative to the high/low range over the period.
// %D is a 3-period simple moving average of %K.
// All slices are the same length as the input slices.
// Leading values are 0 where there is insufficient data.
func Stochastic(highs, lows, closes []float64, period int) StochasticResult {
	if len(highs) == 0 || len(lows) == 0 || len(closes) == 0 || period <= 0 {
		return StochasticResult{K: []float64{}, D: []float64{}}
	}

	n := len(closes)
	k := make([]float64, n)
	d := make([]float64, n)

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
			k[i] = (closes[i] - lowestLow) / rng * 100
		} else {
			k[i] = 50
		}
	}

	// %D is SMA(%K, 3) computed only where %K is valid.
	// First valid %D is at index period + 1.
	for i := period + 1; i < n; i++ {
		d[i] = (k[i-2] + k[i-1] + k[i]) / 3
	}

	return StochasticResult{K: k, D: d}
}
