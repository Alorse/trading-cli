package indicators

import (
	"testing"
)

func TestOBV(t *testing.T) {
	tests := []struct {
		name     string
		closes   []float64
		volumes  []float64
		expected []float64
	}{
		{
			name:     "basic up and down",
			closes:   []float64{10, 11, 10, 12, 11},
			volumes:  []float64{100, 200, 150, 300, 100},
			expected: []float64{100, 300, 150, 450, 350},
		},
		{
			name:     "empty input",
			closes:   []float64{},
			volumes:  []float64{},
			expected: []float64{},
		},
		{
			name:     "single data point",
			closes:   []float64{50},
			volumes:  []float64{1000},
			expected: []float64{1000},
		},
		{
			name:     "flat market",
			closes:   []float64{10, 10, 10, 10},
			volumes:  []float64{100, 200, 150, 300},
			expected: []float64{100, 100, 100, 100},
		},
		{
			name:     "all up days",
			closes:   []float64{1, 2, 3, 4, 5},
			volumes:  []float64{10, 20, 30, 40, 50},
			expected: []float64{10, 30, 60, 100, 150},
		},
		{
			name:     "all down days",
			closes:   []float64{5, 4, 3, 2, 1},
			volumes:  []float64{10, 20, 30, 40, 50},
			expected: []float64{10, -10, -40, -80, -130},
		},
		{
			name:     "mixed with flat",
			closes:   []float64{10, 12, 12, 11, 13},
			volumes:  []float64{100, 50, 200, 150, 300},
			expected: []float64{100, 150, 150, 0, 300},
		},
		{
			name:     "mismatch length",
			closes:   []float64{10, 11, 12},
			volumes:  []float64{100, 200},
			expected: []float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OBV(tt.closes, tt.volumes)
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
