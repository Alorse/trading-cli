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

// TimeframeAnalysis represents bias analysis for a single timeframe
type TimeframeAnalysis struct {
	Timeframe string `json:"timeframe"`
	Bias      int    `json:"bias"`
	Reason    string `json:"reason"`
}

// MTFResult represents the complete multi-timeframe analysis result
type MTFResult struct {
	Symbol         string                `json:"symbol"`
	Exchange       string                `json:"exchange"`
	Timeframes     []TimeframeAnalysis   `json:"timeframes"`
	TotalBias      int                   `json:"totalBias"`
	Alignment      string                `json:"alignment"`
	Confidence     string                `json:"confidence"`
	Recommendation string                `json:"recommendation"`
	DivergentTFs   []string              `json:"divergentTimeframes"`
	Timestamp      string                `json:"timestamp"`
}

// getFloat safely extracts a float64 value
func getFloat(values map[string]interface{}, key string) float64 {
	val, ok := values[key]
	if !ok {
		return 0
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// computeBias computes bias for a specific timeframe
func computeBias(timeframe string, close, ema200, ema100, ema50, ema20, ema9, rsi, change, vwap, macdLine, macdSignal, relVolume float64) int {
	switch timeframe {
	case "1W":
		score := 0
		if ema100 > ema200 {
			score++
		}
		if macdLine > macdSignal {
			score++
		}
		if rsi > 50 {
			score++
		}
		if score >= 2 {
			return 1
		}
		if score <= 0 {
			return -1
		}
		return 0

	case "1D":
		score := 0
		if close > ema50 && ema50 > ema200 {
			score++ // golden cross
		}
		if rsi >= 40 && rsi <= 60 {
			score = score // neutral, no bias
		} else if rsi > 60 {
			score++
		} else {
			score--
		}
		if relVolume > 1.0 {
			score++
		}
		if macdLine > macdSignal {
			score++
		}
		if score >= 2 {
			return 1
		}
		if score <= -1 {
			return -1
		}
		return 0

	case "4h":
		score := 0
		if ema20 > ema50 {
			score++
		}
		if macdLine > macdSignal {
			score++
		}
		if score >= 2 {
			return 1
		}
		if score <= 0 {
			return -1
		}
		return 0

	case "1h":
		score := 0
		if close > ema20 {
			score++
		}
		if relVolume > 1.5 {
			score++
		}
		if close > vwap {
			score++
		}
		if score >= 2 {
			return 1
		}
		if score <= 0 {
			return -1
		}
		return 0

	case "15m":
		score := 0
		if ema9 > ema20 {
			score++
		}
		if close > vwap {
			score++
		}
		if score >= 2 {
			return 1
		}
		if score <= 0 {
			return -1
		}
		return 0

	default:
		return 0
	}
}

// computeReason generates a reason string for the bias
func computeReason(timeframe string, bias int, values map[string]interface{}) string {
	if bias == 0 {
		return "Neutral conditions"
	}

	switch timeframe {
	case "1W":
		ema100 := getFloat(values, "EMA100")
		ema200 := getFloat(values, "EMA200")
		rsi := getFloat(values, "RSI")
		if bias > 0 {
			return fmt.Sprintf("Bullish: EMA100 (%.2f) > EMA200 (%.2f), MACD bullish, RSI %.1f", ema100, ema200, rsi)
		}
		return fmt.Sprintf("Bearish: EMA100 (%.2f) < EMA200 (%.2f), MACD bearish, RSI %.1f", ema100, ema200, rsi)

	case "1D":
		rsi := getFloat(values, "RSI")
		relVolume := getFloat(values, "relative_volume_10d_calc")
		if bias > 0 {
			return fmt.Sprintf("Golden cross, RSI %.1f, volume ratio %.2fx, MACD bullish", rsi, relVolume)
		}
		return fmt.Sprintf("Death cross or weak momentum, RSI %.1f, volume ratio %.2fx, MACD bearish", rsi, relVolume)

	case "4h":
		ema20 := getFloat(values, "EMA20")
		ema50 := getFloat(values, "EMA50")
		if bias > 0 {
			return fmt.Sprintf("EMA20 (%.2f) > EMA50 (%.2f), MACD bullish", ema20, ema50)
		}
		return fmt.Sprintf("EMA20 (%.2f) < EMA50 (%.2f), MACD bearish", ema20, ema50)

	case "1h":
		close := getFloat(values, "close")
		vwap := getFloat(values, "VWAP")
		relVolume := getFloat(values, "relative_volume_10d_calc")
		if bias > 0 {
			return fmt.Sprintf("Close above EMA20 (%.2f) and VWAP (%.2f), volume spike %.2fx", close, vwap, relVolume)
		}
		return fmt.Sprintf("Close below EMA20 (%.2f) or VWAP (%.2f), volume %.2fx", close, vwap, relVolume)

	case "15m":
		ema9 := getFloat(values, "EMA9")
		ema20 := getFloat(values, "EMA20")
		vwap := getFloat(values, "VWAP")
		if bias > 0 {
			return fmt.Sprintf("EMA9 (%.2f) > EMA20 (%.2f), above VWAP (%.2f)", ema9, ema20, vwap)
		}
		return fmt.Sprintf("EMA9 (%.2f) < EMA20 (%.2f), below VWAP (%.2f)", ema9, ema20, vwap)

	default:
		direction := "Bullish"
		if bias < 0 {
			direction = "Bearish"
		}
		return fmt.Sprintf("%s conditions detected", direction)
	}
}

// computeAlignment determines alignment string from total bias
func computeAlignment(totalBias int) string {
	switch {
	case totalBias == 5:
		return "FULLY ALIGNED BULLISH"
	case totalBias == -5:
		return "FULLY ALIGNED BEARISH"
	case totalBias >= 3:
		return "MOSTLY BULLISH"
	case totalBias <= -3:
		return "MOSTLY BEARISH"
	case totalBias > 0:
		return "LEAN BULLISH"
	case totalBias < 0:
		return "LEAN BEARISH"
	default:
		return "MIXED/RANGING"
	}
}

// computeConfidence determines confidence level from total bias
func computeConfidence(totalBias int) string {
	absBias := totalBias
	if absBias < 0 {
		absBias = -absBias
	}

	switch absBias {
	case 5:
		return "Very High"
	case 3, 4:
		return "High"
	case 1, 2:
		return "Medium"
	default:
		return "Low"
	}
}

// computeRecommendation determines recommendation from total bias
func computeRecommendation(totalBias int) string {
	switch totalBias {
	case 5:
		return "STRONG BUY"
	case -5:
		return "STRONG SELL"
	case 3, 4:
		return "BUY"
	case -3, -4:
		return "SELL"
	case 1, 2:
		return "CAUTIOUS BUY"
	case -1, -2:
		return "CAUTIOUS SELL"
	default:
		return "HOLD/NO TRADE"
	}
}

// findDivergentTimeframes identifies timeframes that diverge from the majority
func findDivergentTimeframes(biases map[string]int, totalBias int) []string {
	divergent := make([]string, 0)

	// Determine majority direction
	majorityPositive := totalBias > 0

	for tf, bias := range biases {
		// Check if this timeframe disagrees with majority direction
		if majorityPositive && bias < 0 {
			divergent = append(divergent, tf)
		} else if !majorityPositive && bias > 0 {
			divergent = append(divergent, tf)
		}
	}

	return divergent
}

// classify is a helper function for backwards compatibility with existing tests
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

// RunMultiTimeframe performs multi-timeframe analysis on a symbol
func RunMultiTimeframe(cfg *config.Config, symbol, exchange string) error {
	// Validate inputs
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}

	// Format ticker
	ticker := screener.FormatTicker(exchange, symbol)

	// Get screener for exchange
	screenName, err := client.ScreenerForExchange(exchange)
	if err != nil {
		return fmt.Errorf("invalid exchange: %w", err)
	}

	// Fetch analysis data (same data for all timeframes since TV scanner doesn't differentiate)
	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	results, err := tvClient.GetMultipleAnalysis(ctx, screenName, []string{ticker}, client.DefaultColumns)
	if err != nil {
		return fmt.Errorf("fetch analysis: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no data returned for symbol %s", ticker)
	}

	values := results[0].Values

	// Extract all needed values
	close := getFloat(values, "close")
	ema9 := getFloat(values, "EMA9")
	ema20 := getFloat(values, "EMA20")
	ema50 := getFloat(values, "EMA50")
	ema100 := getFloat(values, "EMA100")
	ema200 := getFloat(values, "EMA200")
	rsi := getFloat(values, "RSI")
	change := getFloat(values, "change")
	vwap := getFloat(values, "VWAP")
	macdLine := getFloat(values, "MACD.macd")
	macdSignal := getFloat(values, "MACD.signal")
	relVolume := getFloat(values, "relative_volume_10d_calc")

	// Analyze each timeframe
	timeframes := []string{"1W", "1D", "4h", "1h", "15m"}
	timeframeAnalyses := make([]TimeframeAnalysis, len(timeframes))
	biasMap := make(map[string]int)
	totalBias := 0

	for i, tf := range timeframes {
		bias := computeBias(tf, close, ema200, ema100, ema50, ema20, ema9, rsi, change, vwap, macdLine, macdSignal, relVolume)
		reason := computeReason(tf, bias, values)

		timeframeAnalyses[i] = TimeframeAnalysis{
			Timeframe: tf,
			Bias:      bias,
			Reason:    reason,
		}

		biasMap[tf] = bias
		totalBias += bias
	}

	// Calculate final metrics
	alignment := computeAlignment(totalBias)
	confidence := computeConfidence(totalBias)
	recommendation := computeRecommendation(totalBias)
	divergentTFs := findDivergentTimeframes(biasMap, totalBias)

	result := &MTFResult{
		Symbol:         ticker,
		Exchange:       exchange,
		Timeframes:     timeframeAnalyses,
		TotalBias:      totalBias,
		Alignment:      alignment,
		Confidence:     confidence,
		Recommendation: recommendation,
		DivergentTFs:   divergentTFs,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}

	return utils.PrintJSON(result)
}
