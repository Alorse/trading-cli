package indicators

// DonchianResult holds Donchian Channel results.
type DonchianResult struct {
	Upper  []float64 // Highest high over period
	Lower  []float64 // Lowest low over period
	Middle []float64 // (Upper + Lower) / 2
}

// DonchianChannel calculates the Donchian Channel.
// Upper = highest high over period
// Lower = lowest low over period
// Middle = (Upper + Lower) / 2
// First (period-1) values are 0.
func DonchianChannel(highs, lows []float64, period int) DonchianResult {
	result := DonchianResult{
		Upper:  make([]float64, len(highs)),
		Lower:  make([]float64, len(highs)),
		Middle: make([]float64, len(highs)),
	}

	if len(highs) == 0 || len(lows) == 0 || period <= 0 {
		return result
	}

	for i := period - 1; i < len(highs); i++ {
		// Find highest high in the period
		maxHigh := highs[i-period+1]
		for j := i - period + 1; j <= i; j++ {
			if highs[j] > maxHigh {
				maxHigh = highs[j]
			}
		}

		// Find lowest low in the period
		minLow := lows[i-period+1]
		for j := i - period + 1; j <= i; j++ {
			if lows[j] < minLow {
				minLow = lows[j]
			}
		}

		result.Upper[i] = maxHigh
		result.Lower[i] = minLow
		result.Middle[i] = (maxHigh + minLow) / 2.0
	}

	return result
}
