package indicators

import (
	"testing"
)

func TestMomentum(t *testing.T) {
	tests := []struct {
		name     string
		closes   []float64
		period   int
		expected []float64
	}{
		{
			name:     "Momentum(3) with known values",
			closes:   []float64{10, 12, 15, 13, 18, 20, 22},
			period:   3,
			expected: []float64{0, 0, 0, 3, 6, 5, 9},
		},
		{
			name:     "Momentum(10) standard period",
			closes:   []float64{100, 102, 101, 103, 105, 104, 106, 108, 107, 109, 110},
			period:   10,
			expected: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10},
		},
		{
			name:     "Momentum(1) immediate difference",
			closes:   []float64{5, 10, 15},
			period:   1,
			expected: []float64{0, 5, 5},
		},
		{
			name:     "Momentum with negative result",
			closes:   []float64{50, 45, 40, 35},
			period:   2,
			expected: []float64{0, 0, -10, -10},
		},
		{
			name:     "empty input",
			closes:   []float64{},
			period:   10,
			expected: []float64{},
		},
		{
			name:     "period zero",
			closes:   []float64{10, 20, 30},
			period:   0,
			expected: []float64{},
		},
		{
			name:     "period negative",
			closes:   []float64{10, 20, 30},
			period:   -1,
			expected: []float64{},
		},
		{
			name:     "period larger than input",
			closes:   []float64{10, 20, 30},
			period:   5,
			expected: []float64{0, 0, 0},
		},
		{
			name:     "single value",
			closes:   []float64{42},
			period:   1,
			expected: []float64{0},
		},
		{
			name:     "Momentum(2) with decimals",
			closes:   []float64{1.5, 2.5, 3.5, 4.5},
			period:   2,
			expected: []float64{0, 0, 2.0, 2.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Momentum(tt.closes, tt.period)
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
