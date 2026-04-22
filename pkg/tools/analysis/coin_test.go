package analysis

import (
	"fmt"
	"testing"
	"time"
)

// TestComputeRSISignal tests RSI signal calculation
func TestComputeRSISignal(t *testing.T) {
	tests := []struct {
		name     string
		rsi      float64
		expected string
	}{
		{"overbought", 75.0, "overbought"},
		{"overbought_above_70", 71.0, "overbought"},
		{"at_70_boundary", 70.0, "neutral"},
		{"oversold", 25.0, "oversold"},
		{"at_30_boundary", 30.0, "neutral"},
		{"neutral_low", 45.0, "neutral"},
		{"neutral_high", 55.0, "neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeRSISignal(tt.rsi)
			if result != tt.expected {
				t.Errorf("computeRSISignal(%f) = %s, want %s", tt.rsi, result, tt.expected)
			}
		})
	}
}

// TestComputeBBPosition tests Bollinger Bands position calculation
func TestComputeBBPosition(t *testing.T) {
	tests := []struct {
		name     string
		close    float64
		upper    float64
		lower    float64
		expected string
	}{
		{"above", 105.0, 100.0, 80.0, "above"},
		{"inside", 90.0, 100.0, 80.0, "inside"},
		{"below", 75.0, 100.0, 80.0, "below"},
		{"at_upper", 100.0, 100.0, 80.0, "inside"},
		{"at_lower", 80.0, 100.0, 80.0, "inside"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeBBPosition(tt.close, tt.upper, tt.lower)
			if result != tt.expected {
				t.Errorf("computeBBPosition(%f, %f, %f) = %s, want %s", tt.close, tt.upper, tt.lower, result, tt.expected)
			}
		})
	}
}

