package indicators

import (
	"testing"
)

func TestStochastic(t *testing.T) {
	tests := []struct {
		name      string
		highs     []float64
		lows      []float64
		closes    []float64
		period    int
		expectedK []float64
		expectedD []float64
	}{
		{
			name:      "Stochastic(3) basic",
			highs:     []float64{10, 12, 11, 13, 14, 12},
			lows:      []float64{8, 9, 9, 10, 11, 10},
			closes:    []float64{9, 11, 10, 12, 13, 11},
			period:    3,
			expectedK: []float64{0, 0, 50, 75, 80, 25},
			expectedD: []float64{0, 0, 0, 0, 68.333333, 60},
		},
		{
			name:      "empty input",
			highs:     []float64{},
			lows:      []float64{},
			closes:    []float64{},
			period:    14,
			expectedK: []float64{},
			expectedD: []float64{},
		},
		{
			name:      "period larger than data",
			highs:     []float64{10, 11, 12},
			lows:      []float64{8, 9, 10},
			closes:    []float64{9, 10, 11},
			period:    5,
			expectedK: []float64{0, 0, 0},
			expectedD: []float64{0, 0, 0},
		},
		{
			name:      "flat market returns 50 for %K",
			highs:     []float64{10, 10, 10, 10},
			lows:      []float64{10, 10, 10, 10},
			closes:    []float64{10, 10, 10, 10},
			period:    2,
			expectedK: []float64{0, 50, 50, 50},
			expectedD: []float64{0, 0, 0, 50},
		},
		{
			name:      "single bar",
			highs:     []float64{10},
			lows:      []float64{8},
			closes:    []float64{9},
			period:    1,
			expectedK: []float64{50},
			expectedD: []float64{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Stochastic(tt.highs, tt.lows, tt.closes, tt.period)

			if len(result.K) != len(tt.expectedK) {
				t.Fatalf("K length mismatch: got %d, want %d", len(result.K), len(tt.expectedK))
			}
			if len(result.D) != len(tt.expectedD) {
				t.Fatalf("D length mismatch: got %d, want %d", len(result.D), len(tt.expectedD))
			}

			for i, val := range result.K {
				if !floatEqual(val, tt.expectedK[i]) {
					t.Errorf("K at index %d: got %f, want %f", i, val, tt.expectedK[i])
				}
			}

			for i, val := range result.D {
				if !floatEqual(val, tt.expectedD[i]) {
					t.Errorf("D at index %d: got %f, want %f", i, val, tt.expectedD[i])
				}
			}
		})
	}
}

func TestStochasticDSMA(t *testing.T) {
	// Verify that %D is exactly the SMA of %K values.
	highs := []float64{10, 12, 11, 13, 14, 12, 15}
	lows := []float64{8, 9, 9, 10, 11, 10, 11}
	closes := []float64{9, 11, 10, 12, 13, 11, 14}
	period := 3

	result := Stochastic(highs, lows, closes, period)

	// Compute SMA of K manually for valid indices.
	for i := period + 1; i < len(result.K); i++ {
		expectedD := (result.K[i-2] + result.K[i-1] + result.K[i]) / 3
		if !floatEqual(result.D[i], expectedD) {
			t.Errorf("D at index %d: got %f, want %f (SMA of K)", i, result.D[i], expectedD)
		}
	}

	// Ensure D is zero before the first valid SMA index.
	for i := 0; i <= period; i++ {
		if result.D[i] != 0 {
			t.Errorf("D at index %d: got %f, want 0 (before enough K values)", i, result.D[i])
		}
	}
}
