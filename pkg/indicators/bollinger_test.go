package indicators

import (
	"testing"
)

func TestBollingerBands(t *testing.T) {
	tests := []struct {
		name       string
		prices     []float64
		period     int
		stdDevMult float64
		validate   func(*testing.T, BBResult, []float64, int)
	}{
		{
			name:       "Bollinger(20, 2.0) standard",
			prices:     []float64{44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.00, 46.00, 46.00, 46.00, 46.00, 46.00, 46.00},
			period:     20,
			stdDevMult: 2.0,
			validate: func(t *testing.T, result BBResult, prices []float64, period int) {
				// Check lengths
				if len(result.Upper) != len(prices) {
					t.Errorf("Upper length mismatch: got %d, want %d", len(result.Upper), len(prices))
				}
				if len(result.Middle) != len(prices) {
					t.Errorf("Middle length mismatch: got %d, want %d", len(result.Middle), len(prices))
				}
				if len(result.Lower) != len(prices) {
					t.Errorf("Lower length mismatch: got %d, want %d", len(result.Lower), len(prices))
				}
				if len(result.Width) != len(prices) {
					t.Errorf("Width length mismatch: got %d, want %d", len(result.Width), len(prices))
				}

				// Check leading zeros
				for i := 0; i < period-1; i++ {
					if result.Upper[i] != 0 || result.Middle[i] != 0 || result.Lower[i] != 0 || result.Width[i] != 0 {
						t.Errorf("at index %d (before period): expected all zeros, got U=%f M=%f L=%f W=%f",
							i, result.Upper[i], result.Middle[i], result.Lower[i], result.Width[i])
					}
				}

				// Check that Upper > Lower
				for i := period - 1; i < len(prices); i++ {
					if result.Upper[i] < result.Lower[i] {
						t.Errorf("at index %d: Upper=%f should be >= Lower=%f", i, result.Upper[i], result.Lower[i])
					}
				}

				// Check that Width > 0 for non-constant data
				for i := period - 1; i < len(prices); i++ {
					if result.Width[i] < 0 {
						t.Errorf("at index %d: Width=%f should be >= 0", i, result.Width[i])
					}
				}
			},
		},
		{
			name:       "Bollinger(3, 2.0) small period",
			prices:     []float64{1, 2, 3, 4, 5, 6, 7},
			period:     3,
			stdDevMult: 2.0,
			validate: func(t *testing.T, result BBResult, prices []float64, period int) {
				if len(result.Upper) != len(prices) || len(result.Middle) != len(prices) ||
					len(result.Lower) != len(prices) || len(result.Width) != len(prices) {
					t.Error("Length mismatch")
				}

				// For non-constant input, width should be > 0
				for i := period - 1; i < len(prices); i++ {
					if result.Width[i] <= 0 {
						t.Errorf("at index %d: Width=%f should be > 0 for non-constant data", i, result.Width[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BollingerBands(tt.prices, tt.period, tt.stdDevMult)
			tt.validate(t, result, tt.prices, tt.period)
		})
	}
}
