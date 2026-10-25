package futures

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

// ScalpCandidate is a ranked entry combining TV technicals + PHEMEX futures data.
type ScalpCandidate struct {
	Symbol          string          `json:"symbol"` // e.g. "BTCUSDT.P"
	Score           int             `json:"score"`  // 0-100 confluence score
	ChangePct1h     float64         `json:"changePct1h"`
	ChangePct24h    float64         `json:"changePct24h"`
	RSI1h           float64         `json:"rsi1h"`
	VolumeRatio     float64         `json:"volumeRatio"` // current vs 10d avg
	FundingRatePct  float64         `json:"fundingRatePct"`
	OpenInterestUSD float64         `json:"openInterestUsd"`
	Turnover24h     float64         `json:"turnover24h"`
	LastPrice       float64         `json:"lastPrice"`
	BBPosition      float64         `json:"bbPosition"` // 0=bottom, 0.5=middle, 1=top
	Reasons         []string        `json:"reasons"`
	Trade           ScalpTradeSetup `json:"trade"`
}

// ScalpTradeSetup is the suggested setup for a 5x scalp.
// Risk model: SL = -1% price = -5% account at 5x; TP1 = +2% price = +10% account.
type ScalpTradeSetup struct {
	Entry             float64 `json:"entry"`
	StopLoss          float64 `json:"stopLoss"`          // entry * 0.99 (-1%)
	TakeProfit1       float64 `json:"takeProfit1"`       // entry * 1.02 (+2%) -> exit 50%
	TakeProfit2       float64 `json:"takeProfit2"`       // entry * 1.04 (+4%) -> exit rest
	RiskAccountPct    float64 `json:"riskAccountPct"`    // 5.0 at 5x leverage
	RewardAccountPct1 float64 `json:"rewardAccountPct1"` // 10.0
	RewardAccountPct2 float64 `json:"rewardAccountPct2"` // 20.0
	SuggestedLeverage int     `json:"suggestedLeverage"`
}

func buildScalpTrade(entry float64) ScalpTradeSetup {
	return ScalpTradeSetup{
		Entry:             entry,
		StopLoss:          entry * 0.99,
		TakeProfit1:       entry * 1.02,
		TakeProfit2:       entry * 1.04,
		RiskAccountPct:    5.0,  // -1% price * 5x leverage
		RewardAccountPct1: 10.0, // +2% price * 5x
		RewardAccountPct2: 20.0, // +4% price * 5x
		SuggestedLeverage: 5,
	}
}

// phemexAPISymbol converts a TV-style ticker (e.g. "PHEMEX:BTCUSDT.P" or "BTCUSDT.P") to PHEMEX API symbol ("BTCUSDT").
func phemexAPISymbol(tvSymbol string) string {
	s := tvSymbol
	if idx := strings.Index(s, ":"); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimSuffix(s, ".P")
	return s
}

