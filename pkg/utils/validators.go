package utils

import (
	"fmt"
	"strings"
)

var validTimeframes = map[string]bool{
	"5m": true, "15m": true, "1h": true, "4h": true,
	"1D": true, "1W": true, "1M": true,
}

var validPeriods = map[string]bool{
	"1mo": true, "3mo": true, "6mo": true, "1y": true, "2y": true,
}

var validStrategies = map[string]bool{
	"rsi": true, "bollinger": true, "macd": true,
	"ema-cross": true, "supertrend": true, "donchian": true,
}

var validIntervals = map[string]bool{
	"1d": true, "1h": true,
}

// ValidateTimeframe checks if the timeframe is supported.
func ValidateTimeframe(tf string) error {
	if !validTimeframes[tf] {
		return fmt.Errorf("invalid timeframe %q: must be one of %s", tf, formatKeys(validTimeframes))
	}
	return nil
}

// ValidatePeriod checks if the backtest period is supported.
func ValidatePeriod(p string) error {
	if !validPeriods[p] {
		return fmt.Errorf("invalid period %q: must be one of %s", p, formatKeys(validPeriods))
	}
	return nil
}

// ValidateStrategy checks if the backtest strategy is supported.
func ValidateStrategy(s string) error {
	if !validStrategies[s] {
		return fmt.Errorf("invalid strategy %q: must be one of %s", s, formatKeys(validStrategies))
	}
	return nil
}

// ValidateInterval checks if the chart interval is supported.
func ValidateInterval(i string) error {
	if !validIntervals[i] {
		return fmt.Errorf("invalid interval %q: must be one of %s", i, formatKeys(validIntervals))
	}
	return nil
}

// ValidateRange checks if a value is within [min, max].
func ValidateRange(name string, value, min, max float64) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %.1f and %.1f, got %.1f", name, min, max, value)
	}
	return nil
}

// ValidateIntRange checks if an int value is within [min, max].
func ValidateIntRange(name string, value, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d, got %d", name, min, max, value)
	}
	return nil
}

func formatKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
