package mtf

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		total          int
		n              int
		wantAlignment  string
		wantConfidence string
		wantRec        string
	}{
		{5, 5, "FULLY ALIGNED BULLISH", "Very High", "STRONG BUY"},
		{-5, 5, "FULLY ALIGNED BEARISH", "Very High", "STRONG SELL"},
		{3, 5, "MOSTLY BULLISH", "High", "BUY"},
		{-3, 5, "MOSTLY BEARISH", "High", "SELL"},
		{1, 5, "LEAN BULLISH", "Medium", "CAUTIOUS BUY"},
		{-1, 5, "LEAN BEARISH", "Medium", "CAUTIOUS SELL"},
		{0, 5, "MIXED/RANGING", "Low", "HOLD/NO TRADE"},
	}

	for _, tc := range cases {
		a, c, r := classify(tc.total, tc.n)
		if a != tc.wantAlignment {
			t.Errorf("total=%d: alignment=%q want %q", tc.total, a, tc.wantAlignment)
		}
		if c != tc.wantConfidence {
			t.Errorf("total=%d: confidence=%q want %q", tc.total, c, tc.wantConfidence)
		}
		if r != tc.wantRec {
			t.Errorf("total=%d: recommendation=%q want %q", tc.total, r, tc.wantRec)
		}
	}
}
