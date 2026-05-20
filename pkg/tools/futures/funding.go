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

// FundingRateEntry is a single row in the funding rate output.
type FundingRateEntry struct {
	Symbol          string  `json:"symbol"`
	FundingRate     float64 `json:"fundingRate"`     // current period (decimal)
	FundingRatePct  float64 `json:"fundingRatePct"`  // current * 100
	PredictedRate   float64 `json:"predictedRate"`
	PredictedPct    float64 `json:"predictedPct"`
	LastPrice       float64 `json:"lastPrice"`
	ChangePct24h    float64 `json:"changePct24h"`
	Turnover24h     float64 `json:"turnover24h"`
	OpenInterest    float64 `json:"openInterest"`
	OpenInterestUSD float64 `json:"openInterestUsd"`
	// Hint: positive funding → longs pay shorts → potential long-squeeze setup if extreme.
	// Negative funding → shorts pay longs → potential short-squeeze setup if extreme.
	Bias string `json:"bias"`
}

// RunFundingRates lists PHEMEX perpetuals sorted by absolute funding rate.
// Only USDT/USDC quote perpetuals are returned. Min turnover filter excludes dead pairs.
func RunFundingRates(cfg *config.Config, mode string, limit int, minTurnover float64) error {
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

	var rows []FundingRateEntry
	for _, t := range tickers {
		if !strings.HasSuffix(t.Symbol, "USDT") && !strings.HasSuffix(t.Symbol, "USDC") {
			continue
		}
		if t.Turnover < minTurnover {
			continue
		}
		bias := "neutral"
		if t.FundingRate > 0.0001 {
			bias = "long-heavy (shorts favored)"
		} else if t.FundingRate < -0.0001 {
			bias = "short-heavy (longs favored)"
		}
		rows = append(rows, FundingRateEntry{
			Symbol:          t.Symbol,
			FundingRate:     t.FundingRate,
			FundingRatePct:  t.FundingRate * 100,
			PredictedRate:   t.PredFundingRate,
			PredictedPct:    t.PredFundingRate * 100,
			LastPrice:       t.LastPrice,
			ChangePct24h:    t.ChangePct,
			Turnover24h:     t.Turnover,
			OpenInterest:    t.OpenInterest,
			OpenInterestUSD: t.OpenInterest * t.LastPrice,
			Bias:            bias,
		})
	}

	switch mode {
	case "long-squeeze", "positive":
		// Highest positive funding first - longs overcrowded, expect mean reversion down.
		sort.Slice(rows, func(i, j int) bool { return rows[i].FundingRate > rows[j].FundingRate })
	case "short-squeeze", "negative":
		// Most negative funding first - shorts overcrowded, expect mean reversion up.
		sort.Slice(rows, func(i, j int) bool { return rows[i].FundingRate < rows[j].FundingRate })
	default:
		// Sort by absolute funding (extremes both sides).
		sort.Slice(rows, func(i, j int) bool {
			ai := rows[i].FundingRate
			aj := rows[j].FundingRate
			if ai < 0 {
				ai = -ai
			}
			if aj < 0 {
				aj = -aj
			}
			return ai > aj
		})
	}

	if len(rows) > limit {
		rows = rows[:limit]
	}

	if err := utils.PrintJSON(rows); err != nil {
		return fmt.Errorf("print json: %w", err)
	}
	return nil
}
