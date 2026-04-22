package indicators

import (
	"testing"
)

func TestRSI(t *testing.T) {
	tests := []struct {
		name          string
		prices        []float64
		period        int
		minExpected   float64
		maxExpected   float64
		checkExactIdx int // If >= 0, check exact value at this index
		exactValue    float64
	}{
		{
			name:          "RSI(14) with 15 closes uptrend",
			prices:        []float64{44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28},
			period:        14,
			minExpected:   50, // Expect above 50 for uptrend
			maxExpected:   100,
			checkExactIdx: 14,
			exactValue:    70.464, // Strong uptrend
		},
		{
			name:          "RSI(2) simple",
			prices:        []float64{10, 20, 30, 40},
			period:        2,
			minExpected:   0,
			maxExpected:   100,
			checkExactIdx: -1,
		},
		{
			name:          "RSI(5) with all gains",
			prices:        []float64{1, 2, 3, 4, 5, 6},
			period:        5,
			minExpected:   0,
			maxExpected:   100,
			checkExactIdx: 5,
			exactValue:    100.0, // All gains, no losses
		},
		{
			name:          "RSI(5) with all losses",
			prices:        []float64{6, 5, 4, 3, 2, 1},
			period:        5,
			minExpected:   0,
			maxExpected:   100,
			checkExactIdx: 5,
			exactValue:    0.0, // All losses, no gains
		},
		{
			name:          "RSI period larger than data",
			prices:        []float64{1, 2, 3},
			period:        5,
			minExpected:   0,
			maxExpected:   0,
			checkExactIdx: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RSI(tt.prices, tt.period)
			if len(result) != len(tt.prices) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.prices))
			}

			// Check that first (period) values are 0
			for i := 0; i < tt.period && i < len(result); i++ {
				if result[i] != 0 {
					t.Errorf("at index %d (before period): got %f, want 0", i, result[i])
				}
			}

			// Check all values are in range [0, 100]
			for i := tt.period; i < len(result); i++ {
				if result[i] < tt.minExpected || result[i] > tt.maxExpected {
					t.Errorf("at index %d: got %f, want in range [%f, %f]", i, result[i], tt.minExpected, tt.maxExpected)
				}
			}

			// Check exact value if specified
			if tt.checkExactIdx >= 0 && tt.checkExactIdx < len(result) {
				if !floatEqual(result[tt.checkExactIdx], tt.exactValue) {
					t.Errorf("at index %d: got %f, want %f (tolerance may need adjustment)", tt.checkExactIdx, result[tt.checkExactIdx], tt.exactValue)
				}
			}
		})
	}
}
