package yahoo

import (
	"context"
	"fmt"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// yahooChartBase is the API base URL for Yahoo Finance charts
var yahooChartBase = "https://query1.finance.yahoo.com/v8/finance/chart"

// PriceOutput represents the JSON output for a price query
type PriceOutput struct {
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
	Source        string    `json:"source"`
	Timestamp     time.Time `json:"timestamp"`
}

// RunYahooPrice fetches and displays the current price for a given symbol
func RunYahooPrice(cfg *config.Config, symbol string) error {
	httpClient := client.NewHTTPClient(cfg)
	yahooClient := client.NewYahooClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	result, err := yahooClient.GetFullChart(ctx, symbol, "1d", "5d")
	if err != nil {
		return fmt.Errorf("failed to fetch chart data: %w", err)
	}

	// Calculate change and percent change
	change := result.Meta.RegularMarketPrice - result.Meta.PreviousClose
	changePct := 0.0
	if result.Meta.PreviousClose != 0 {
		changePct = (change / result.Meta.PreviousClose) * 100
	}

	output := PriceOutput{
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
		Source:        "Yahoo Finance",
		Timestamp:     time.Now().UTC(),
	}

	return utils.PrintJSON(output)
}
