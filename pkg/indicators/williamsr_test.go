package indicators

import (
	"testing"
)

func TestWilliamsR(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		closes   []float64
		period   int
		expected []float64
	}{
		{
			name:     "WilliamsR(3) basic",
			highs:    []float64{10, 12, 11, 13, 14},
			lows:     []float64{8, 9, 9, 10, 11},
			closes:   []float64{9, 11, 10, 12, 13},
			period:   3,
			expected: []float64{0, 0, -50, -25, -20},
		},
		{
			name:     "empty input",
			highs:    []float64{},
			lows:     []float64{},
			closes:   []float64{},
			period:   14,
			expected: []float64{},
		},
		{
			name:     "period <= 0",
			highs:    []float64{10, 12, 11},
			lows:     []float64{8, 9, 9},
			closes:   []float64{9, 11, 10},
			period:   0,
			expected: []float64{},
		},
		{
			name:     "period larger than data",
			highs:    []float64{10, 11, 12},
			lows:     []float64{8, 9, 10},
			closes:   []float64{9, 10, 11},
			period:   5,
			expected: []float64{0, 0, 0},
		},
		{
			name:     "flat market returns -50",
			highs:    []float64{10, 10, 10, 10},
			lows:     []float64{10, 10, 10, 10},
			closes:   []float64{10, 10, 10, 10},
			period:   2,
			expected: []float64{0, -50, -50, -50},
		},
		{
			name:     "oversold at -100",
			highs:    []float64{10, 12},
			lows:     []float64{8, 8},
			closes:   []float64{10, 8},
			period:   2,
			expected: []float64{0, -100},
		},
		{
			name:     "overbought at 0",
			highs:    []float64{10, 12},
			lows:     []float64{8, 9},
			closes:   []float64{10, 12},
			period:   2,
			expected: []float64{0, 0},
		},
		{
			name:     "single bar",
			highs:    []float64{10},
			lows:     []float64{8},
			closes:   []float64{9},
			period:   1,
			expected: []float64{-50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WilliamsR(tt.highs, tt.lows, tt.closes, tt.period)

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
