package indicators

import (
	"math"
)

// ATR calculates the Average True Range.
// True Range = max(high-low, abs(high-prevClose), abs(low-prevClose))
// ATR = Wilder smoothed TR over period.
// First period values are 0.
func ATR(highs, lows, closes []float64, period int) []float64 {
	if len(highs) == 0 || len(lows) == 0 || len(closes) == 0 || period <= 0 {
		return []float64{}
	}

	result := make([]float64, len(highs))

	// Calculate True Range for each bar
	trueRanges := make([]float64, len(highs))
	for i := 0; i < len(highs); i++ {
		if i == 0 {
			// First bar: TR = high - low
			trueRanges[i] = highs[i] - lows[i]
		} else {
			// TR = max(high - low, abs(high - prevClose), abs(low - prevClose))
			highLow := highs[i] - lows[i]
			highClose := math.Abs(highs[i] - closes[i-1])
			lowClose := math.Abs(lows[i] - closes[i-1])

			trueRanges[i] = highLow
			if highClose > trueRanges[i] {
				trueRanges[i] = highClose
			}
			if lowClose > trueRanges[i] {
				trueRanges[i] = lowClose
			}
		}
	}

	// Calculate ATR using Wilder's smoothing
	if len(trueRanges) > period {
		// First ATR is simple average of first period values
		sum := 0.0
		for i := 0; i < period; i++ {
			sum += trueRanges[i]
		}
		atr := sum / float64(period)
		result[period-1] = atr

		// Apply Wilder's smoothing for remaining values
		for i := period; i < len(trueRanges); i++ {
			atr = (atr*(float64(period)-1) + trueRanges[i]) / float64(period)
			result[i] = atr
		}
	}

	return result
}
