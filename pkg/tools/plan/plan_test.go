package plan

import (
	"testing"
)

// TestComputeStockScore verifies stock score calculation with known inputs
func TestComputeStockScore(t *testing.T) {
	// Test case: all indicators bullish
	values := map[string]interface{}{
		"close":           100.0,
		"EMA50":           95.0,
		"EMA200":          90.0,
		"RSI":             65.0,
		"MACD.macd":       2.0,
		"MACD.signal":     1.0,
		"volume":          1000.0,
		"volume.SMA20":    500.0,
		"ADX":             30.0,
		"Recommend.All":   1.0,
		"BB.lower":        98.0,
		"BB.upper":        102.0,
	}

	score, components := computeStockScore(values)

	// Expected: 20 + 20 + 20 + 20 + 20 = 100
	if score != 100 {
		t.Errorf("Expected score 100, got %d", score)
	}

	if components.EMAAlignment != 20 {
		t.Errorf("Expected EMAAlignment 20, got %d", components.EMAAlignment)
	}

	if components.RSIScore != 20 {
		t.Errorf("Expected RSIScore 20, got %d", components.RSIScore)
	}

	if components.MACDScore != 20 {
		t.Errorf("Expected MACDScore 20, got %d", components.MACDScore)
	}

	if components.VolumeScore != 20 {
		t.Errorf("Expected VolumeScore 20, got %d", components.VolumeScore)
	}

	if components.ADXScore != 20 {
		t.Errorf("Expected ADXScore 20, got %d", components.ADXScore)
	}
}

// TestComputeStockScoreBearish tests stock score with bearish indicators
func TestComputeStockScoreBearish(t *testing.T) {
	// Test case: all indicators bearish
	values := map[string]interface{}{
		"close":           50.0,
		"EMA50":           55.0,
		"EMA200":          60.0,
		"RSI":             25.0,
		"MACD.macd":       -1.0,
		"MACD.signal":     -0.5,
		"volume":          200.0,
		"volume.SMA20":    500.0,
		"ADX":             15.0,
		"Recommend.All":   -1.0,
		"BB.lower":        45.0,
		"BB.upper":        55.0,
	}

	score, _ := computeStockScore(values)

	// Expected: mostly 0s, close > BB.lower = 10
	if score != 10 {
		t.Errorf("Expected score 10, got %d", score)
	}
}

// TestGradeScore tests letter grade calculation
func TestGradeScore(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{85, "A"},
		{75, "B"},
		{65, "C"},
		{55, "D"},
		{45, "F"},
	}

	for _, tc := range tests {
		grade := gradeScore(tc.score)
		if grade != tc.expected {
			t.Errorf("Score %d: expected grade %s, got %s", tc.score, tc.expected, grade)
		}
	}
}

// TestComputeTradeQuality verifies trade quality calculation
func TestComputeTradeQuality(t *testing.T) {
	// Test case: strong quality
	score := 75
	rr2 := 2.5
	volumeRatio := 1.8
	stopLossPct := 3.0
	rsi := 55.0

	quality, breakdown := computeTradeQuality(score, rr2, volumeRatio, stopLossPct, rsi)

	// Expected: 30 + 30 + 20 + 10 + 10 = 100
	if quality != 100 {
		t.Errorf("Expected quality 100, got %d", quality)
	}

	if breakdown.Structure != 30 {
		t.Errorf("Expected Structure 30, got %d", breakdown.Structure)
	}

	if breakdown.RewardRisk != 30 {
		t.Errorf("Expected RewardRisk 30, got %d", breakdown.RewardRisk)
	}

	if breakdown.Volume != 20 {
		t.Errorf("Expected Volume 20, got %d", breakdown.Volume)
	}

	if breakdown.StopSize != 10 {
		t.Errorf("Expected StopSize 10, got %d", breakdown.StopSize)
	}

	if breakdown.Liquidity != 10 {
		t.Errorf("Expected Liquidity 10, got %d", breakdown.Liquidity)
	}
}

// TestComputeTradeQualityPoor tests quality with poor conditions
func TestComputeTradeQualityPoor(t *testing.T) {
	score := 40
	rr2 := 0.8
	volumeRatio := 0.5
	stopLossPct := 12.0
	rsi := 80.0

	quality, breakdown := computeTradeQuality(score, rr2, volumeRatio, stopLossPct, rsi)

	// Expected: 10 + 10 + 0 + 0 + 0 = 20
	if quality != 20 {
		t.Errorf("Expected quality 20, got %d", quality)
	}

	if breakdown.Structure != 10 {
		t.Errorf("Expected Structure 10, got %d", breakdown.Structure)
	}

	if breakdown.RewardRisk != 10 {
		t.Errorf("Expected RewardRisk 10, got %d", breakdown.RewardRisk)
	}
}

