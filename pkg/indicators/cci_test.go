package indicators

import (
	"testing"
)

func TestCCI(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		closes   []float64
		period   int
		validate func(*testing.T, []float64)
	}{
		{
			name:   "CCI(3) exact values",
			highs:  []float64{10, 11, 12, 13},
			lows:   []float64{8, 9, 10, 11},
			closes: []float64{9, 10, 11, 12},
			period: 3,
			validate: func(t *testing.T, result []float64) {
				if len(result) != 4 {
					t.Fatalf("length mismatch: got %d, want 4", len(result))
				}
				// First period values are 0
				for i := 0; i < 2; i++ {
					if result[i] != 0 {
						t.Errorf("at index %d: expected 0, got %f", i, result[i])
					}
				}
				// CCI[2] = (11 - 10) / (0.015 * 2/3) = 100
				if !floatEqual(result[2], 100.0) {
					t.Errorf("at index 2: got %f, want 100.0", result[2])
				}
				// CCI[3] = (12 - 11) / (0.015 * 2/3) = 100
				if !floatEqual(result[3], 100.0) {
					t.Errorf("at index 3: got %f, want 100.0", result[3])
				}
			},
		},
		{
			name:   "CCI(3) downtrend exact values",
			highs:  []float64{13, 12, 11, 10},
			lows:   []float64{11, 10, 9, 8},
			closes: []float64{12, 11, 10, 9},
			period: 3,
			validate: func(t *testing.T, result []float64) {
				if len(result) != 4 {
					t.Fatalf("length mismatch: got %d, want 4", len(result))
				}
				// CCI[2] = (10 - 11) / (0.015 * 2/3) = -100
				if !floatEqual(result[2], -100.0) {
					t.Errorf("at index 2: got %f, want -100.0", result[2])
				}
				// CCI[3] = (9 - 10) / (0.015 * 2/3) = -100
				if !floatEqual(result[3], -100.0) {
					t.Errorf("at index 3: got %f, want -100.0", result[3])
				}
			},
		},
		{
			name:   "CCI(20) standard period",
			highs:  []float64{50, 51, 52, 51, 50, 49, 48, 49, 50, 51, 52, 51, 50, 49, 48, 49, 50, 51, 52, 51, 50},
			lows:   []float64{48, 49, 50, 49, 48, 47, 46, 47, 48, 49, 50, 49, 48, 47, 46, 47, 48, 49, 50, 49, 48},
			closes: []float64{49, 50, 51, 50, 49, 48, 47, 48, 49, 50, 51, 50, 49, 48, 47, 48, 49, 50, 51, 50, 49},
			period: 20,
			validate: func(t *testing.T, result []float64) {
				if len(result) != 21 {
					t.Fatalf("length mismatch: got %d, want 21", len(result))
				}
				// First 19 values are 0
				for i := 0; i < 19; i++ {
					if result[i] != 0 {
						t.Errorf("at index %d: expected 0, got %f", i, result[i])
					}
				}
				// Exact value manually verified:
				// TP[20]=49, SMA=49.2, meanDev=1.02
				// CCI = (49-49.2) / (0.015*1.02) = -13.071895
				if !floatEqual(result[20], -13.071895) {
					t.Errorf("at index 20: got %f, want -13.071895", result[20])
				}
			},
		},
		{
			name:   "empty input",
			highs:  []float64{},
			lows:   []float64{},
			closes: []float64{},
			period: 14,
			validate: func(t *testing.T, result []float64) {
				if len(result) != 0 {
					t.Fatalf("expected empty result, got %d", len(result))
				}
			},
		},
		{
			name:   "period <= 0",
			highs:  []float64{10, 11, 12},
			lows:   []float64{8, 9, 10},
			closes: []float64{9, 10, 11},
			period: 0,
			validate: func(t *testing.T, result []float64) {
				if len(result) != 0 {
					t.Fatalf("expected empty result, got %d", len(result))
				}
			},
		},
		{
			name:   "period larger than data",
			highs:  []float64{10, 11, 12},
			lows:   []float64{8, 9, 10},
			closes: []float64{9, 10, 11},
			period: 5,
			validate: func(t *testing.T, result []float64) {
				if len(result) != 3 {
					t.Fatalf("length mismatch: got %d, want 3", len(result))
				}
				for i := 0; i < len(result); i++ {
					if result[i] != 0 {
						t.Errorf("at index %d: expected 0, got %f", i, result[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CCI(tt.highs, tt.lows, tt.closes, tt.period)
			tt.validate(t, result)
		})
	}
}
