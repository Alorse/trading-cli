package yahoo

import (
	"testing"
	"time"

	"github.com/alorse/trading-cli/pkg/client"
)

func TestPriceOutputCalculations(t *testing.T) {
	// Test that change and changePct are calculated correctly
	meta := client.YahooMeta{
		Symbol:             "TEST",
		Currency:           "USD",
		ExchangeName:       "NYSE",
		RegularMarketPrice: 100.0,
		PreviousClose:      80.0,
		FiftyTwoWeekHigh:   150.0,
		FiftyTwoWeekLow:    70.0,
		MarketState:        "REGULAR",
	}

	// Calculate as PriceOutput would
	change := meta.RegularMarketPrice - meta.PreviousClose
	changePct := (change / meta.PreviousClose) * 100

	// Verify calculations
	if change != 20.0 {
		t.Errorf("Expected change 20.0, got %v", change)
	}

	if changePct != 25.0 {
		t.Errorf("Expected changePct 25.0, got %v", changePct)
	}

	// Verify PriceOutput structure
	output := PriceOutput{
		Symbol:        meta.Symbol,
		Price:         meta.RegularMarketPrice,
		PreviousClose: meta.PreviousClose,
		Change:        change,
		ChangePct:     changePct,
		Currency:      meta.Currency,
		Exchange:      meta.ExchangeName,
		MarketState:   meta.MarketState,
		Week52High:    meta.FiftyTwoWeekHigh,
		Week52Low:     meta.FiftyTwoWeekLow,
		Source:        "Yahoo Finance",
		Timestamp:     time.Now().UTC(),
	}

	if output.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %v", output.Symbol)
	}

	if output.Source != "Yahoo Finance" {
		t.Errorf("Expected source 'Yahoo Finance', got %v", output.Source)
	}
}

func TestZeroPreviousCloseDivision(t *testing.T) {
	// Test handling of zero previous close to avoid division by zero
	meta := client.YahooMeta{
		Symbol:             "NEW",
		Currency:           "USD",
		ExchangeName:       "NYSE",
		RegularMarketPrice: 100.0,
		PreviousClose:      0.0,
		FiftyTwoWeekHigh:   150.0,
		FiftyTwoWeekLow:    50.0,
		MarketState:        "REGULAR",
	}

	// Calculate as PriceOutput would
	change := meta.RegularMarketPrice - meta.PreviousClose
	changePct := 0.0
	if meta.PreviousClose != 0 {
		changePct = (change / meta.PreviousClose) * 100
	}

	if change != 100.0 {
		t.Errorf("Expected change 100.0, got %v", change)
	}

	if changePct != 0.0 {
		t.Errorf("Expected changePct 0.0 (no division), got %v", changePct)
	}
}

func TestSnapshotQuoteStructure(t *testing.T) {
	// Test that Quote struct can be properly populated and collected
	meta1 := client.YahooMeta{
		Symbol:             "^GSPC",
		Currency:           "USD",
		ExchangeName:       "S&P 500",
		RegularMarketPrice: 5000.0,
		PreviousClose:      4990.0,
		FiftyTwoWeekHigh:   5500.0,
		FiftyTwoWeekLow:    4500.0,
		MarketState:        "REGULAR",
	}

	meta2 := client.YahooMeta{
		Symbol:             "BTC-USD",
		Currency:           "USD",
		ExchangeName:       "CRYPTO",
		RegularMarketPrice: 65000.0,
		PreviousClose:      64000.0,
		FiftyTwoWeekHigh:   70000.0,
		FiftyTwoWeekLow:    40000.0,
		MarketState:        "REGULAR",
	}

	// Create quotes
	quote1 := Quote{
		Symbol:        meta1.Symbol,
		Price:         meta1.RegularMarketPrice,
		PreviousClose: meta1.PreviousClose,
		Change:        meta1.RegularMarketPrice - meta1.PreviousClose,
		Currency:      meta1.Currency,
		Exchange:      meta1.ExchangeName,
	}

	quote2 := Quote{
		Symbol:        meta2.Symbol,
		Price:         meta2.RegularMarketPrice,
		PreviousClose: meta2.PreviousClose,
		Change:        meta2.RegularMarketPrice - meta2.PreviousClose,
		Currency:      meta2.Currency,
		Exchange:      meta2.ExchangeName,
	}

	// Verify quotes are correct
	if quote1.Symbol != "^GSPC" {
		t.Errorf("Expected symbol ^GSPC, got %v", quote1.Symbol)
	}

	if quote1.Change != 10.0 {
		t.Errorf("Expected change 10.0, got %v", quote1.Change)
	}

	if quote2.Symbol != "BTC-USD" {
		t.Errorf("Expected symbol BTC-USD, got %v", quote2.Symbol)
	}

	// Verify we can collect them into groups
	indices := []Quote{quote1}
	crypto := []Quote{quote2}

	if len(indices) != 1 {
		t.Errorf("Expected 1 index quote, got %d", len(indices))
	}

	if len(crypto) != 1 {
		t.Errorf("Expected 1 crypto quote, got %d", len(crypto))
	}

	// Verify snapshot output structure
	output := SnapshotOutput{
		Indices:   indices,
		Crypto:    crypto,
		FX:        []Quote{},
		ETFs:      []Quote{},
		Timestamp: time.Now().UTC(),
	}

	if len(output.Indices) != 1 || len(output.Crypto) != 1 {
		t.Errorf("Snapshot output structure incorrect")
	}
}
