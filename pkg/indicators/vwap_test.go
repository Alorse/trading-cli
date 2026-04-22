package indicators

import (
	"testing"
)

func TestVWAP(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		closes   []float64
		volumes  []float64
		expected []float64
	}{
		{
			name:     "happy path with known values",
			highs:    []float64{12, 14, 16},
			lows:     []float64{8, 10, 12},
			closes:   []float64{10, 12, 14},
			volumes:  []float64{100, 100, 100},
			expected: []float64{10, 11, 12},
		},
		{
			name:     "empty input",
			highs:    []float64{},
			lows:     []float64{},
			closes:   []float64{},
			volumes:  []float64{},
			expected: []float64{},
		},
		{
			name:     "mismatched lengths",
			highs:    []float64{10, 12},
			lows:     []float64{8},
			closes:   []float64{9, 11},
			volumes:  []float64{100, 200},
			expected: []float64{},
		},
		{
			name:     "VWAP[0] equals first typical price",
			highs:    []float64{15},
			lows:     []float64{9},
			closes:   []float64{12},
			volumes:  []float64{100},
			expected: []float64{12},
		},
		{
			name:     "simple dataset with exact values",
			highs:    []float64{10, 12, 14, 16},
			lows:     []float64{8, 10, 12, 14},
			closes:   []float64{9, 11, 13, 15},
			volumes:  []float64{100, 200, 300, 400},
			expected: []float64{9, 10.333333, 11.666667, 13},
		},
		{
			name:     "zero volume scenario",
			highs:    []float64{10, 12},
			lows:     []float64{8, 10},
			closes:   []float64{9, 11},
			volumes:  []float64{0, 100},
			expected: []float64{0, 11},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VWAP(tt.highs, tt.lows, tt.closes, tt.volumes)
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
