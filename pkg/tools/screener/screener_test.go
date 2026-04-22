package screener

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alorse/trading-cli/pkg/client"
)

// TestLoadSymbols tests loading symbols from a temporary file
func TestLoadSymbols(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a subdirectory for symbols
	symbolsDir := filepath.Join(tmpDir, "data", "symbols")
	if err := os.MkdirAll(symbolsDir, 0755); err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	// Change working directory to temp dir
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create a test symbols file
	testFile := filepath.Join(symbolsDir, "test.txt")
	content := `# Comment line
SYMBOL1
SYMBOL2

# Another comment
SYMBOL3
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test loading symbols
	symbols, err := LoadSymbols("test")
	if err != nil {
		t.Fatalf("LoadSymbols failed: %v", err)
	}

	expected := []string{"SYMBOL1", "SYMBOL2", "SYMBOL3"}
	if len(symbols) != len(expected) {
		t.Errorf("expected %d symbols, got %d", len(expected), len(symbols))
	}

	for i, sym := range symbols {
		if sym != expected[i] {
			t.Errorf("symbol %d: expected %q, got %q", i, expected[i], sym)
		}
	}
}

// TestLoadSymbols_FileNotFound tests error handling for missing file
func TestLoadSymbols_FileNotFound(t *testing.T) {
	_, err := LoadSymbols("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestFormatTicker tests ticker formatting
func TestFormatTicker(t *testing.T) {
	cases := []struct {
		exchange string
		symbol   string
		expected string
	}{
		{"kucoin", "BTCUSDT", "KUCOIN:BTCUSDT"},
		{"BINANCE", "ETHUSDT", "BINANCE:ETHUSDT"},
		{"nasdaq", "AAPL", "NASDAQ:AAPL"},
		// Already has exchange prefix
		{"binance", "BINANCE:BTCUSDT", "BINANCE:BTCUSDT"},
		{"nasdaq", "NYSE:IBM", "NYSE:IBM"},
	}

	for _, tc := range cases {
		result := FormatTicker(tc.exchange, tc.symbol)
		if result != tc.expected {
			t.Errorf("FormatTicker(%q, %q): expected %q, got %q", tc.exchange, tc.symbol, tc.expected, result)
		}
	}
}

// TestGetFloat tests float extraction
func TestGetFloat(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]interface{}
		key      string
		expected float64
	}{
		{"float64 value", map[string]interface{}{"price": 100.5}, "price", 100.5},
		{"int value", map[string]interface{}{"price": 100}, "price", 100.0},
		{"int64 value", map[string]interface{}{"price": int64(100)}, "price", 100.0},
		{"missing key", map[string]interface{}{}, "price", 0.0},
		{"non-numeric value", map[string]interface{}{"price": "invalid"}, "price", 0.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := getFloat(tc.values, tc.key)
			if result != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, result)
			}
		})
	}
}

// TestGetInt tests int extraction
func TestGetInt(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]interface{}
		key      string
		expected int64
	}{
		{"float64 value", map[string]interface{}{"volume": 1000.0}, "volume", 1000},
		{"int value", map[string]interface{}{"volume": 2000}, "volume", 2000},
		{"int64 value", map[string]interface{}{"volume": int64(3000)}, "volume", 3000},
		{"missing key", map[string]interface{}{}, "volume", 0},
		{"non-numeric value", map[string]interface{}{"volume": "invalid"}, "volume", 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := getInt(tc.values, tc.key)
			if result != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

// TestComputeBBW tests Bollinger Band Width calculation
func TestComputeBBW(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]interface{}
		expected float64
	}{
		{
			"normal values",
			map[string]interface{}{
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			0.2, // (110 - 90) / 100 = 0.2
		},
		{
			"missing BB.upper",
			map[string]interface{}{
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			0.0,
		},
		{
			"zero SMA20",
			map[string]interface{}{
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    0,
			},
			0.0,
		},
		{
			"empty values",
			map[string]interface{}{},
			0.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := computeBBW(tc.values)
			if result != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, result)
			}
		})
	}
}

// TestComputeBBRating tests Bollinger Band rating calculation
func TestComputeBBRating(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]interface{}
		expected int
	}{
		{
			"close > bbUpper (+3)",
			map[string]interface{}{
				"close":    111.0,
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			3,
		},
		{
			"close > middle + (upper-middle)/2 (+2)",
			map[string]interface{}{
				"close":    105.5,
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			2,
		},
		{
			"close > middle (+1)",
			map[string]interface{}{
				"close":    102.0,
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			1,
		},
		{
			"close < middle (-1)",
			map[string]interface{}{
				"close":    98.0,
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			-1,
		},
		{
			"close < middle - (middle-lower)/2 (-2)",
			map[string]interface{}{
				"close":    94.5,
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			-2,
		},
		{
			"close < bbLower (-3)",
			map[string]interface{}{
				"close":    89.0,
				"BB.upper": 110.0,
				"BB.lower": 90.0,
				"SMA20":    100.0,
			},
			-3,
		},
		{
			"missing data returns 0",
			map[string]interface{}{
				"close":    100.0,
				"BB.upper": 110.0,
			},
			0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := computeBBRating(tc.values)
			if result != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

// TestBuildEntry tests building a ScreenerEntry from TVSymbolData
func TestBuildEntry(t *testing.T) {
	cases := []struct {
		name          string
		data          client.TVSymbolData
		expectNil     bool
		expectedEntry *ScreenerEntry
	}{
		{
			"valid entry",
			client.TVSymbolData{
				Symbol: "BINANCE:BTCUSDT",
				Values: map[string]interface{}{
					"open":     50000.0,
					"close":    50500.0,
					"SMA20":    49000.0,
					"BB.upper": 51000.0,
					"BB.lower": 48000.0,
					"EMA50":    49500.0,
					"RSI":      65.0,
					"volume":   1234567.0,
					"change":   1.0,
				},
			},
			false,
			&ScreenerEntry{
				Symbol:        "BINANCE:BTCUSDT",
				ChangePercent: 1.0,
				Indicators: ScreenerIndicators{
					Open:    50000.0,
					Close:   50500.0,
					SMA20:   49000.0,
					BBUpper: 51000.0,
					BBLower: 48000.0,
					EMA50:   49500.0,
					RSI:     65.0,
					Volume:  1234567.0,
				},
			},
		},
		{
			"missing EMA50 returns nil",
			client.TVSymbolData{
				Symbol: "BINANCE:ETHUSDT",
				Values: map[string]interface{}{
					"close": 3000.0,
					"RSI":   60.0,
				},
			},
			true,
			nil,
		},
		{
			"missing RSI returns nil",
			client.TVSymbolData{
				Symbol: "BINANCE:ADAUSDT",
				Values: map[string]interface{}{
					"close": 1.2,
					"EMA50": 1.1,
				},
			},
			true,
			nil,
		},
		{
			"zero EMA50 returns nil",
			client.TVSymbolData{
				Symbol: "BINANCE:BNBUSDT",
				Values: map[string]interface{}{
					"EMA50": 0.0,
					"RSI":   50.0,
				},
			},
			true,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildEntry(tc.data)
			if tc.expectNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected non-nil entry")
					return
				}
				if result.Symbol != tc.expectedEntry.Symbol {
					t.Errorf("symbol: expected %q, got %q", tc.expectedEntry.Symbol, result.Symbol)
				}
				if result.ChangePercent != tc.expectedEntry.ChangePercent {
					t.Errorf("changePercent: expected %f, got %f", tc.expectedEntry.ChangePercent, result.ChangePercent)
				}
				if result.Indicators.EMA50 != tc.expectedEntry.Indicators.EMA50 {
					t.Errorf("EMA50: expected %f, got %f", tc.expectedEntry.Indicators.EMA50, result.Indicators.EMA50)
				}
			}
		})
	}
}
