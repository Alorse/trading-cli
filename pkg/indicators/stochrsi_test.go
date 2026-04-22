package indicators

import (
	"testing"
)

func TestStochRSI(t *testing.T) {
	tests := []struct {
		name      string
		closes    []float64
		rsiP      int
		stochP    int
		smoothK   int
		smoothD   int
		expectedK []float64
		expectedD []float64
	}{
		{
			name:      "StochRSI(2,3,2,2) basic oscillation",
			closes:    []float64{10, 11, 10, 11, 10, 11, 10, 11, 10, 11, 10, 11, 10, 11},
			rsiP:      2,
			stochP:    3,
			smoothK:   2,
			smoothD:   2,
			expectedK: []float64{0, 0, 0, 0, 0, 41.66666666666667, 41.66666666666667, 47.72727272727273, 47.72727272727273, 49.41860465116279, 49.41860465116279, 49.853801169590646, 49.853801169590646, 49.96339677891655},
			expectedD: []float64{0, 0, 0, 0, 0, 0, 41.66666666666667, 44.6969696969697, 47.72727272727273, 48.57293868921776, 49.41860465116279, 49.63620291037672, 49.853801169590646, 49.90859897425359},
		},
		{
			name:      "empty input",
			closes:    []float64{},
			rsiP:      14,
			stochP:    14,
			smoothK:   3,
			smoothD:   3,
			expectedK: []float64{},
			expectedD: []float64{},
		},
		{
			name:      "period larger than data",
			closes:    []float64{10, 11, 12},
			rsiP:      5,
			stochP:    5,
			smoothK:   3,
			smoothD:   3,
			expectedK: []float64{0, 0, 0},
			expectedD: []float64{0, 0, 0},
		},
		{
			name:      "flat market returns 50 for raw StochRSI",
			closes:    []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
			rsiP:      2,
			stochP:    3,
			smoothK:   2,
			smoothD:   2,
			expectedK: []float64{0, 0, 0, 0, 0, 50, 50, 50, 50, 50},
			expectedD: []float64{0, 0, 0, 0, 0, 0, 50, 50, 50, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StochRSI(tt.closes, tt.rsiP, tt.stochP, tt.smoothK, tt.smoothD)

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

func TestStochRSIDefault(t *testing.T) {
	closes := []float64{
		44.34, 44.09, 44.15, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08,
		45.89, 46.03, 45.61, 46.28, 46.28, 46.80, 47.20, 47.00, 46.50, 46.80,
		47.50, 48.00, 47.80, 48.20, 48.50, 48.30, 48.00, 47.50, 47.20, 47.00,
		46.80, 47.20, 47.50, 48.00, 48.50, 49.00, 48.50, 48.00, 47.50, 47.00,
	}
	result := StochRSIDefault(closes)

	wantK := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5.877986522942284, 15.830341078923354, 31.932896193086265, 47.51527618670906, 63.7204556935291, 63.94047589097877, 50.089923511300555, 23.932389448499446, 7.609814136886867}
	wantD := []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1.9593288409807614, 7.236109200621879, 17.880407931650634, 31.75950448623956, 47.72287602444147, 58.39206925707231, 59.25028503193614, 45.98759628359292, 27.21070903222896}

	if len(result.K) != len(wantK) {
		t.Fatalf("K length mismatch: got %d, want %d", len(result.K), len(wantK))
	}
	if len(result.D) != len(wantD) {
		t.Fatalf("D length mismatch: got %d, want %d", len(result.D), len(wantD))
	}

	for i, val := range result.K {
		if !floatEqual(val, wantK[i]) {
			t.Errorf("K at index %d: got %f, want %f", i, val, wantK[i])
		}
	}

	for i, val := range result.D {
		if !floatEqual(val, wantD[i]) {
			t.Errorf("D at index %d: got %f, want %f", i, val, wantD[i])
		}
	}
}

func TestStochRSIDSMA(t *testing.T) {
	// Verify that %D is exactly the SMA of %K values.
	closes := []float64{10, 11, 10, 11, 10, 11, 10, 11, 10, 11, 10, 11, 10, 11}
	result := StochRSI(closes, 2, 3, 2, 2)

	firstDIdx := 2 + 3 + 2 + 2 - 3 // rsiPeriod + stochPeriod + smoothK + smoothD - 3
	for i := firstDIdx; i < len(result.K); i++ {
		expectedD := (result.K[i-1] + result.K[i]) / 2
		if !floatEqual(result.D[i], expectedD) {
			t.Errorf("D at index %d: got %f, want %f (SMA of K)", i, result.D[i], expectedD)
		}
	}

	// Ensure D is zero before the first valid SMA index.
	for i := 0; i < firstDIdx; i++ {
		if result.D[i] != 0 {
			t.Errorf("D at index %d: got %f, want 0 (before enough K values)", i, result.D[i])
		}
	}
}

func TestStochRSIFlatRSI(t *testing.T) {
	// When all closes are the same, RSI is 0 everywhere.
	// All RSI values in any stochPeriod window are the same (0).
	// StochRSI should return 50 for those positions.
	closes := []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	result := StochRSI(closes, 2, 3, 2, 2)

	// First valid K is at index 2+3+2-2 = 5
	firstKIdx := 2 + 3 + 2 - 2
	for i := firstKIdx; i < len(result.K); i++ {
		if !floatEqual(result.K[i], 50) {
			t.Errorf("K at index %d: got %f, want 50 (flat RSI)", i, result.K[i])
		}
	}
}
