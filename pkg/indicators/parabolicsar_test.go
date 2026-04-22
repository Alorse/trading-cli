package indicators

import (
	"testing"
)

func TestParabolicSAR(t *testing.T) {
	tests := []struct {
		name     string
		highs    []float64
		lows     []float64
		afStart  float64
		afStep   float64
		afMax    float64
		expected []float64
	}{
		{
			name:     "uptrend then reversal 5 bars",
			highs:    []float64{10, 11, 12, 10, 9},
			lows:     []float64{8, 9, 10, 9, 8},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{0, 8, 8, 8.16, 12},
		},
		{
			name:     "uptrend continuation 5 bars",
			highs:    []float64{10, 11, 12, 13, 14},
			lows:     []float64{8, 9, 10, 11, 12},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{0, 8, 8, 8.16, 8.4504},
		},
		{
			name:     "downtrend continuation 5 bars",
			highs:    []float64{14, 13, 12, 11, 10},
			lows:     []float64{12, 11, 10, 9, 8},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{0, 14, 14, 13.84, 13.5496},
		},
		{
			name:     "empty input",
			highs:    []float64{},
			lows:     []float64{},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{},
		},
		{
			name:     "single data point",
			highs:    []float64{10},
			lows:     []float64{8},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{},
		},
		{
			name:     "custom AF parameters",
			highs:    []float64{10, 11, 12, 13, 14},
			lows:     []float64{8, 9, 10, 11, 12},
			afStart:  0.01,
			afStep:   0.01,
			afMax:    0.10,
			expected: []float64{0, 8, 8, 8.08, 8.2276},
		},
		{
			name:     "long uptrend hits AF max",
			highs:    []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22},
			lows:     []float64{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{0, 8, 8, 8.16, 8.4504, 8.894368, 9.5049312, 10.284339456, 11.2245319322, 12.308606823, 13.5130575949, 14.8104460759, 16.0483568607},
		},
		{
			name:     "downtrend reversal 5 bars",
			highs:    []float64{10, 9, 8, 9, 10},
			lows:     []float64{8, 7, 6, 7, 8},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{0, 10, 10, 9.84, 6},
		},
		{
			name:     "long downtrend hits AF max",
			highs:    []float64{22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10},
			lows:     []float64{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8},
			afStart:  0.02,
			afStep:   0.02,
			afMax:    0.20,
			expected: []float64{0, 22, 22, 21.84, 21.5496, 21.105632, 20.4950688, 19.715660544, 18.7754680678, 17.691393177, 16.4869424051, 15.1895539241, 13.9516431393},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParabolicSAR(tt.highs, tt.lows, tt.afStart, tt.afStep, tt.afMax)

			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}

			for i := 0; i < len(result); i++ {
				if !floatEqual(result[i], tt.expected[i]) {
					t.Errorf("at index %d: got %f, want %f", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParabolicSARDefault(t *testing.T) {
	highs := []float64{10, 11, 12, 10, 9}
	lows := []float64{8, 9, 10, 9, 8}

	resultDefault := ParabolicSARDefault(highs, lows)
	resultExplicit := ParabolicSAR(highs, lows, 0.02, 0.02, 0.20)

	if len(resultDefault) != len(resultExplicit) {
		t.Fatalf("length mismatch: got %d, want %d", len(resultDefault), len(resultExplicit))
	}

	for i := 0; i < len(resultDefault); i++ {
		if !floatEqual(resultDefault[i], resultExplicit[i]) {
			t.Errorf("at index %d: got %f, want %f", i, resultDefault[i], resultExplicit[i])
		}
	}
}
