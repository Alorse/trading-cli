package yahoo

import (
	"context"
	"sync"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// Quote represents a single price quote from Yahoo Finance
type Quote struct {
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	PreviousClose float64   `json:"previousClose"`
	Change        float64   `json:"change"`
	ChangePct     float64   `json:"changePct"`
	Currency      string    `json:"currency"`
	Exchange      string    `json:"exchange"`
	MarketState   string    `json:"marketState"`
	Week52High    float64   `json:"week52High"`
	Week52Low     float64   `json:"week52Low"`
}

// SnapshotOutput represents the JSON output for a market snapshot
type SnapshotOutput struct {
	Indices   []Quote   `json:"indices"`
	Crypto    []Quote   `json:"crypto"`
	FX        []Quote   `json:"fx"`
	ETFs      []Quote   `json:"etfs"`
	Timestamp time.Time `json:"timestamp"`
}

// RunMarketSnapshot fetches quotes for market indices, crypto, FX, and ETFs
func RunMarketSnapshot(cfg *config.Config) error {
	httpClient := client.NewHTTPClient(cfg)
	yahooClient := client.NewYahooClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	symbols := struct {
		Indices []string
		Crypto  []string
		FX      []string
		ETFs    []string
	}{
		Indices: []string{"^GSPC", "^DJI", "^IXIC", "^VIX"},
		Crypto:  []string{"BTC-USD", "ETH-USD", "SOL-USD", "BNB-USD"},
		FX:      []string{"EURUSD=X", "GBPUSD=X", "JPYUSD=X"},
		ETFs:    []string{"SPY", "QQQ", "GLD"},
	}

	// Fetch all symbols concurrently using a wait group
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := struct {
		Indices []Quote
		Crypto  []Quote
		FX      []Quote
		ETFs    []Quote
	}{}

	// Helper function to fetch a quote
	fetchQuote := func(symbol string) (*Quote, error) {
		result, err := yahooClient.GetFullChart(ctx, symbol, "1d", "5d")
		if err != nil {
			return nil, err
		}

		change := result.Meta.RegularMarketPrice - result.Meta.PreviousClose
		changePct := 0.0
		if result.Meta.PreviousClose != 0 {
			changePct = (change / result.Meta.PreviousClose) * 100
		}

		return &Quote{
			Symbol:        result.Meta.Symbol,
			Price:         result.Meta.RegularMarketPrice,
			PreviousClose: result.Meta.PreviousClose,
			Change:        change,
			ChangePct:     changePct,
			Currency:      result.Meta.Currency,
			Exchange:      result.Meta.ExchangeName,
			MarketState:   result.Meta.MarketState,
			Week52High:    result.Meta.FiftyTwoWeekHigh,
			Week52Low:     result.Meta.FiftyTwoWeekLow,
		}, nil
	}

	// Fetch indices concurrently
	for _, symbol := range symbols.Indices {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			quote, err := fetchQuote(sym)
			if err != nil {
				return // Skip failed requests
			}
			mu.Lock()
			results.Indices = append(results.Indices, *quote)
			mu.Unlock()
		}(symbol)
	}

	// Fetch crypto concurrently
	for _, symbol := range symbols.Crypto {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			quote, err := fetchQuote(sym)
			if err != nil {
				return
			}
			mu.Lock()
			results.Crypto = append(results.Crypto, *quote)
			mu.Unlock()
		}(symbol)
	}

	// Fetch FX concurrently
	for _, symbol := range symbols.FX {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			quote, err := fetchQuote(sym)
			if err != nil {
				return
			}
			mu.Lock()
			results.FX = append(results.FX, *quote)
			mu.Unlock()
		}(symbol)
	}

	// Fetch ETFs concurrently
	for _, symbol := range symbols.ETFs {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			quote, err := fetchQuote(sym)
			if err != nil {
				return
			}
			mu.Lock()
			results.ETFs = append(results.ETFs, *quote)
			mu.Unlock()
		}(symbol)
	}

	wg.Wait()

	output := SnapshotOutput{
		Indices:   results.Indices,
		Crypto:    results.Crypto,
		FX:        results.FX,
		ETFs:      results.ETFs,
		Timestamp: time.Now().UTC(),
	}

	return utils.PrintJSON(output)
}
