package indicators

import (
	"testing"
)

func TestVWMA(t *testing.T) {
	tests := []struct {
		name     string
		closes   []float64
		volumes  []float64
		period   int
		expected []float64
	}{
		{
			name:     "VWMA(3) with equal volumes matches SMA",
			closes:   []float64{10, 20, 30, 40},
			volumes:  []float64{1, 1, 1, 1},
			period:   3,
			expected: []float64{0, 0, 20, 30},
		},
		{
			name:     "VWMA(3) with weighted volumes",
			closes:   []float64{10, 20, 30, 40},
			volumes:  []float64{1, 2, 3, 4},
			period:   3,
			expected: []float64{0, 0, 23.3333, 32.2222},
		},
		{
			name:     "VWMA(2) simple",
			closes:   []float64{10, 20, 30, 40},
			volumes:  []float64{1, 2, 3, 4},
			period:   2,
			expected: []float64{0, 16.6667, 26.0, 35.7143},
		},
		{
			name:     "VWMA(1) identity",
			closes:   []float64{5, 10, 15},
			volumes:  []float64{100, 200, 300},
			period:   1,
			expected: []float64{5, 10, 15},
		},
		{
			name:     "VWMA period larger than data",
			closes:   []float64{10, 20, 30},
			volumes:  []float64{1, 2, 3},
			period:   5,
			expected: []float64{0, 0, 0},
		},
		{
			name:     "zero total volume in window",
			closes:   []float64{10, 20, 30, 40},
			volumes:  []float64{0, 0, 1, 1},
			period:   2,
			expected: []float64{0, 0, 30, 35},
		},
		{
			name:     "empty closes",
			closes:   []float64{},
			volumes:  []float64{},
			period:   3,
			expected: []float64{},
		},
		{
			name:     "mismatched lengths",
			closes:   []float64{10, 20, 30},
			volumes:  []float64{1, 2},
			period:   2,
			expected: []float64{},
		},
		{
			name:     "period zero",
			closes:   []float64{10, 20, 30},
			volumes:  []float64{1, 2, 3},
			period:   0,
			expected: []float64{},
		},
		{
			name:     "period negative",
			closes:   []float64{10, 20, 30},
			volumes:  []float64{1, 2, 3},
			period:   -1,
			expected: []float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VWMA(tt.closes, tt.volumes, tt.period)
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
