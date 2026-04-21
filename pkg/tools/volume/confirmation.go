package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

// VolumeConfirmationOutput represents volume confirmation analysis
type VolumeConfirmationOutput struct {
	Symbol       string                 `json:"symbol"`
	Timeframe    string                 `json:"timeframe"`
	Price        VolumePrice            `json:"price"`
	Volume       VolumeAnalysis         `json:"volume"`
	Signals      []string               `json:"signals"`
	Assessment   string                 `json:"assessment"`
	Timestamp    string                 `json:"timestamp"`
}

type VolumePrice struct {
	Close          float64 `json:"close"`
	ChangePercent  float64 `json:"changePercent"`
	Range          float64 `json:"range"`
	BodyRatio      float64 `json:"bodyRatio"`
}

type VolumeAnalysis struct {
	Current  float64 `json:"current"`
	Avg20    float64 `json:"avg20"`
	Ratio    float64 `json:"ratio"`
	Strength string  `json:"strength"`
}

// computeVolumeAssessment returns a strength assessment string
func computeVolumeAssessment(ratio float64) string {
	if ratio >= 3 {
		return "VERY STRONG"
	}
	if ratio >= 2 {
		return "STRONG"
	}
	if ratio >= 1.5 {
		return "MEDIUM"
	}
	if ratio >= 1 {
		return "NORMAL"
	}
	return "WEAK"
}

// generateSignals generates signal array based on conditions
func generateSignals(change, ratio, close, bbUpper, bbLower float64) []string {
	var signals []string

	// abs(change) > 3 AND ratio >= 2: "STRONG BREAKOUT"
	if absFloat(change) > 3 && ratio >= 2 {
		signals = append(signals, "STRONG BREAKOUT")
	}

	// abs(change) > 1 AND (close > BB.upper OR close < BB.lower): "BB BREAKOUT CONFIRMED"
	if absFloat(change) > 1 && (close > bbUpper || close < bbLower) {
		signals = append(signals, "BB BREAKOUT CONFIRMED")
	}

	// ratio >= 2 AND abs(change) < 1: "VOLUME DIVERGENCE"
	if ratio >= 2 && absFloat(change) < 1 {
		signals = append(signals, "VOLUME DIVERGENCE")
	}

	// ratio < 1: "WEAK SIGNAL"
	if ratio < 1 {
		signals = append(signals, "WEAK SIGNAL")
	}

	return signals
}

// RunVolumeConfirmation performs volume confirmation analysis on a single symbol
func RunVolumeConfirmation(cfg *config.Config, symbol, exchange, timeframe string) error {
	// Validate inputs
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if timeframe == "" {
		return fmt.Errorf("timeframe cannot be empty")
	}

	// Format ticker
	ticker := screener.FormatTicker(exchange, symbol)

	// Get screener for exchange
	screenName, err := client.ScreenerForExchange(exchange)
	if err != nil {
		return fmt.Errorf("invalid exchange: %w", err)
	}

	// Fetch analysis data
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

	// Extract values
	close := getFloat(values, "close")
	open := getFloat(values, "open")
	high := getFloat(values, "high")
	low := getFloat(values, "low")
	volume := getFloat(values, "volume")
	change := getFloat(values, "change")
	bbUpper := getFloat(values, "BB.upper")
	bbLower := getFloat(values, "BB.lower")
	volumeAvg20 := getFloat(values, "average_volume_10d_calc")

	// Calculate derived values
	priceRange := high - low
	bodyRatio := 0.0
	if priceRange > 0 {
		bodyRatio = absFloat(close-open) / priceRange
	}

	ratio := computeVolumeRatio(volume, volumeAvg20)
	strength := computeVolumeAssessment(ratio)

	// Generate signals
	signals := generateSignals(change, ratio, close, bbUpper, bbLower)

	// Determine assessment
	assessment := "NEUTRAL"
	if len(signals) > 0 && signals[0] == "STRONG BREAKOUT" {
		if change > 0 {
			assessment = "BULLISH BREAKOUT"
		} else {
			assessment = "BEARISH BREAKDOWN"
		}
	}

	output := &VolumeConfirmationOutput{
		Symbol:    ticker,
		Timeframe: timeframe,
		Price: VolumePrice{
			Close:         close,
			ChangePercent: change,
			Range:         priceRange,
			BodyRatio:     bodyRatio,
		},
		Volume: VolumeAnalysis{
			Current:  volume,
			Avg20:    volumeAvg20,
			Ratio:    ratio,
			Strength: strength,
		},
		Signals:    signals,
		Assessment: assessment,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	return utils.PrintJSON(output)
}
