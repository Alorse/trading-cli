package indicators

import (
	"testing"
)

func TestHullMA(t *testing.T) {
	tests := []struct {
		name     string
		closes   []float64
		period   int
		expected []float64
	}{
		{
			name:     "HullMA(4) of [1,2,3,4,5,6,7]",
			closes:   []float64{1, 2, 3, 4, 5, 6, 7},
			period:   4,
			expected: []float64{0, 0, 0, 0, 5, 6, 7},
		},
		{
			name:     "HullMA(1) identity-like",
			closes:   []float64{5, 10, 15},
			period:   1,
			expected: []float64{5, 10, 15},
		},
		{
			name:     "HullMA with single value",
			closes:   []float64{100},
			period:   1,
			expected: []float64{100},
		},
		{
			name:     "HullMA empty input",
			closes:   []float64{},
			period:   4,
			expected: []float64{},
		},
		{
			name:     "HullMA period <= 0",
			closes:   []float64{1, 2, 3},
			period:   0,
			expected: []float64{},
		},
		{
			name:     "HullMA period larger than data",
			closes:   []float64{1, 2, 3},
			period:   5,
			expected: []float64{0, 0, 0},
		},
		{
			name:     "HullMA(9) trend",
			closes:   []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24},
			period:   9,
			expected: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 19.3333, 20.3333, 21.3333, 22.3333, 23.3333},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HullMA(tt.closes, tt.period)
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

func TestWMA(t *testing.T) {
	tests := []struct {
		name     string
		closes   []float64
		period   int
		expected []float64
	}{
		{
			name:     "WMA(3) of [1,2,3,4,5]",
			closes:   []float64{1, 2, 3, 4, 5},
			period:   3,
			expected: []float64{0, 0, 2.3333, 3.3333, 4.3333},
		},
		{
			name:     "WMA(2) simple",
			closes:   []float64{10, 20, 30, 40},
			period:   2,
			expected: []float64{0, 16.6667, 26.6667, 36.6667},
		},
		{
			name:     "WMA(1) identity",
			closes:   []float64{5, 10, 15},
			period:   1,
			expected: []float64{5, 10, 15},
		},
		{
			name:     "WMA empty input",
			closes:   []float64{},
			period:   3,
			expected: []float64{},
		},
		{
			name:     "WMA period <= 0",
			closes:   []float64{1, 2, 3},
			period:   0,
			expected: []float64{},
		},
		{
			name:     "WMA period larger than data",
			closes:   []float64{1, 2},
			period:   5,
			expected: []float64{0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wma(tt.closes, tt.period)
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

func TestHullMAExactValue(t *testing.T) {
	closes := []float64{1, 2, 3, 4, 5, 6, 7}
	result := HullMA(closes, 4)
	if len(result) != 7 {
		t.Fatalf("expected length 7, got %d", len(result))
	}
	// First valid HullMA at index 4 should be exactly 5
	if !floatEqual(result[4], 5.0) {
		t.Errorf("at index 4: got %f, want 5.0", result[4])
	}
	if !floatEqual(result[5], 6.0) {
		t.Errorf("at index 5: got %f, want 6.0", result[5])
	}
	if !floatEqual(result[6], 7.0) {
		t.Errorf("at index 6: got %f, want 7.0", result[6])
	}
	// Leading values should be zero
	for i := 0; i < 4; i++ {
		if result[i] != 0 {
			t.Errorf("at index %d: got %f, want 0", i, result[i])
		}
	}
}
