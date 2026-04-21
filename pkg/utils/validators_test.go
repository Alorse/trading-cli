package utils

import "testing"

func TestValidateTimeframe(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"5m", true},
		{"15m", true},
		{"1h", true},
		{"4h", true},
		{"1D", true},
		{"1W", true},
		{"1M", true},
		{"2h", false},
		{"", false},
		{"daily", false},
	}
	for _, tt := range tests {
		err := ValidateTimeframe(tt.input)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateTimeframe(%q): valid=%v, got err=%v", tt.input, tt.valid, err)
		}
	}
}

func TestValidatePeriod(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1mo", true},
		{"3mo", true},
		{"6mo", true},
		{"1y", true},
		{"2y", true},
		{"5y", false},
		{"", false},
	}
	for _, tt := range tests {
		err := ValidatePeriod(tt.input)
		if (err == nil) != tt.valid {
			t.Errorf("ValidatePeriod(%q): valid=%v, got err=%v", tt.input, tt.valid, err)
		}
	}
}

func TestValidateStrategy(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"rsi", true},
		{"bollinger", true},
		{"macd", true},
		{"ema-cross", true},
		{"supertrend", true},
		{"donchian", true},
		{"invalid", false},
	}
	for _, tt := range tests {
		err := ValidateStrategy(tt.input)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateStrategy(%q): valid=%v, got err=%v", tt.input, tt.valid, err)
		}
	}
}

func TestValidateRange(t *testing.T) {
	err := ValidateRange("limit", 5, 1, 10)
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}
	err = ValidateRange("limit", 15, 1, 10)
	if err == nil {
		t.Error("expected error for out of range")
	}
}

func TestValidateIntRange(t *testing.T) {
	err := ValidateIntRange("limit", 25, 1, 50)
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}
	err = ValidateIntRange("limit", 100, 1, 50)
	if err == nil {
		t.Error("expected error for out of range")
	}
}
