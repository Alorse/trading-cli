package indicators

// BBResult holds Bollinger Bands results.
type BBResult struct {
	Upper  []float64 // Upper band
	Middle []float64 // SMA (middle band)
	Lower  []float64 // Lower band
	Width  []float64 // (Upper - Lower) / Middle
}

// BollingerBands calculates Bollinger Bands over a given period.
// Standard: period=20, stdDevMult=2.0
// Middle band = SMA(period)
// Upper band = Middle + (stdDevMult * StdDev)
// Lower band = Middle - (stdDevMult * StdDev)
// Width = (Upper - Lower) / Middle
// All slices have same length as prices; leading zeros where not enough data.
func BollingerBands(prices []float64, period int, stdDevMult float64) BBResult {
	result := BBResult{
		Upper:  make([]float64, len(prices)),
		Middle: make([]float64, len(prices)),
		Lower:  make([]float64, len(prices)),
		Width:  make([]float64, len(prices)),
	}

	if len(prices) == 0 || period <= 0 {
		return result
	}

	// Calculate SMA (middle band)
	middle := SMA(prices, period)
	copy(result.Middle, middle)

	// Calculate StdDev
	stdDev := StdDev(prices, period)

	// Calculate upper and lower bands
	for i := 0; i < len(prices); i++ {
		result.Upper[i] = result.Middle[i] + stdDevMult*stdDev[i]
		result.Lower[i] = result.Middle[i] - stdDevMult*stdDev[i]

		// Calculate width
		if result.Middle[i] != 0 {
			result.Width[i] = (result.Upper[i] - result.Lower[i]) / result.Middle[i]
		}
	}

	return result
}
