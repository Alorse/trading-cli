package indicators

import (
	"testing"
)

func TestMACD(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		fast     int
		slow     int
		signal   int
		validate func(*testing.T, MACDResult, []float64, int, int, int)
	}{
		{
			name:   "MACD with standard parameters",
			prices: []float64{44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.00, 46.00, 46.00},
			fast:   12,
			slow:   26,
			signal: 9,
			validate: func(t *testing.T, result MACDResult, prices []float64, fast, slow, signal int) {
				// Check all slices have same length as prices
				if len(result.MACD) != len(prices) {
					t.Errorf("MACD length mismatch: got %d, want %d", len(result.MACD), len(prices))
				}
				if len(result.Signal) != len(prices) {
					t.Errorf("Signal length mismatch: got %d, want %d", len(result.Signal), len(prices))
				}
				if len(result.Histogram) != len(prices) {
					t.Errorf("Histogram length mismatch: got %d, want %d", len(result.Histogram), len(prices))
				}

				// Check that leading values are 0 (up to slow-1 for MACD)
				leadingZeros := slow - 1
				if leadingZeros > len(prices) {
					leadingZeros = len(prices)
				}
				for i := 0; i < leadingZeros; i++ {
					if result.MACD[i] != 0 {
						t.Errorf("MACD[%d] should be 0, got %f", i, result.MACD[i])
					}
				}

				// Check histogram = MACD - Signal (for all indices)
				for i := 0; i < len(result.Histogram); i++ {
					expected := result.MACD[i] - result.Signal[i]
					if !floatEqual(result.Histogram[i], expected) {
						t.Errorf("Histogram[%d]: got %f, want %f (MACD=%f, Signal=%f)", i, result.Histogram[i], expected, result.MACD[i], result.Signal[i])
					}
				}
			},
		},
		{
			name:   "MACD with small dataset",
			prices: []float64{1, 2, 3, 4, 5},
			fast:   2,
			slow:   3,
			signal: 2,
			validate: func(t *testing.T, result MACDResult, prices []float64, fast, slow, signal int) {
				if len(result.MACD) != len(prices) {
					t.Errorf("Length mismatch: got %d, want %d", len(result.MACD), len(prices))
				}
				if len(result.Signal) != len(prices) {
					t.Errorf("Length mismatch: got %d, want %d", len(result.Signal), len(prices))
				}
				if len(result.Histogram) != len(prices) {
					t.Errorf("Length mismatch: got %d, want %d", len(result.Histogram), len(prices))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MACD(tt.prices, tt.fast, tt.slow, tt.signal)
			tt.validate(t, result, tt.prices, tt.fast, tt.slow, tt.signal)
		})
	}
}
