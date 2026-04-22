package indicators

// OBV calculates the On-Balance Volume.
// It is a cumulative indicator that uses volume flow to predict changes in price.
// OBV[i] = OBV[i-1] + volumes[i]  if close[i] > close[i-1]
// OBV[i] = OBV[i-1] - volumes[i]  if close[i] < close[i-1]
// OBV[i] = OBV[i-1]               if close[i] == close[i-1]
// Returns a slice of the same length as the inputs.
func OBV(closes []float64, volumes []float64) []float64 {
	if len(closes) == 0 || len(volumes) == 0 || len(closes) != len(volumes) {
		return []float64{}
	}

	result := make([]float64, len(closes))
	result[0] = volumes[0]

	for i := 1; i < len(closes); i++ {
		if closes[i] > closes[i-1] {
			result[i] = result[i-1] + volumes[i]
		} else if closes[i] < closes[i-1] {
			result[i] = result[i-1] - volumes[i]
		} else {
			result[i] = result[i-1]
		}
	}

	return result
}
