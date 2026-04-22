package indicators

import (
	"math"
)

// ADXResult holds the results of ADX calculation.
type ADXResult struct {
	ADX     []float64
	PlusDI  []float64
	MinusDI []float64
	DX      []float64
}

// ADX calculates the Average Directional Index over a given period.
// True Range, +DM, and -DM are computed from highs, lows, and closes.
// +DI and -DI start at index period using Wilder smoothing.
// DX follows from +DI and -DI at the same indices.
// First ADX value is at index 2*period-1 (simple average of first period DX values).
// All leading values are 0. Result slices match input length.
func ADX(highs, lows, closes []float64, period int) ADXResult {
	n := len(highs)
	if n == 0 || len(lows) != n || len(closes) != n || period <= 0 {
		return ADXResult{}
	}

	adx := make([]float64, n)
	plusDI := make([]float64, n)
	minusDI := make([]float64, n)
	dx := make([]float64, n)

	if n <= period {
		return ADXResult{
			ADX:     adx,
			PlusDI:  plusDI,
			MinusDI: minusDI,
			DX:      dx,
		}
	}

	// Calculate True Range for each bar
	trueRanges := make([]float64, n)
	for i := 0; i < n; i++ {
		if i == 0 {
			trueRanges[i] = highs[i] - lows[i]
		} else {
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

	// Calculate +DM and -DM
	plusDM := make([]float64, n)
	minusDM := make([]float64, n)
	for i := 1; i < n; i++ {
		upMove := highs[i] - highs[i-1]
		downMove := lows[i-1] - lows[i]

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}
	}

	// First smoothed values at index period
	var smoothedTR, smoothedPlusDM, smoothedMinusDM float64
	for i := 0; i < period; i++ {
		smoothedTR += trueRanges[i]
	}
	for i := 1; i <= period; i++ {
		smoothedPlusDM += plusDM[i]
		smoothedMinusDM += minusDM[i]
	}
	smoothedTR /= float64(period)
	smoothedPlusDM /= float64(period)
	smoothedMinusDM /= float64(period)

	// Calculate DI and DX at index period
	if smoothedTR == 0 {
		plusDI[period] = 0
		minusDI[period] = 0
	} else {
		plusDI[period] = 100.0 * smoothedPlusDM / smoothedTR
		minusDI[period] = 100.0 * smoothedMinusDM / smoothedTR
	}

	diSum := plusDI[period] + minusDI[period]
	if diSum == 0 {
		dx[period] = 0
	} else {
		dx[period] = 100.0 * math.Abs(plusDI[period]-minusDI[period]) / diSum
	}

	// Calculate DI and DX for remaining indices using Wilder smoothing
	for i := period + 1; i < n; i++ {
		smoothedTR = (smoothedTR*(float64(period)-1) + trueRanges[i]) / float64(period)
		smoothedPlusDM = (smoothedPlusDM*(float64(period)-1) + plusDM[i]) / float64(period)
		smoothedMinusDM = (smoothedMinusDM*(float64(period)-1) + minusDM[i]) / float64(period)

		if smoothedTR == 0 {
			plusDI[i] = 0
			minusDI[i] = 0
		} else {
			plusDI[i] = 100.0 * smoothedPlusDM / smoothedTR
			minusDI[i] = 100.0 * smoothedMinusDM / smoothedTR
		}

		diSum = plusDI[i] + minusDI[i]
		if diSum == 0 {
			dx[i] = 0
		} else {
			dx[i] = 100.0 * math.Abs(plusDI[i]-minusDI[i]) / diSum
		}
	}

	// Calculate ADX
	if n >= 2*period {
		// First ADX at index 2*period-1: simple average of DX values from period to 2*period-1
		var sumDX float64
		for i := period; i < 2*period; i++ {
			sumDX += dx[i]
		}
		adx[2*period-1] = sumDX / float64(period)

		// Apply Wilder smoothing for remaining ADX values
		for i := 2 * period; i < n; i++ {
			adx[i] = (adx[i-1]*(float64(period)-1) + dx[i]) / float64(period)
		}
	}

	return ADXResult{
		ADX:     adx,
		PlusDI:  plusDI,
		MinusDI: minusDI,
		DX:      dx,
	}
}
