package volume

import (
	"testing"
)

// TestComputeVolumeRatio tests volume ratio calculation
func TestComputeVolumeRatio(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		avg20    float64
		expected float64
	}{
		{"double_volume", 2000.0, 1000.0, 2.0},
		{"triple_volume", 3000.0, 1000.0, 3.0},
		{"half_volume", 500.0, 1000.0, 0.5},
		{"equal_volume", 1000.0, 1000.0, 1.0},
		{"zero_avg", 1000.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeVolumeRatio(tt.current, tt.avg20)
			if result != tt.expected {
				t.Errorf("computeVolumeRatio(%f, %f) = %f, want %f", tt.current, tt.avg20, result, tt.expected)
			}
		})
	}
}

// TestComputeVolumeStrength tests volume strength calculation
func TestComputeVolumeStrength(t *testing.T) {
	tests := []struct {
		name     string
		ratio    float64
		expected float64
	}{
		{"cap_at_10", 15.0, 10.0},
		{"under_10", 5.0, 5.0},
		{"double", 2.0, 2.0},
		{"normal", 1.0, 1.0},
		{"weak", 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeVolumeStrength(tt.ratio)
			if result != tt.expected {
				t.Errorf("computeVolumeStrength(%f) = %f, want %f", tt.ratio, result, tt.expected)
			}
		})
	}
}

// TestComputeBreakoutType tests breakout type determination
func TestComputeBreakoutType(t *testing.T) {
	tests := []struct {
		name     string
		change   float64
		expected string
	}{
		{"bullish_positive", 5.0, "bullish"},
		{"bullish_zero_boundary", 0.1, "bullish"},
		{"bearish_negative", -5.0, "bearish"},
		{"bearish_zero_boundary", -0.1, "bearish"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeBreakoutType(tt.change)
			if result != tt.expected {
				t.Errorf("computeBreakoutType(%f) = %s, want %s", tt.change, result, tt.expected)
			}
		})
	}
}

// TestComputeVolumeAssessment tests volume strength assessment
func TestComputeVolumeAssessment(t *testing.T) {
	tests := []struct {
		name     string
		ratio    float64
		expected string
	}{
		{"very_strong", 3.5, "VERY STRONG"},
		{"strong", 2.5, "STRONG"},
		{"medium", 1.7, "MEDIUM"},
		{"normal", 1.2, "NORMAL"},
		{"weak", 0.8, "WEAK"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeVolumeAssessment(tt.ratio)
			if result != tt.expected {
				t.Errorf("computeVolumeAssessment(%f) = %s, want %s", tt.ratio, result, tt.expected)
			}
		})
	}
}

// TestGenerateSignals tests signal generation
func TestGenerateSignals(t *testing.T) {
	tests := []struct {
		name        string
		change      float64
		ratio       float64
		close       float64
		bbUpper     float64
		bbLower     float64
		expectedMin int
		expectedMax int
		shouldHave  string
	}{
		{
			"strong_breakout",
			5.0, 2.5, 110.0, 100.0, 80.0,
			1, 2, "STRONG BREAKOUT",
		},
		{
			"volume_divergence",
			0.5, 2.5, 90.0, 100.0, 80.0,
			1, 1, "VOLUME DIVERGENCE",
		},
		{
			"weak_signal",
			0.5, 0.8, 90.0, 100.0, 80.0,
			1, 1, "WEAK SIGNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signals := generateSignals(tt.change, tt.ratio, tt.close, tt.bbUpper, tt.bbLower)

			if len(signals) < tt.expectedMin || len(signals) > tt.expectedMax {
				t.Errorf("generateSignals() returned %d signals, want between %d and %d", len(signals), tt.expectedMin, tt.expectedMax)
			}

			found := false
			for _, sig := range signals {
				if sig == tt.shouldHave {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("generateSignals() should contain '%s', got %v", tt.shouldHave, signals)
			}
		})
	}
}
