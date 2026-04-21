package indicators

// SupertrendResult holds Supertrend line and direction values.
type SupertrendResult struct {
	Value     []float64 // Supertrend line value
	Direction []int     // +1 = bullish, -1 = bearish, 0 = no data
}

// Supertrend calculates the Supertrend indicator.
// Standard: period=10, multiplier=3.0
// Upper Band = (high + low)/2 + multiplier * ATR
// Lower Band = (high + low)/2 - multiplier * ATR
// Direction flips when close crosses the band.
func Supertrend(highs, lows, closes []float64, period int, multiplier float64) SupertrendResult {
	result := SupertrendResult{
		Value:     make([]float64, len(highs)),
		Direction: make([]int, len(highs)),
	}

	if len(highs) == 0 || len(lows) == 0 || len(closes) == 0 || period <= 0 {
		return result
	}

	// Calculate ATR
	atr := ATR(highs, lows, closes, period)

	// Calculate basic bands
	upperBand := make([]float64, len(highs))
	lowerBand := make([]float64, len(highs))

	for i := 0; i < len(highs); i++ {
		hl2 := (highs[i] + lows[i]) / 2.0
		upperBand[i] = hl2 + multiplier*atr[i]
		lowerBand[i] = hl2 - multiplier*atr[i]
	}

	// Calculate final bands with continuity
	finalUpperBand := make([]float64, len(highs))
	finalLowerBand := make([]float64, len(highs))

	for i := 0; i < len(highs); i++ {
		if i == 0 {
			finalUpperBand[i] = upperBand[i]
			finalLowerBand[i] = lowerBand[i]
		} else {
			// Upper band should not be lower than previous bar
			finalUpperBand[i] = upperBand[i]
			if finalUpperBand[i] < finalUpperBand[i-1] {
				finalUpperBand[i] = finalUpperBand[i-1]
			}

			// Lower band should not be higher than previous bar
			finalLowerBand[i] = lowerBand[i]
			if finalLowerBand[i] > finalLowerBand[i-1] {
				finalLowerBand[i] = finalLowerBand[i-1]
			}
		}
	}

	// Determine Supertrend and direction
	var currentDirection int = 0
	for i := period; i < len(closes); i++ {
		if i == period {
			// Initialize direction based on close vs bands
			if closes[i] <= finalUpperBand[i] {
				currentDirection = -1
				result.Value[i] = finalUpperBand[i]
			} else {
				currentDirection = 1
				result.Value[i] = finalLowerBand[i]
			}
		} else {
			// Check for direction reversal
			if currentDirection == 1 {
				if closes[i] <= finalLowerBand[i] {
					currentDirection = -1
					result.Value[i] = finalUpperBand[i]
				} else {
					result.Value[i] = finalLowerBand[i]
				}
			} else {
				if closes[i] >= finalUpperBand[i] {
					currentDirection = 1
					result.Value[i] = finalLowerBand[i]
				} else {
					result.Value[i] = finalUpperBand[i]
				}
			}
		}
		result.Direction[i] = currentDirection
	}

	return result
}
