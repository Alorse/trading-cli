package indicators

// ParabolicSAR calculates the Parabolic Stop and Reverse indicator.
// Standard parameters: afStart=0.02, afStep=0.02, afMax=0.20.
// First valid value is at index 1; index 0 is 0.
func ParabolicSAR(highs, lows []float64, afStart, afStep, afMax float64) []float64 {
	if len(highs) < 2 || len(lows) < 2 {
		return []float64{}
	}

	result := make([]float64, len(highs))

	// Determine initial trend from first 2 bars
	trendUp := highs[1] > highs[0]

	var ep, sar, af float64
	af = afStart

	if trendUp {
		ep = max(highs[0], highs[1])
		sar = min(lows[0], lows[1])
	} else {
		ep = min(lows[0], lows[1])
		sar = max(highs[0], highs[1])
	}

	result[1] = sar

	for i := 2; i < len(highs); i++ {
		sar = sar + af*(ep-sar)

		if trendUp {
			// SAR cannot exceed the lower of previous 2 lows
			minLow := min(lows[i-1], lows[i-2])
			if sar > minLow {
				sar = minLow
			}

			// Check for reversal
			if lows[i] < sar {
				trendUp = false
				sar = ep
				ep = lows[i]
				af = afStart
			} else {
				// Check for new extreme
				if highs[i] > ep {
					ep = highs[i]
					af += afStep
					if af > afMax {
						af = afMax
					}
				}
			}
		} else {
			// SAR cannot be below the higher of previous 2 highs
			maxHigh := max(highs[i-1], highs[i-2])
			if sar < maxHigh {
				sar = maxHigh
			}

			// Check for reversal
			if highs[i] > sar {
				trendUp = true
				sar = ep
				ep = highs[i]
				af = afStart
			} else {
				// Check for new extreme
				if lows[i] < ep {
					ep = lows[i]
					af += afStep
					if af > afMax {
						af = afMax
					}
				}
			}
		}

		result[i] = sar
	}

	return result
}

// ParabolicSARDefault calls ParabolicSAR with standard parameters.
func ParabolicSARDefault(highs, lows []float64) []float64 {
	return ParabolicSAR(highs, lows, 0.02, 0.02, 0.20)
}
