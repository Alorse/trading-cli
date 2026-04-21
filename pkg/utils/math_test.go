package utils

import (
	"math"
	"testing"
)

func TestRoundTo(t *testing.T) {
	tests := []struct {
		input    float64
		decimals int
		expected float64
	}{
		{3.14159, 2, 3.14},
		{3.145, 2, 3.15},
		{100.0, 0, 100.0},
		{0.123456, 4, 0.1235},
	}
	for _, tt := range tests {
		got := RoundTo(tt.input, tt.decimals)
		if math.Abs(got-tt.expected) > 1e-9 {
			t.Errorf("RoundTo(%v, %d) = %v, want %v", tt.input, tt.decimals, got, tt.expected)
		}
	}
}

func TestBetween(t *testing.T) {
	if !Between(5, 1, 10) {
		t.Error("5 should be between 1 and 10")
	}
	if Between(0, 1, 10) {
		t.Error("0 should not be between 1 and 10")
	}
	if !Between(1, 1, 10) {
		t.Error("1 should be between 1 and 10 (inclusive)")
	}
}

func TestClamp(t *testing.T) {
	if Clamp(5, 1, 10) != 5 {
		t.Error("5 should stay 5")
	}
	if Clamp(-1, 0, 10) != 0 {
		t.Error("-1 should clamp to 0")
	}
	if Clamp(15, 0, 10) != 10 {
		t.Error("15 should clamp to 10")
	}
}

func TestSafeDivide(t *testing.T) {
	if SafeDivide(10, 2) != 5 {
		t.Error("10/2 should be 5")
	}
	if SafeDivide(10, 0) != 0 {
		t.Error("10/0 should return 0")
	}
}

func TestMax(t *testing.T) {
	if Max(3, 5) != 5 {
		t.Error("Max(3,5) should be 5")
	}
	if Max(5, 3) != 5 {
		t.Error("Max(5,3) should be 5")
	}
}

func TestMin(t *testing.T) {
	if Min(3, 5) != 3 {
		t.Error("Min(3,5) should be 3")
	}
}

func TestAbs(t *testing.T) {
	if Abs(-5) != 5 {
		t.Error("Abs(-5) should be 5")
	}
	if Abs(5) != 5 {
		t.Error("Abs(5) should be 5")
	}
}

func TestPercentChange(t *testing.T) {
	got := PercentChange(100, 110)
	if got != 10 {
		t.Errorf("PercentChange(100, 110) = %v, want 10", got)
	}
	if PercentChange(100, 0) != -100 {
		t.Error("PercentChange(100, 0) should be -100")
	}
	if PercentChange(0, 100) != 0 {
		t.Error("PercentChange(0, 100) should be 0 (safe divide)")
	}
}
