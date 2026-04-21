package indicators

import (
	"testing"
)

func TestEMA(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		period   int
		expected []float64
	}{
		{
			name:     "EMA(3) with [1,2,3,4,5,6,7]",
			prices:   []float64{1, 2, 3, 4, 5, 6, 7},
			period:   3,
			expected: []float64{0, 0, 2, 3, 4, 5, 6}, // Seeded with SMA(3)=2, multiplier=0.5
		},
		{
			name:     "EMA(2) simple",
			prices:   []float64{10, 20, 30, 40},
			period:   2,
			expected: []float64{0, 15, 25, 35}, // Seeded with SMA(2)=15, multiplier=2/3
		},
		{
			name:     "EMA with single value",
			prices:   []float64{100},
			period:   1,
			expected: []float64{100},
		},
		{
			name:     "EMA period larger than data",
			prices:   []float64{1, 2, 3},
			period:   5,
			expected: []float64{0, 0, 0},
		},
		{
			name:     "EMA(4) trend",
			prices:   []float64{44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84},
			period:   4,
			expected: []float64{0, 0, 0, 44.0475, 44.1605, 44.4283, 44.6970, 44.9862, 45.3277},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EMA(tt.prices, tt.period)
			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}
			for i, val := range result {
				if !floatEqual(val, tt.expected[i]) {
					t.Errorf("at index %d: got %f, want %f", i, val, tt.expected[i])
				}
			}
		})
	}
}
