package indicators

import (
	"testing"
)

func TestADX(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		closes   []float64
		period   int
		validate func(*testing.T, ADXResult)
	}{
		{
			name:   "ADX(3) with exact values",
			highs:  []float64{10, 12, 11, 13, 14, 12, 13, 15, 14, 16},
			lows:   []float64{8, 9, 9, 10, 11, 10, 11, 12, 12, 13},
			closes: []float64{9, 11, 10, 12, 13, 11, 12, 14, 13, 15},
			period: 3,
			validate: func(t *testing.T, r ADXResult) {
				if len(r.ADX) != 10 {
					t.Fatalf("ADX length mismatch: got %d, want 10", len(r.ADX))
				}
				if len(r.PlusDI) != 10 {
					t.Fatalf("PlusDI length mismatch: got %d, want 10", len(r.PlusDI))
				}
				if len(r.MinusDI) != 10 {
					t.Fatalf("MinusDI length mismatch: got %d, want 10", len(r.MinusDI))
				}
				if len(r.DX) != 10 {
					t.Fatalf("DX length mismatch: got %d, want 10", len(r.DX))
				}

				// Leading zeros before index period
				for i := 0; i < 3; i++ {
					if r.PlusDI[i] != 0 {
						t.Errorf("PlusDI at %d: got %f, want 0", i, r.PlusDI[i])
					}
					if r.MinusDI[i] != 0 {
						t.Errorf("MinusDI at %d: got %f, want 0", i, r.MinusDI[i])
					}
					if r.DX[i] != 0 {
						t.Errorf("DX at %d: got %f, want 0", i, r.DX[i])
					}
					if r.ADX[i] != 0 {
						t.Errorf("ADX at %d: got %f, want 0", i, r.ADX[i])
					}
				}

				// ADX leading zeros before index 2*period-1 = 5
				for i := 3; i < 5; i++ {
					if r.ADX[i] != 0 {
						t.Errorf("ADX at %d: got %f, want 0", i, r.ADX[i])
					}
				}

				// Exact values at known indices
				expectedPlusDI := []float64{0, 0, 0, 57.142857, 47.826087, 30.136986, 35.500000, 47.278383, 34.311512, 46.658524}
				expectedMinusDI := []float64{0, 0, 0, 0, 0, 12.328767, 9.000000, 5.598756, 4.063205, 2.512650}
				expectedDX := []float64{0, 0, 0, 100.000000, 100.000000, 41.935484, 59.550562, 78.823529, 78.823529, 89.779986}
				expectedADX := []float64{0, 0, 0, 0, 0, 80.645161, 73.613628, 75.350262, 76.508018, 80.932007}

				for i := 0; i < 10; i++ {
					if !floatEqual(r.PlusDI[i], expectedPlusDI[i]) {
						t.Errorf("PlusDI at %d: got %f, want %f", i, r.PlusDI[i], expectedPlusDI[i])
					}
					if !floatEqual(r.MinusDI[i], expectedMinusDI[i]) {
						t.Errorf("MinusDI at %d: got %f, want %f", i, r.MinusDI[i], expectedMinusDI[i])
					}
					if !floatEqual(r.DX[i], expectedDX[i]) {
						t.Errorf("DX at %d: got %f, want %f", i, r.DX[i], expectedDX[i])
					}
					if !floatEqual(r.ADX[i], expectedADX[i]) {
						t.Errorf("ADX at %d: got %f, want %f", i, r.ADX[i], expectedADX[i])
					}
				}
			},
		},
		{
			name:   "empty inputs",
			highs:  []float64{},
			lows:   []float64{},
			closes: []float64{},
			period: 14,
			validate: func(t *testing.T, r ADXResult) {
				if len(r.ADX) != 0 || len(r.PlusDI) != 0 || len(r.MinusDI) != 0 || len(r.DX) != 0 {
					t.Errorf("expected all empty slices for empty input")
				}
			},
		},
		{
			name:   "period larger than data",
			highs:  []float64{10, 11, 12},
			lows:   []float64{8, 9, 10},
			closes: []float64{9, 10, 11},
			period: 5,
			validate: func(t *testing.T, r ADXResult) {
				if len(r.ADX) != 3 {
					t.Fatalf("ADX length mismatch: got %d, want 3", len(r.ADX))
				}
				for i := 0; i < 3; i++ {
					if r.ADX[i] != 0 || r.PlusDI[i] != 0 || r.MinusDI[i] != 0 || r.DX[i] != 0 {
						t.Errorf("expected all zeros at index %d when period > data", i)
					}
				}
			},
		},
		{
			name:   "flat market all DMs zero",
			highs:  []float64{10, 10, 10, 10, 10, 10, 10, 10},
			lows:   []float64{10, 10, 10, 10, 10, 10, 10, 10},
			closes: []float64{10, 10, 10, 10, 10, 10, 10, 10},
			period: 3,
			validate: func(t *testing.T, r ADXResult) {
				if len(r.ADX) != 8 {
					t.Fatalf("ADX length mismatch: got %d, want 8", len(r.ADX))
				}
				for i := 0; i < 8; i++ {
					if r.ADX[i] != 0 || r.PlusDI[i] != 0 || r.MinusDI[i] != 0 || r.DX[i] != 0 {
						t.Errorf("expected all zeros at index %d for flat market", i)
					}
				}
			},
		},
		{
			name:   "ADX values between 0 and 100",
			highs:  []float64{10, 12, 11, 13, 14, 12, 13, 15, 14, 16},
			lows:   []float64{8, 9, 9, 10, 11, 10, 11, 12, 12, 13},
			closes: []float64{9, 11, 10, 12, 13, 11, 12, 14, 13, 15},
			period: 3,
			validate: func(t *testing.T, r ADXResult) {
				for i := 0; i < len(r.ADX); i++ {
					if r.ADX[i] < 0 || r.ADX[i] > 100 {
						t.Errorf("ADX at %d: %f out of range [0, 100]", i, r.ADX[i])
					}
				}
				for i := 0; i < len(r.PlusDI); i++ {
					if r.PlusDI[i] < 0 || r.PlusDI[i] > 100 {
						t.Errorf("PlusDI at %d: %f out of range [0, 100]", i, r.PlusDI[i])
					}
				}
				for i := 0; i < len(r.MinusDI); i++ {
					if r.MinusDI[i] < 0 || r.MinusDI[i] > 100 {
						t.Errorf("MinusDI at %d: %f out of range [0, 100]", i, r.MinusDI[i])
					}
				}
				for i := 0; i < len(r.DX); i++ {
					if r.DX[i] < 0 || r.DX[i] > 100 {
						t.Errorf("DX at %d: %f out of range [0, 100]", i, r.DX[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ADX(tt.highs, tt.lows, tt.closes, tt.period)
			tt.validate(t, result)
		})
	}
}