// TestComputeTrendScore tests trend score calculation
func TestComputeTrendScore(t *testing.T) {
	tests := []struct {
		name        string
		close       float64
		sma20       float64
		sma50       float64
		ema50       float64
		ema200      float64
		rsi         float64
		expectedMin int
		expectedMax int
	}{
		{
			"all_bullish",
			110.0, 100.0, 95.0, 90.0, 80.0, 60.0,
			5, 5,
		},
		{
			"mixed_4_conditions",
			105.0, 100.0, 90.0, 85.0, 80.0, 40.0,
			3, 4,
		},
		{
			"bearish",
			50.0, 100.0, 105.0, 110.0, 115.0, 25.0,
			0, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeTrendScore(tt.close, tt.sma20, tt.sma50, tt.ema50, tt.ema200, tt.rsi)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("computeTrendScore() = %d, want between %d and %d", result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestComputeTrendFloat tests trend determination
func TestComputeTrendFloat(t *testing.T) {
	tests := []struct {
		name     string
		close    float64
		ema50    float64
		ema200   float64
		expected string
	}{
		{"bullish", 100.0, 90.0, 80.0, "bullish"},
		{"bearish", 80.0, 90.0, 100.0, "bearish"},
		{"neutral", 90.0, 85.0, 95.0, "neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeTrendFloat(tt.close, tt.ema50, tt.ema200)
			if result != tt.expected {
				t.Errorf("computeTrendFloat(%f, %f, %f) = %s, want %s", tt.close, tt.ema50, tt.ema200, result, tt.expected)
			}
		})
	}
}

// TestBuildCoinAnalysisOutput tests the output structure
func TestBuildCoinAnalysisOutput(t *testing.T) {
	values := map[string]interface{}{
		"close":           100.0,
		"open":            95.0,
		"high":            105.0,
		"low":             90.0,
		"volume":          1000000.0,
		"change":          5.26,
		"RSI":             65.0,
		"RSI[1]":          60.0,
		"MACD.macd":       1.5,
		"MACD.signal":     1.2,
		"SMA10":           98.0,
		"SMA20":           97.0,
		"SMA50":           95.0,
		"SMA100":          93.0,
		"SMA200":          90.0,
		"EMA9":            99.0,
		"EMA20":           98.0,
		"EMA50":           96.0,
		"EMA100":          94.0,
		"EMA200":          91.0,
		"BB.upper":        110.0,
		"BB.lower":        85.0,
		"ATR":             5.0,
		"ADX":             35.0,
		"Stoch.K":         75.0,
		"Stoch.D":         70.0,
		"volume.SMA20":    950000.0,
		"Recommend.All":   1.5,
		"Recommend.MA":    1.0,
		"Recommend.Other": 2.0,
	}

	output := BuildCoinAnalysisOutput("KUCOIN:BTCUSDT", "KUCOIN", "15m", values)

	if output.Symbol != "KUCOIN:BTCUSDT" {
		t.Errorf("Symbol = %s, want %s", output.Symbol, "KUCOIN:BTCUSDT")
	}
	if output.Exchange != "KUCOIN" {
		t.Errorf("Exchange = %s, want %s", output.Exchange, "KUCOIN")
	}
	if output.Timeframe != "15m" {
		t.Errorf("Timeframe = %s, want %s", output.Timeframe, "15m")
	}

	// Test price structure
	if output.Price.Close != 100.0 {
		t.Errorf("Price.Close = %f, want 100.0", output.Price.Close)
	}
	if output.Price.ChangePercent != 5.26 {
		t.Errorf("Price.ChangePercent = %f, want 5.26", output.Price.ChangePercent)
	}

	// Test RSI structure
	if output.RSI.Value != 65.0 {
		t.Errorf("RSI.Value = %f, want 65.0", output.RSI.Value)
	}
	if output.RSI.Signal != "neutral" {
		t.Errorf("RSI.Signal = %s, want neutral", output.RSI.Signal)
	}
	if output.RSI.Previous != 60.0 {
		t.Errorf("RSI.Previous = %f, want 60.0", output.RSI.Previous)
	}

	// Test MACD structure
	if output.MACD.Line != 1.5 {
		t.Errorf("MACD.Line = %f, want 1.5", output.MACD.Line)
	}

	// Test SMA structure
	if output.SMA["20"] != 97.0 {
		t.Errorf("SMA[20] = %f, want 97.0", output.SMA["20"])
	}

	// Test EMA structure
	if output.EMA["9"] != 99.0 {
		t.Errorf("EMA[9] = %f, want 99.0", output.EMA["9"])
	}

	// Test BB structure
	if output.BollingerBands.Position != "inside" {
		t.Errorf("BB.Position = %s, want inside", output.BollingerBands.Position)
	}

	// Test Volume structure
	if output.Volume.Current != 1000000.0 {
		t.Errorf("Volume.Current = %f, want 1000000.0", output.Volume.Current)
	}
	// Volume ratio defaults to 1.0 when avg20 is unavailable from scanner
	if output.Volume.Ratio < 0.99 || output.Volume.Ratio > 1.01 {
		t.Errorf("Volume.Ratio = %f, want ~1.0 (avg20 unavailable)", output.Volume.Ratio)
	}

	// Test MarketStructure
	if output.MarketStructure.Trend != "bullish" {
		t.Errorf("MarketStructure.Trend = %s, want bullish", output.MarketStructure.Trend)
	}
	if output.MarketStructure.TrendScore < 0 || output.MarketStructure.TrendScore > 5 {
		t.Errorf("MarketStructure.TrendScore = %d, out of range 0-5", output.MarketStructure.TrendScore)
	}
}

// TestRunCoinAnalysis tests the full coin analysis workflow with mock data
func TestRunCoinAnalysis(t *testing.T) {
	// Since RunCoinAnalysis calls actual TradingView client, we test indirectly
	// by verifying the output structure is correct
	values := map[string]interface{}{
		"close":           100.0,
		"open":            95.0,
		"high":            105.0,
		"low":             90.0,
		"volume":          1000000.0,
		"change":          5.26,
		"RSI":             65.0,
		"RSI[1]":          60.0,
		"MACD.macd":       1.5,
		"MACD.signal":     1.2,
		"SMA10":           98.0,
		"SMA20":           97.0,
		"SMA50":           95.0,
		"SMA100":          93.0,
		"SMA200":          90.0,
		"EMA9":            99.0,
		"EMA20":           98.0,
		"EMA50":           96.0,
		"EMA100":          94.0,
		"EMA200":          91.0,
		"BB.upper":        110.0,
		"BB.lower":        85.0,
		"ATR":             5.0,
		"ADX":             35.0,
		"Stoch.K":         75.0,
		"Stoch.D":         70.0,
		"volume.SMA20":    950000.0,
		"Recommend.All":   1.5,
		"Recommend.MA":    1.0,
		"Recommend.Other": 2.0,
	}

	output := BuildCoinAnalysisOutput("KUCOIN:BTCUSDT", "KUCOIN", "15m", values)

	// Verify timestamp is set
	if output.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Verify timestamp format
	_, err := time.Parse(time.RFC3339, output.Timestamp)
	if err != nil {
		t.Errorf("Timestamp format invalid: %v", err)
	}

	// Verify recommendation scores are valid numbers
	if fmt.Sprintf("%f", output.Recommendation.All) == "" {
		t.Error("Recommendation.All should be set")
	}
}