// TestGetRecommendation verifies recommendation logic
func TestGetRecommendation(t *testing.T) {
	tests := []struct {
		score       int
		quality     int
		rr2         float64
		expected    string
	}{
		{75, 70, 2.5, "QUALIFIED"},
		{72, 65, 2.1, "QUALIFIED"},
		{75, 55, 1.5, "CONDITIONAL"},
		{60, 50, 1.2, "WATCHLIST"},
		{40, 30, 0.8, "AVOID"},
	}

	for _, tc := range tests {
		rec := getRecommendation(tc.score, tc.quality, tc.rr2)
		if rec != tc.expected {
			t.Errorf("Score %d, Quality %d, RR2 %.1f: expected %s, got %s",
				tc.score, tc.quality, tc.rr2, tc.expected, rec)
		}
	}
}

// TestFibonacciLevels tests Fibonacci level calculation
func TestFibonacciLevels(t *testing.T) {
	swingHigh := 100.0
	swingLow := 80.0

	retLevels, extLevels := computeFibonacciLevels(swingHigh, swingLow, "uptrend")

	// Check retracement levels
	if len(retLevels) != 7 {
		t.Errorf("Expected 7 retracement levels, got %d", len(retLevels))
	}

	// 0.618 level should be: 80 + 0.618 * 20 = 92.36
	expectedLevel := 80.0 + (0.618 * (swingHigh - swingLow))
	actual618 := retLevels[4].Level // Index 4 = ratio 0.618
	if abs(actual618-expectedLevel) > 0.01 {
		t.Errorf("0.618 level: expected %.2f, got %.2f", expectedLevel, actual618)
	}

	// Check extension levels
	if len(extLevels) != 3 {
		t.Errorf("Expected 3 extension levels, got %d", len(extLevels))
	}

	// 1.618 extension: 80 + 1.618 * 20 = 112.36
	expected1618 := 80.0 + (1.618 * (swingHigh - swingLow))
	actual1618 := extLevels[1].Level // Index 1 = ratio 1.618
	if abs(actual1618-expected1618) > 0.01 {
		t.Errorf("1.618 extension: expected %.2f, got %.2f", expected1618, actual1618)
	}
}

// TestFibonacciLevelsDowntrend tests Fibonacci levels in downtrend
func TestFibonacciLevelsDowntrend(t *testing.T) {
	swingHigh := 100.0
	swingLow := 80.0

	retLevels, _ := computeFibonacciLevels(swingHigh, swingLow, "downtrend")

	// For downtrend, 0.618 level should be: 100 - 0.618 * 20 = 87.64
	expectedLevel := 100.0 - (0.618 * (swingHigh - swingLow))
	actual618 := retLevels[4].Level
	if abs(actual618-expectedLevel) > 0.01 {
		t.Errorf("Downtrend 0.618 level: expected %.2f, got %.2f", expectedLevel, actual618)
	}
}

// TestComputeGoldenPocket tests golden pocket zone calculation
func TestComputeGoldenPocket(t *testing.T) {
	swingHigh := 100.0
	swingLow := 80.0

	gp := computeGoldenPocket(swingHigh, swingLow, "uptrend")

	// Expected: lower = 80 + 0.618*20 = 92.36, upper = 80 + 0.786*20 = 95.72
	expectedLower := 80.0 + (0.618 * 20.0)
	expectedUpper := 80.0 + (0.786 * 20.0)

	if abs(gp.Lower-expectedLower) > 0.01 {
		t.Errorf("Golden pocket lower: expected %.2f, got %.2f", expectedLower, gp.Lower)
	}

	if abs(gp.Upper-expectedUpper) > 0.01 {
		t.Errorf("Golden pocket upper: expected %.2f, got %.2f", expectedUpper, gp.Upper)
	}
}

// TestComputeCurrentDepth tests current depth calculation
func TestComputeCurrentDepth(t *testing.T) {
	swingHigh := 100.0
	swingLow := 80.0

	// Uptrend: current at 90 = (90-80)/(100-80)*100 = 50%
	depth := computeCurrentDepth(90.0, swingHigh, swingLow, "uptrend")
	if abs(depth-50.0) > 0.1 {
		t.Errorf("Uptrend depth at 90: expected 50.0, got %.2f", depth)
	}

	// Downtrend: current at 90 = (100-90)/(100-80)*100 = 50%
	depth = computeCurrentDepth(90.0, swingHigh, swingLow, "downtrend")
	if abs(depth-50.0) > 0.1 {
		t.Errorf("Downtrend depth at 90: expected 50.0, got %.2f", depth)
	}
}

// TestMapLookback tests lookback string mapping
func TestMapLookback(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1M", "1mo"},
		{"3M", "3mo"},
		{"6M", "6mo"},
		{"52W", "1y"},
		{"ALL", "5y"},
		{"unknown", "1y"},
	}

	for _, tc := range tests {
		result := mapLookback(tc.input)
		if result != tc.expected {
			t.Errorf("mapLookback(%s): expected %s, got %s", tc.input, tc.expected, result)
		}
	}
}
