package futures

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// OpenInterestEntry is a single row in the open interest output.
type OpenInterestEntry struct {
	Symbol          string  `json:"symbol"`
	OpenInterest    float64 `json:"openInterest"`
	OpenInterestUSD float64 `json:"openInterestUsd"`
	LastPrice       float64 `json:"lastPrice"`
	ChangePct24h    float64 `json:"changePct24h"`
	Turnover24h     float64 `json:"turnover24h"`
	FundingRatePct  float64 `json:"fundingRatePct"`
	// OI/Turnover ratio: high values mean a lot of OI relative to daily volume = positioning-heavy.
	OIToTurnover float64 `json:"oiToTurnover"`
}

// RunOpenInterest lists PHEMEX perpetuals sorted by USD open interest.
// minTurnover filters out illiquid pairs.
func RunOpenInterest(cfg *config.Config, limit int, minTurnover float64) error {
	if limit <= 0 {
		limit = 25
	}
	httpClient := client.NewHTTPClient(cfg)
	phemex := client.NewPhemexClient(httpClient)

	ctx := context.Background()
	tickers, err := phemex.GetAllTickers(ctx)
	if err != nil {
		return fmt.Errorf("fetch tickers: %w", err)
	}

	var rows []OpenInterestEntry
	for _, t := range tickers {
		if !strings.HasSuffix(t.Symbol, "USDT") && !strings.HasSuffix(t.Symbol, "USDC") {
			continue
		}
		if t.Turnover < minTurnover {
			continue
		}
		oiUSD := t.OpenInterest * t.LastPrice
		ratio := 0.0
		if t.Turnover > 0 {
			ratio = oiUSD / t.Turnover
		}
		rows = append(rows, OpenInterestEntry{
			Symbol:          t.Symbol,
			OpenInterest:    t.OpenInterest,
			OpenInterestUSD: oiUSD,
			LastPrice:       t.LastPrice,
			ChangePct24h:    t.ChangePct,
			Turnover24h:     t.Turnover,
			FundingRatePct:  t.FundingRate * 100,
			OIToTurnover:    ratio,
		})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].OpenInterestUSD > rows[j].OpenInterestUSD })
	if len(rows) > limit {
		rows = rows[:limit]
	}

	if err := utils.PrintJSON(rows); err != nil {
		return fmt.Errorf("print json: %w", err)
	}
	return nil
}
