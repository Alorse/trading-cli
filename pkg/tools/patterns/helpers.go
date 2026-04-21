package patterns

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func computeCandleBodyRatio(close, open, high, low float64) float64 {
	rangeSize := high - low
	if rangeSize == 0 {
		return 0
	}
	return absFloat(close-open) / rangeSize
}

func scoreBullishCandle(change, bodyRatio, close, sma20, rsi, volume float64) int {
	score := 0
	if change > 2.0 {
		score++
	}
	if bodyRatio > 0.6 {
		score++
	}
	if close > sma20 {
		score++
	}
	if rsi > 45 && rsi < 80 {
		score++
	}
	if volume > 1000 {
		score++
	}
	return score
}

func scoreBearishCandle(change, bodyRatio, close, sma20, rsi, volume float64) int {
	score := 0
	if change < -2.0 {
		score++
	}
	if bodyRatio > 0.6 {
		score++
	}
	if close < sma20 {
		score++
	}
	if rsi > 20 && rsi < 55 {
		score++
	}
	if volume > 1000 {
		score++
	}
	return score
}

func scoreAdvancedCandle(bodyRatio, change, volume, rsi, close, ema50 float64) int {
	score := 0
	if bodyRatio > 0.7 {
		score += 2
	} else if bodyRatio > 0.5 {
		score++
	}
	absChange := absFloat(change)
	if absChange >= 10.0 {
		score += 2
	} else if absChange >= 5.0 {
		score++
	}
	if volume > 5000 {
		score++
	}
	if (change > 0 && rsi > 50) || (change < 0 && rsi < 50) {
		score++
	}
	if (change > 0 && close > ema50) || (change < 0 && close < ema50) {
		score++
	}
	return score
}
