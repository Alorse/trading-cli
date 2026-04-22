package patterns

import (
	"testing"
)

// TestComputeCandleBodyRatio tests candle body ratio calculation
func TestComputeCandleBodyRatio(t *testing.T) {
	tests := []struct {
		name     string
		close    float64
		open     float64
		high     float64
		low      float64
		expected float64
	}{
		{"large_body", 105.0, 95.0, 110.0, 90.0, 0.5},
		{"small_body", 100.5, 100.0, 105.0, 95.0, 0.05},
		{"no_range", 100.0, 100.0, 100.0, 100.0, 0.0},
		{"full_range_body", 110.0, 90.0, 110.0, 90.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeCandleBodyRatio(tt.close, tt.open, tt.high, tt.low)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("computeCandleBodyRatio(%f, %f, %f, %f) = %f, want %f",
					tt.close, tt.open, tt.high, tt.low, result, tt.expected)
			}
		})
	}
}

// TestScoreBullishCandle tests bullish candle scoring
func TestScoreBullishCandle(t *testing.T) {
	tests := []struct {
		name        string
		change      float64
		bodyRatio   float64
		close       float64
		sma20       float64
		rsi         float64
		volume      float64
		expectedMin int
		expectedMax int
	}{
		{
			"perfect_bullish",
			3.0, 0.8, 105.0, 100.0, 65.0, 1500.0,
			4, 5,
		},
		{
			"partial_bullish",
			0.5, 0.4, 101.0, 102.0, 70.0, 800.0,
			0, 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scoreBullishCandle(tt.change, tt.bodyRatio, tt.close, tt.sma20, tt.rsi, tt.volume)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("scoreBullishCandle() = %d, want between %d and %d", result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestScoreBearishCandle tests bearish candle scoring
func TestScoreBearishCandle(t *testing.T) {
	tests := []struct {
		name        string
		change      float64
		bodyRatio   float64
		close       float64
		sma20       float64
		rsi         float64
		volume      float64
		expectedMin int
		expectedMax int
	}{
		{
			"perfect_bearish",
			-3.0, 0.8, 95.0, 100.0, 35.0, 1500.0,
			4, 5,
		},
		{
			"partial_bearish",
			-0.5, 0.4, 99.0, 98.0, 30.0, 800.0,
			0, 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scoreBearishCandle(tt.change, tt.bodyRatio, tt.close, tt.sma20, tt.rsi, tt.volume)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("scoreBearishCandle() = %d, want between %d and %d", result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestScoreAdvancedCandle tests advanced candle scoring (0-7 scale)
func TestScoreAdvancedCandle(t *testing.T) {
	tests := []struct {
		name        string
		bodyRatio   float64
		change      float64
		volume      float64
		rsi         float64
		close       float64
		ema50       float64
		expectedMin int
		expectedMax int
	}{
		{
			"perfect_advanced",
			0.75, 1.5, 6000.0, 65.0, 105.0, 100.0,
			4, 7,
		},
		{
			"low_score",
			0.3, 0.2, 3000.0, 40.0, 99.0, 100.5,
			0, 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scoreAdvancedCandle(tt.bodyRatio, tt.change, tt.volume, tt.rsi, tt.close, tt.ema50)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("scoreAdvancedCandle() = %d, want between %d and %d", result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestAbsFloat tests absolute value for floats
func TestAbsFloatPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"positive", 5.0, 5.0},
		{"negative", -5.0, 5.0},
		{"zero", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := absFloat(tt.input)
			if result != tt.expected {
				t.Errorf("absFloat(%f) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}
