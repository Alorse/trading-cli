package indicators

import (
	"testing"
)

func TestATR(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		closes   []float64
		period   int
		validate func(*testing.T, []float64, []float64, []float64, []float64)
	}{
		{
			name:   "ATR(14) standard",
			highs:  []float64{50, 51, 52, 51, 50, 49, 48, 49, 50, 51, 52, 51, 50, 49},
			lows:   []float64{48, 49, 50, 49, 48, 47, 46, 47, 48, 49, 50, 49, 48, 47},
			closes: []float64{49, 50, 51, 50, 49, 48, 47, 48, 49, 50, 51, 50, 49, 48},
			period: 14,
			validate: func(t *testing.T, result, highs, lows, closes []float64) {
				// Check length
				if len(result) != len(highs) {
					t.Errorf("length mismatch: got %d, want %d", len(result), len(highs))
				}

				// Check first period values are 0
				for i := 0; i < 14 && i < len(result); i++ {
					if result[i] != 0 {
						t.Errorf("at index %d: expected 0, got %f", i, result[i])
					}
				}

				// Check ATR >= high - low for each bar
				for i := 14; i < len(result); i++ {
					highLow := highs[i] - lows[i]
					if result[i] < highLow {
						t.Errorf("at index %d: ATR=%f should be >= high-low=%f", i, result[i], highLow)
					}
				}
			},
		},
		{
			name:   "ATR(2) small period",
			highs:  []float64{100, 105, 110, 108, 103},
			lows:   []float64{95, 100, 105, 103, 98},
			closes: []float64{102, 104, 107, 106, 101},
			period: 2,
			validate: func(t *testing.T, result, highs, lows, closes []float64) {
				if len(result) != len(highs) {
					t.Errorf("length mismatch: got %d, want %d", len(result), len(highs))
				}

				// Check ATR >= high - low
				for i := 2; i < len(result); i++ {
					highLow := highs[i] - lows[i]
					if result[i] < highLow {
						t.Errorf("at index %d: ATR=%f should be >= high-low=%f", i, result[i], highLow)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ATR(tt.highs, tt.lows, tt.closes, tt.period)
			tt.validate(t, result, tt.highs, tt.lows, tt.closes)
		})
	}
}
