package indicators

import (
	"testing"
)

func TestUltimateOscillator(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		closes   []float64
		expected []float64
	}{
		{
			name:     "happy path with 30 bars",
			highs:    []float64{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12},
			lows:     []float64{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			closes:   []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
			expected: nil, // validated separately
		},
		{
			name:     "empty input",
			highs:    []float64{},
			lows:     []float64{},
			closes:   []float64{},
			expected: []float64{},
		},
		{
			name:     "single data point",
			highs:    []float64{10},
			lows:     []float64{8},
			closes:   []float64{9},
			expected: []float64{},
		},
		{
			name:     "fewer than 28 bars",
			highs:    []float64{12, 14, 14, 14, 14, 14, 14, 14, 14, 14},
			lows:     []float64{8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			closes:   []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
			expected: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "exact values close at high",
			highs:    []float64{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12},
			lows:     []float64{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			closes:   []float64{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12},
			expected: nil, // validated separately
		},
		{
			name:     "exact values close at low",
			highs:    []float64{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12},
			lows:     []float64{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			closes:   []float64{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			expected: nil, // validated separately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UltimateOscillator(tt.highs, tt.lows, tt.closes)

			if tt.expected != nil {
				if len(result) != len(tt.expected) {
					t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
				}
				for i, val := range result {
					if !floatEqual(val, tt.expected[i]) {
						t.Errorf("at index %d: got %f, want %f", i, val, tt.expected[i])
					}
				}
				return
			}

			// Validate length matches input
			if len(result) != len(tt.closes) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.closes))
			}

			// Validate leading zeros before index 27
			for i := 0; i < 27 && i < len(result); i++ {
				if result[i] != 0 {
					t.Errorf("at index %d (before first valid): got %f, want 0", i, result[i])
				}
			}

			// Validate all values are within 0..100
			for i := 27; i < len(result); i++ {
				if result[i] < 0 || result[i] > 100 {
					t.Errorf("at index %d: got %f, want in range [0, 100]", i, result[i])
				}
			}

			// Validate exact known values for specific datasets
			switch tt.name {
			case "happy path with 30 bars":
				// BP=2, TR=4 for all bars -> avg=0.5 for all periods -> UO=50
				for i := 27; i < len(result); i++ {
					if !floatEqual(result[i], 50.0) {
						t.Errorf("at index %d: got %f, want 50.0", i, result[i])
					}
				}
			case "exact values close at high":
				// BP=4, TR=4 for all bars -> avg=1.0 for all periods -> UO=100
				for i := 27; i < len(result); i++ {
					if !floatEqual(result[i], 100.0) {
						t.Errorf("at index %d: got %f, want 100.0", i, result[i])
					}
				}
			case "exact values close at low":
				// BP=0, TR=4 for all bars -> avg=0.0 for all periods -> UO=0
				for i := 27; i < len(result); i++ {
					if !floatEqual(result[i], 0.0) {
						t.Errorf("at index %d: got %f, want 0.0", i, result[i])
					}
				}
			}
		})
	}
}

func TestUltimateOscillatorRangeBounds(t *testing.T) {
	// Generate random-like data and verify all outputs are within [0, 100]
	highs := []float64{50, 52, 48, 51, 49, 53, 47, 50, 52, 48, 51, 49, 53, 47, 50, 52, 48, 51, 49, 53, 47, 50, 52, 48, 51, 49, 53, 47, 50, 52}
	lows := []float64{48, 50, 46, 49, 47, 51, 45, 48, 50, 46, 49, 47, 51, 45, 48, 50, 46, 49, 47, 51, 45, 48, 50, 46, 49, 47, 51, 45, 48, 50}
	closes := []float64{49, 51, 47, 50, 48, 52, 46, 49, 51, 47, 50, 48, 52, 46, 49, 51, 47, 50, 48, 52, 46, 49, 51, 47, 50, 48, 52, 46, 49, 51}

	result := UltimateOscillator(highs, lows, closes)

	if len(result) != len(closes) {
		t.Fatalf("length mismatch: got %d, want %d", len(result), len(closes))
	}

	for i := 27; i < len(result); i++ {
		if result[i] < 0 || result[i] > 100 {
			t.Errorf("at index %d: got %f, want in range [0, 100]", i, result[i])
		}
	}
}
