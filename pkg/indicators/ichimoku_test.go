package indicators

import (
	"testing"
)

func TestIchimoku(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		closes   []float64
		validate func(*testing.T, IchimokuResult, []float64, []float64, []float64)
	}{
		{
			name:   "happy path with 80 sequential bars",
			highs:  makeSeq(80),
			lows:   makeSeq(80),
			closes: makeSeq(80),
			validate: func(t *testing.T, r IchimokuResult, highs, lows, closes []float64) {
				n := len(highs)

				// All slices must match input length.
				if len(r.Tenkan) != n || len(r.Kijun) != n || len(r.SenkouA) != n ||
					len(r.SenkouB) != n || len(r.Chikou) != n {
					t.Fatalf("length mismatch: expected %d", n)
				}

				// Tenkan: first 8 values are 0, then i-3.
				for i := 0; i < 8; i++ {
					if r.Tenkan[i] != 0 {
						t.Errorf("Tenkan[%d]: got %f, want 0", i, r.Tenkan[i])
					}
				}
				for i := 8; i < n; i++ {
					want := float64(i) - 3.0
					if !floatEqual(r.Tenkan[i], want) {
						t.Errorf("Tenkan[%d]: got %f, want %f", i, r.Tenkan[i], want)
					}
				}

				// Kijun: first 25 values are 0, then i-11.5.
				for i := 0; i < 25; i++ {
					if r.Kijun[i] != 0 {
						t.Errorf("Kijun[%d]: got %f, want 0", i, r.Kijun[i])
					}
				}
				for i := 25; i < n; i++ {
					want := float64(i) - 11.5
					if !floatEqual(r.Kijun[i], want) {
						t.Errorf("Kijun[%d]: got %f, want %f", i, r.Kijun[i], want)
					}
				}

				// SenkouA: first 51 values are 0, then i-33.25.
				for i := 0; i < 51; i++ {
					if r.SenkouA[i] != 0 {
						t.Errorf("SenkouA[%d]: got %f, want 0", i, r.SenkouA[i])
					}
				}
				for i := 51; i < n; i++ {
					want := float64(i) - 33.25
					if !floatEqual(r.SenkouA[i], want) {
						t.Errorf("SenkouA[%d]: got %f, want %f", i, r.SenkouA[i], want)
					}
				}

				// SenkouB: first 77 values are 0, then i-50.5.
				for i := 0; i < 77; i++ {
					if r.SenkouB[i] != 0 {
						t.Errorf("SenkouB[%d]: got %f, want 0", i, r.SenkouB[i])
					}
				}
				for i := 77; i < n; i++ {
					want := float64(i) - 50.5
					if !floatEqual(r.SenkouB[i], want) {
						t.Errorf("SenkouB[%d]: got %f, want %f", i, r.SenkouB[i], want)
					}
				}

				// Chikou: first 54 values are i+27, last 26 are 0.
				for i := 0; i <= 53; i++ {
					want := float64(i) + 27.0
					if !floatEqual(r.Chikou[i], want) {
						t.Errorf("Chikou[%d]: got %f, want %f", i, r.Chikou[i], want)
					}
				}
				for i := 54; i < n; i++ {
					if r.Chikou[i] != 0 {
						t.Errorf("Chikou[%d]: got %f, want 0", i, r.Chikou[i])
					}
				}
			},
		},
		{
			name:   "empty input",
			highs:  []float64{},
			lows:   []float64{},
			closes: []float64{},
			validate: func(t *testing.T, r IchimokuResult, highs, lows, closes []float64) {
				if len(r.Tenkan) != 0 || len(r.Kijun) != 0 || len(r.SenkouA) != 0 ||
					len(r.SenkouB) != 0 || len(r.Chikou) != 0 {
					t.Error("expected all slices to be empty")
				}
			},
		},
		{
			name:   "small dataset verifying Tenkan and Kijun",
			highs:  []float64{10, 12, 11, 15, 14, 13, 16, 18, 17, 20},
			lows:   []float64{8, 9, 8, 10, 11, 10, 12, 14, 13, 15},
			closes: []float64{9, 11, 10, 14, 13, 12, 15, 17, 16, 18},
			validate: func(t *testing.T, r IchimokuResult, highs, lows, closes []float64) {
				n := len(highs)
				if len(r.Tenkan) != n || len(r.Kijun) != n {
					t.Fatal("length mismatch")
				}

				// Tenkan[8]: window 0..8, maxHigh=18, minLow=8, HL2=13.
				if !floatEqual(r.Tenkan[8], 13.0) {
					t.Errorf("Tenkan[8]: got %f, want 13.0", r.Tenkan[8])
				}

				// Tenkan[9]: window 1..9, maxHigh=20, minLow=8 (lows[2]), HL2=14.0.
				if !floatEqual(r.Tenkan[9], 14.0) {
					t.Errorf("Tenkan[9]: got %f, want 14.0", r.Tenkan[9])
				}

				// Kijun: period 26 > 10 bars, all zeros.
				for i := 0; i < n; i++ {
					if r.Kijun[i] != 0 {
						t.Errorf("Kijun[%d]: got %f, want 0", i, r.Kijun[i])
					}
				}
			},
		},
		{
			name:   "verify SenkouA shift logic",
			highs:  makeSeq(60),
			lows:   makeSeq(60),
			closes: makeSeq(60),
			validate: func(t *testing.T, r IchimokuResult, highs, lows, closes []float64) {
				// With 60 bars, Tenkan is valid from index 8..59,
				// Kijun from 25..59.
				// SenkouA is shifted 26 ahead, so first non-zero should be at index 51
				// (from i=25: 25+26=51).
				for i := 0; i < 51; i++ {
					if r.SenkouA[i] != 0 {
						t.Errorf("SenkouA[%d]: got %f, want 0 (before shift)", i, r.SenkouA[i])
					}
				}

				// Verify the shifted value at index 51 comes from i=25.
				// Tenkan[25] = 22, Kijun[25] = 13.5, so SenkouA[51] = 17.75.
				if !floatEqual(r.SenkouA[51], 17.75) {
					t.Errorf("SenkouA[51]: got %f, want 17.75", r.SenkouA[51])
				}

				// Verify another shifted value at index 52 from i=26.
				// Tenkan[26] = 23, Kijun[26] = 14.5, so SenkouA[52] = 18.75.
				if !floatEqual(r.SenkouA[52], 18.75) {
					t.Errorf("SenkouA[52]: got %f, want 18.75", r.SenkouA[52])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Ichimoku(tt.highs, tt.lows, tt.closes)
			tt.validate(t, result, tt.highs, tt.lows, tt.closes)
		})
	}
}

// makeSeq returns a slice of float64 values 1..n.
func makeSeq(n int) []float64 {
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = float64(i + 1)
	}
	return out
}
