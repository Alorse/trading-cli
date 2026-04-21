package mtf

import (
	"context"
	"fmt"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

type TimeframeAnalysis struct {
	Timeframe string `json:"timeframe"`
	Bias      int    `json:"bias"`
	Reason    string `json:"reason"`
}

type MTFResult struct {
	Symbol         string              `json:"symbol"`
	Exchange       string              `json:"exchange"`
	Timeframes     []TimeframeAnalysis `json:"timeframes"`
	TotalBias      int                 `json:"totalBias"`
	Alignment      string              `json:"alignment"`
	Confidence     string              `json:"confidence"`
	Recommendation string              `json:"recommendation"`
	DivergentTFs   []string            `json:"divergentTimeframes"`
	Timestamp      string              `json:"timestamp"`
}

func RunMultiTimeframe(cfg *config.Config, symbol, exchange string) error {
	if symbol == "" {
		return fmt.Errorf("--symbol is required")
	}

	ticker := screener.FormatTicker(exchange, symbol)
	tvScreener, err := client.ScreenerForExchange(exchange)
	if err != nil {
		return err
	}

	columns := []string{
		"close", "change", "EMA9", "EMA20", "EMA50", "EMA100", "EMA200",
		"RSI", "MACD.macd", "MACD.signal", "VWAP", "volume", "volume.SMA20",
	}

	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := tvClient.GetMultipleAnalysis(ctx, tvScreener, []string{ticker}, columns)
	if err != nil {
		return fmt.Errorf("fetch analysis: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("no data returned for %s", ticker)
	}

	v := data[0].Values
	g := func(key string) float64 {
		val, ok := v[key]
		if !ok {
			return 0
		}
		f, ok := val.(float64)
		if ok {
			return f
		}
		return 0
	}

	close_ := g("close")
	change := g("change")
	ema9 := g("EMA9")
	ema20 := g("EMA20")
	ema50 := g("EMA50")
	ema200 := g("EMA200")
	rsi := g("RSI")
	macdLine := g("MACD.macd")
	macdSignal := g("MACD.signal")
	vwap := g("VWAP")

	timeframes := []struct {
		name   string
		bias   func() (int, string)
	}{
		{"1W", func() (int, string) {
			if close_ > ema200 && rsi > 50 {
				return 1, "price > EMA200 and RSI > 50"
			}
			if close_ < ema200 && rsi < 50 {
				return -1, "price < EMA200 and RSI < 50"
			}
			return 0, "neutral"
		}},
		{"1D", func() (int, string) {
			if close_ > ema50 && close_ > ema200 {
				return 1, "price > EMA50 > EMA200 (golden cross zone)"
			}
			if close_ < ema50 && close_ < ema200 {
				return -1, "price < EMA50 < EMA200 (death cross zone)"
			}
			return 0, "neutral"
		}},
		{"4h", func() (int, string) {
			if ema20 > ema50 && change > 0 {
				return 1, "EMA20 > EMA50 with bullish momentum"
			}
			if ema20 < ema50 && change < 0 {
				return -1, "EMA20 < EMA50 with bearish momentum"
			}
			return 0, "neutral"
		}},
		{"1h", func() (int, string) {
			if close_ > vwap && change > 0 {
				return 1, "price > VWAP with positive change"
			}
			if close_ < vwap && change < 0 {
				return -1, "price < VWAP with negative change"
			}
			return 0, "neutral"
		}},
		{"15m", func() (int, string) {
			if ema9 > ema20 && macdLine > macdSignal {
				return 1, "EMA9 > EMA20 and MACD bullish"
			}
			if ema9 < ema20 && macdLine < macdSignal {
				return -1, "EMA9 < EMA20 and MACD bearish"
			}
			return 0, "neutral"
		}},
	}

	var analyses []TimeframeAnalysis
	totalBias := 0
	for _, tf := range timeframes {
		bias, reason := tf.bias()
		totalBias += bias
		analyses = append(analyses, TimeframeAnalysis{
			Timeframe: tf.name,
			Bias:      bias,
			Reason:    reason,
		})
	}

	alignment, confidence, recommendation := classify(totalBias, len(timeframes))

	// Find divergent timeframes (those opposing the majority)
	majority := 1
	if totalBias < 0 {
		majority = -1
	}
	var divergent []string
	if totalBias != 0 {
		for _, a := range analyses {
			if a.Bias != 0 && a.Bias != majority {
				divergent = append(divergent, a.Timeframe)
			}
		}
	}
	if divergent == nil {
		divergent = []string{}
	}

	return utils.PrintJSON(MTFResult{
		Symbol:         ticker,
		Exchange:       exchange,
		Timeframes:     analyses,
		TotalBias:      totalBias,
		Alignment:      alignment,
		Confidence:     confidence,
		Recommendation: recommendation,
		DivergentTFs:   divergent,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	})
}

func classify(total, n int) (alignment, confidence, recommendation string) {
	switch {
	case total == n:
		return "FULLY ALIGNED BULLISH", "Very High", "STRONG BUY"
	case total == -n:
		return "FULLY ALIGNED BEARISH", "Very High", "STRONG SELL"
	case total >= 3:
		return "MOSTLY BULLISH", "High", "BUY"
	case total <= -3:
		return "MOSTLY BEARISH", "High", "SELL"
	case total > 0:
		return "LEAN BULLISH", "Medium", "CAUTIOUS BUY"
	case total < 0:
		return "LEAN BEARISH", "Medium", "CAUTIOUS SELL"
	default:
		return "MIXED/RANGING", "Low", "HOLD/NO TRADE"
	}
}
