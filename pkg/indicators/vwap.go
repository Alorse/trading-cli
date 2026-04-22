package indicators

// VWAP calculates the Volume Weighted Average Price.
// It is a cumulative indicator: each bar uses the sum of typical price * volume
// divided by cumulative volume from the start of the series.
// Returns a slice of the same length as inputs. Returns empty slice if inputs
// are empty or have mismatched lengths.
func VWAP(highs, lows, closes, volumes []float64) []float64 {
	if len(highs) == 0 || len(lows) == 0 || len(closes) == 0 || len(volumes) == 0 {
		return []float64{}
	}
	if len(highs) != len(lows) || len(highs) != len(closes) || len(highs) != len(volumes) {
		return []float64{}
	}

	result := make([]float64, len(highs))
	cumTPVol := 0.0
	cumVol := 0.0

	for i := 0; i < len(highs); i++ {
		tp := (highs[i] + lows[i] + closes[i]) / 3.0
		cumTPVol += tp * volumes[i]
		cumVol += volumes[i]
		if cumVol == 0 {
			result[i] = 0
			continue
		}
		result[i] = cumTPVol / cumVol
	}

	return result
}
