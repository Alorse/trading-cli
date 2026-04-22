package indicators

// StochRSIResult holds the %K and %D values from the Stochastic RSI.
type StochRSIResult struct {
	K []float64 // %K (smoothed StochRSI)
	D []float64 // %D (SMA of %K)
}

// StochRSI calculates the Stochastic RSI.
// It applies the Stochastic oscillator formula to RSI values.
// rsiPeriod is the period for RSI calculation (typically 14).
// stochPeriod is the lookback period for min/max RSI (typically 14).
// smoothK is the SMA period for smoothing %K (typically 3).
// smoothD is the SMA period for smoothing %D (typically 3).
// Leading values are 0 where there is insufficient data.
// If all RSI values in the stochPeriod window are the same, returns 50 for that %K value.
func StochRSI(closes []float64, rsiPeriod, stochPeriod, smoothK, smoothD int) StochRSIResult {
	if len(closes) == 0 || rsiPeriod <= 0 || stochPeriod <= 0 || smoothK <= 0 || smoothD <= 0 {
		return StochRSIResult{K: []float64{}, D: []float64{}}
	}

	rsiValues := RSI(closes, rsiPeriod)
	n := len(rsiValues)

	// Compute raw StochRSI
	raw := make([]float64, n)
	firstRawIdx := rsiPeriod + stochPeriod - 1

	for i := firstRawIdx; i < n; i++ {
		minRSI := rsiValues[i]
		maxRSI := rsiValues[i]

		for j := i - stochPeriod + 1; j <= i; j++ {
			if rsiValues[j] < minRSI {
				minRSI = rsiValues[j]
			}
			if rsiValues[j] > maxRSI {
				maxRSI = rsiValues[j]
			}
		}

		rng := maxRSI - minRSI
		if rng > 0 {
			raw[i] = (rsiValues[i] - minRSI) / rng * 100
		} else {
			raw[i] = 50
		}
	}

	// Smooth raw StochRSI with SMA(smoothK) to get %K
	k := make([]float64, n)
	firstKIdx := firstRawIdx + smoothK - 1
	for i := firstKIdx; i < n; i++ {
		sum := 0.0
		for j := 0; j < smoothK; j++ {
			sum += raw[i-smoothK+1+j]
		}
		k[i] = sum / float64(smoothK)
	}

	// Smooth %K with SMA(smoothD) to get %D
	d := make([]float64, n)
	firstDIdx := firstKIdx + smoothD - 1
	for i := firstDIdx; i < n; i++ {
		sum := 0.0
		for j := 0; j < smoothD; j++ {
			sum += k[i-smoothD+1+j]
		}
		d[i] = sum / float64(smoothD)
	}

	return StochRSIResult{K: k, D: d}
}

// StochRSIDefault calculates Stochastic RSI with default parameters:
// rsiPeriod=14, stochPeriod=14, smoothK=3, smoothD=3.
func StochRSIDefault(closes []float64) StochRSIResult {
	return StochRSI(closes, 14, 14, 3, 3)
}
