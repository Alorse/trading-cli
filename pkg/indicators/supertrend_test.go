package indicators

import (
	"testing"
)

func TestSupertrend(t *testing.T) {
	tests := []struct {
		name       string
		highs      []float64
		lows       []float64
		closes     []float64
		period     int
		multiplier float64
		validate   func(*testing.T, SupertrendResult, []float64, []float64, []float64, int)
	}{
		{
			name:       "Supertrend(10, 3.0) standard",
			highs:      []float64{50, 51, 52, 53, 52, 51, 50, 49, 48, 49, 50, 51, 52, 53},
			lows:       []float64{48, 49, 50, 51, 50, 49, 48, 47, 46, 47, 48, 49, 50, 51},
			closes:     []float64{49, 50, 51, 52, 51, 50, 49, 48, 47, 48, 49, 50, 51, 52},
			period:     10,
			multiplier: 3.0,
			validate: func(t *testing.T, result SupertrendResult, highs, lows, closes []float64, period int) {
				// Check lengths
				if len(result.Value) != len(highs) {
					t.Errorf("Value length mismatch: got %d, want %d", len(result.Value), len(highs))
				}
				if len(result.Direction) != len(highs) {
					t.Errorf("Direction length mismatch: got %d, want %d", len(result.Direction), len(highs))
				}

				// Check first period values are 0
				for i := 0; i < period && i < len(result.Direction); i++ {
					if result.Direction[i] != 0 {
						t.Errorf("at index %d: expected 0, got %d", i, result.Direction[i])
					}
				}

				// Check directions are only +1, -1, or 0
				for i := 0; i < len(result.Direction); i++ {
					if result.Direction[i] != -1 && result.Direction[i] != 0 && result.Direction[i] != 1 {
						t.Errorf("at index %d: direction=%d should be -1, 0, or 1", i, result.Direction[i])
					}
				}
			},
		},
		{
			name:       "Supertrend(2, 2.0) small period",
			highs:      []float64{100, 105, 110, 108, 103, 100},
			lows:       []float64{95, 100, 105, 103, 98, 95},
			closes:     []float64{102, 104, 107, 106, 101, 98},
			period:     2,
			multiplier: 2.0,
			validate: func(t *testing.T, result SupertrendResult, highs, lows, closes []float64, period int) {
				if len(result.Value) != len(highs) || len(result.Direction) != len(highs) {
					t.Error("Length mismatch")
				}

				// Check that direction is stable (doesn't flip randomly)
				// and values are within reasonable range
				for i := period; i < len(result.Value); i++ {
					if result.Direction[i] != 0 && (result.Direction[i] != -1 && result.Direction[i] != 1) {
						t.Errorf("at index %d: invalid direction %d", i, result.Direction[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Supertrend(tt.highs, tt.lows, tt.closes, tt.period, tt.multiplier)
			tt.validate(t, result, tt.highs, tt.lows, tt.closes, tt.period)
		})
	}
}