// RunScalpScan finds high-confluence scalp opportunities on PHEMEX perpetuals.
// Combines TradingView 1h technicals (price/RSI/BB/volume) with PHEMEX futures data
// (funding rate, open interest, 24h change). Scores each by a confluence heuristic
// and returns the top N.
//
// minScore filters out low-quality candidates (default 50).
// minTurnover filters out illiquid contracts (default 1M USDT).
func RunScalpScan(cfg *config.Config, limit int, minScore int, minTurnover float64) error {
	if limit <= 0 {
		limit = 15
	}
	if minScore < 0 {
		minScore = 0
	}

	// 1. Load PHEMEX futures symbol list from embedded data.
	tvSymbols, err := screener.LoadSymbols("PHEMEX", true)
	if err != nil {
		return fmt.Errorf("load phemex symbols: %w", err)
	}

	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)
	phemex := client.NewPhemexClient(httpClient)

	// 2. Fetch TV 1h analysis in parallel with PHEMEX tickers.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	type tvResult struct {
		data []client.TVSymbolData
		err  error
	}
	type phResult struct {
		data []client.PhemexTicker
		err  error
	}
	tvCh := make(chan tvResult, 1)
	phCh := make(chan phResult, 1)

	columns := []string{
		"open", "high", "low", "close", "volume",
		"change", "RSI", "BB.upper", "BB.lower", "SMA20", "EMA50",
		"relative_volume_10d_calc",
	}
	columns = screener.ApplyTimeframe(columns, "1h")

	go func() {
		results, err := tvClient.GetMultipleAnalysis(ctx, "crypto", tvSymbols, columns)
		tvCh <- tvResult{data: results, err: err}
	}()
	go func() {
		tickers, err := phemex.GetAllTickers(ctx)
		phCh <- phResult{data: tickers, err: err}
	}()

	tvRes := <-tvCh
	phRes := <-phCh
	if tvRes.err != nil {
		return fmt.Errorf("tv fetch: %w", tvRes.err)
	}
	if phRes.err != nil {
		return fmt.Errorf("phemex fetch: %w", phRes.err)
	}

	// Normalize TV result keys back to unsuffixed.
	tvRes.data = screener.NormalizeResults(tvRes.data, "1h")

	// Index PHEMEX tickers by API symbol.
	phMap := make(map[string]client.PhemexTicker, len(phRes.data))
	for _, t := range phRes.data {
		phMap[t.Symbol] = t
	}

	// 3. Build and score candidates.
	getF := func(v map[string]interface{}, key string) float64 {
		if x, ok := v[key]; ok {
			switch t := x.(type) {
			case float64:
				return t
			}
		}
		return 0
	}

	var candidates []ScalpCandidate
	for _, row := range tvRes.data {
		vals := row.Values
		apiSym := phemexAPISymbol(row.Symbol)
		ph, ok := phMap[apiSym]
		if !ok {
			continue
		}
		if ph.Turnover < minTurnover {
			continue
		}

		close := getF(vals, "close")
		change1h := getF(vals, "change")
		rsi := getF(vals, "RSI")
		bbUp := getF(vals, "BB.upper")
		bbLo := getF(vals, "BB.lower")
		volRatio := getF(vals, "relative_volume_10d_calc")
		if close == 0 || rsi == 0 || bbUp == 0 {
			continue
		}

		// BB position: 0 = at lower band, 1 = at upper band.
		bbPos := 0.5
		if bbUp > bbLo {
			bbPos = (close - bbLo) / (bbUp - bbLo)
		}

		score := 0
		reasons := []string{}

		// Momentum freshness: 1h change between +1% and +5% is the sweet spot.
		switch {
		case change1h >= 1.0 && change1h <= 5.0:
			score += 25
			reasons = append(reasons, fmt.Sprintf("fresh-momentum (1h %+.2f%%)", change1h))
		case change1h > 5.0 && change1h <= 8.0:
			score += 10
			reasons = append(reasons, fmt.Sprintf("strong-momentum (1h %+.2f%%)", change1h))
		case change1h > 8.0:
			score -= 10
			reasons = append(reasons, fmt.Sprintf("late-FOMO-risk (1h %+.2f%%)", change1h))
		case change1h >= 0 && change1h < 1.0:
			score += 5
		default:
			score -= 5
		}

		// Volume confirmation.
		if volRatio >= 3.0 {
			score += 25
			reasons = append(reasons, fmt.Sprintf("volume-explosion (%.1fx avg)", volRatio))
		} else if volRatio >= 1.5 {
			score += 15
			reasons = append(reasons, fmt.Sprintf("volume-spike (%.1fx avg)", volRatio))
		} else if volRatio < 0.8 {
			score -= 10
		}

		// RSI: bullish-but-not-cooked range.
		switch {
		case rsi >= 50 && rsi <= 68:
			score += 15
			reasons = append(reasons, fmt.Sprintf("rsi-bullish (%.1f)", rsi))
		case rsi > 68 && rsi <= 75:
			score += 5
		case rsi > 75:
			score -= 15
			reasons = append(reasons, fmt.Sprintf("rsi-overbought (%.1f)", rsi))
		case rsi < 35:
			// Oversold could be a reversal setup but risky for a 5x long.
			score -= 5
		}

		// BB position: prefer middle / lower-half of bands (room to run up).
		switch {
		case bbPos < 0.6:
			score += 10
			reasons = append(reasons, "bb-room-to-run")
		case bbPos > 0.9:
			score -= 10
			reasons = append(reasons, "bb-pinned-upper")
		}

		// Funding rate signal.
		fpct := ph.FundingRate * 100.0
		switch {
		case fpct > 0.08:
			// Very high positive funding → longs overcrowded → contrarian short setup OR long-squeeze risk.
			score -= 15
			reasons = append(reasons, fmt.Sprintf("funding-extreme-positive (%+.3f%%)", fpct))
		case fpct >= -0.02 && fpct <= 0.05:
			score += 10
			reasons = append(reasons, fmt.Sprintf("funding-healthy (%+.3f%%)", fpct))
		case fpct < -0.05:
			// Negative funding while we want to long: shorts paying us → bullish bias.
			score += 8
			reasons = append(reasons, fmt.Sprintf("funding-shorts-pay (%+.3f%%)", fpct))
		}

		// 24h day bias.
		if ph.ChangePct > 0 {
			score += 5
		} else if ph.ChangePct < -5 {
			score -= 5
		}

		if score < minScore {
			continue
		}

		candidates = append(candidates, ScalpCandidate{
			Symbol:          strings.TrimPrefix(row.Symbol, "PHEMEX:"),
			Score:           score,
			ChangePct1h:     change1h,
			ChangePct24h:    ph.ChangePct,
			RSI1h:           rsi,
			VolumeRatio:     volRatio,
			FundingRatePct:  fpct,
			OpenInterestUSD: ph.OpenInterest * ph.LastPrice,
			Turnover24h:     ph.Turnover,
			LastPrice:       ph.LastPrice,
			BBPosition:      bbPos,
			Reasons:         reasons,
			Trade:           buildScalpTrade(ph.LastPrice),
		})
	}

	// Sort by score desc, then by volume ratio.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].VolumeRatio > candidates[j].VolumeRatio
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	out := struct {
		GeneratedAt string           `json:"generatedAt"`
		Count       int              `json:"count"`
		MinScore    int              `json:"minScore"`
		MinTurnover float64          `json:"minTurnover"`
		Note        string           `json:"note"`
		Candidates  []ScalpCandidate `json:"candidates"`
	}{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Count:       len(candidates),
		MinScore:    minScore,
		MinTurnover: minTurnover,
		Note:        "Scalp setups: 5x leverage, -1% SL (-5% account), +2%/+4% TP (10%/20% account). Paper-trade first.",
		Candidates:  candidates,
	}

	if err := utils.PrintJSON(out); err != nil {
		return fmt.Errorf("print json: %w", err)
	}
	return nil
}
