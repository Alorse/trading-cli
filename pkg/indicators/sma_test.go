package indicators

import (
	"testing"
)

func TestSMA(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		period   int
		expected []float64
	}{
		{
			name:     "SMA(5) of [1,2,3,4,5,6,7]",
			prices:   []float64{1, 2, 3, 4, 5, 6, 7},
			period:   5,
			expected: []float64{0, 0, 0, 0, 3, 4, 5},
		},
		{
			name:     "SMA(2) simple",
			prices:   []float64{10, 20, 30, 40},
			period:   2,
			expected: []float64{0, 15, 25, 35},
		},
		{
			name:     "SMA(1) identity",
			prices:   []float64{5, 10, 15},
			period:   1,
			expected: []float64{5, 10, 15},
		},
		{
			name:     "SMA with single value",
			prices:   []float64{100},
			period:   1,
			expected: []float64{100},
		},
		{
			name:     "SMA period larger than data",
			prices:   []float64{1, 2, 3},
			period:   5,
			expected: []float64{0, 0, 0},
		},
		{
			name:     "SMA with decimals",
			prices:   []float64{1.5, 2.5, 3.5, 4.5},
			period:   2,
			expected: []float64{0, 2.0, 3.0, 4.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SMA(tt.prices, tt.period)
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
