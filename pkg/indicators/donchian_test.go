package indicators

import (
	"testing"
)

func TestDonchianChannel(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		period   int
		validate func(*testing.T, DonchianResult, []float64, []float64, int)
	}{
		{
			name:   "Donchian(20) standard",
			highs:  []float64{50, 51, 52, 53, 52, 51, 50, 49, 48, 49, 50, 51, 52, 53, 54, 53, 52, 51, 50, 49, 48},
			lows:   []float64{45, 46, 47, 48, 47, 46, 45, 44, 43, 44, 45, 46, 47, 48, 49, 48, 47, 46, 45, 44, 43},
			period: 20,
			validate: func(t *testing.T, result DonchianResult, highs, lows []float64, period int) {
				// Check lengths
				if len(result.Upper) != len(highs) {
					t.Errorf("Upper length mismatch: got %d, want %d", len(result.Upper), len(highs))
				}
				if len(result.Lower) != len(highs) {
					t.Errorf("Lower length mismatch: got %d, want %d", len(result.Lower), len(highs))
				}
				if len(result.Middle) != len(highs) {
					t.Errorf("Middle length mismatch: got %d, want %d", len(result.Middle), len(highs))
				}

				// Check leading zeros
				for i := 0; i < period-1; i++ {
					if result.Upper[i] != 0 || result.Lower[i] != 0 || result.Middle[i] != 0 {
						t.Errorf("at index %d (before period): expected zeros, got U=%f L=%f M=%f",
							i, result.Upper[i], result.Lower[i], result.Middle[i])
					}
				}

				// Check Upper >= Lower always
				for i := period - 1; i < len(result.Upper); i++ {
					if result.Upper[i] < result.Lower[i] {
						t.Errorf("at index %d: Upper=%f should be >= Lower=%f", i, result.Upper[i], result.Lower[i])
					}
				}

				// Check Middle = (Upper + Lower) / 2
				for i := period - 1; i < len(result.Upper); i++ {
					expectedMiddle := (result.Upper[i] + result.Lower[i]) / 2.0
					if !floatEqual(result.Middle[i], expectedMiddle) {
						t.Errorf("at index %d: Middle=%f should be %f", i, result.Middle[i], expectedMiddle)
					}
				}
			},
		},
		{
			name:   "Donchian(5) small period",
			highs:  []float64{10, 15, 12, 14, 13, 16, 15, 14},
			lows:   []float64{8, 12, 10, 11, 10, 13, 12, 11},
			period: 5,
			validate: func(t *testing.T, result DonchianResult, highs, lows []float64, period int) {
				if len(result.Upper) != len(highs) || len(result.Lower) != len(highs) || len(result.Middle) != len(highs) {
					t.Error("Length mismatch")
				}

				// Check Upper >= Lower
				for i := period - 1; i < len(result.Upper); i++ {
					if result.Upper[i] < result.Lower[i] {
						t.Errorf("at index %d: Upper should be >= Lower", i)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DonchianChannel(tt.highs, tt.lows, tt.period)
			tt.validate(t, result, tt.highs, tt.lows, tt.period)
		})
	}
}
