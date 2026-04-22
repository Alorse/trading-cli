package plan

import (
	"context"
	"fmt"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// FibonacciRetracementOutput represents the complete Fibonacci analysis
type FibonacciRetracementOutput struct {
	Symbol              string           `json:"symbol"`
	Exchange            string           `json:"exchange"`
	Lookback            string           `json:"lookback"`
	Timeframe           string           `json:"timeframe"`
	Trend               string           `json:"trend"`
	SwingHigh           float64          `json:"swingHigh"`
	SwingLow            float64          `json:"swingLow"`
	CurrentPrice        float64          `json:"currentPrice"`
	RetracementLevels   []FibonacciLevel `json:"retracementLevels"`
	ExtensionLevels     []FibonacciLevel `json:"extensionLevels"`
	NearestLevel        FibonacciLevel   `json:"nearestLevel"`
	GoldenPocket        GoldenPocket     `json:"goldenPocket"`
	CurrentDepthPercent float64          `json:"currentDepthPercent"`
	Timestamp           time.Time        `json:"timestamp"`
}

// FibonacciLevel represents a single Fibonacci level
type FibonacciLevel struct {
	Ratio    float64 `json:"ratio"`
	Level    float64 `json:"level"`
	Distance float64 `json:"distance"`
}

// GoldenPocket represents the golden pocket zone (0.618-0.786)
type GoldenPocket struct {
	Lower    float64 `json:"lower"`
	Upper    float64 `json:"upper"`
	IsInZone bool    `json:"isInZone"`
}

// mapLookback converts lookback string to Yahoo Finance range
func mapLookback(lookback string) string {
	switch lookback {
	case "1M":
		return "1mo"
	case "3M":
		return "3mo"
	case "6M":
		return "6mo"
	case "52W":
		return "1y"
	case "ALL":
		return "5y"
	default:
		return "1y"
	}
}

// findSwings finds the swing high and low in the candles
func findSwings(candles []client.YahooOHLCV) (float64, float64) {
	if len(candles) == 0 {
		return 0, 0
	}

	swingHigh := candles[0].Close
	swingLow := candles[0].Close

	for _, candle := range candles {
		if candle.Close > swingHigh {
			swingHigh = candle.Close
		}
		if candle.Close < swingLow {
			swingLow = candle.Close
		}
	}

	return swingHigh, swingLow
}

// detectTrend determines trend based on current price and swings
func detectTrend(currentPrice, swingHigh, swingLow float64, candles []client.YahooOHLCV) string {
	midpoint := (swingHigh + swingLow) / 2

	// Check if swing high is recent (in the last 25% of candles)
	recentStart := len(candles) - (len(candles) / 4)
	if recentStart < 0 {
		recentStart = 0
	}

	swingHighRecent := false
	for i := recentStart; i < len(candles); i++ {
		if candles[i].Close >= swingHigh*0.99 { // within 1% of high
			swingHighRecent = true
			break
		}
	}

	if currentPrice > midpoint && swingHighRecent {
		return "uptrend"
	}
	if currentPrice < midpoint && !swingHighRecent {
		return "downtrend"
	}

	// Default based on position
	if currentPrice > midpoint {
		return "uptrend"
	}
	return "downtrend"
}

// computeFibonacciLevels computes retracement and extension levels
func computeFibonacciLevels(swingHigh, swingLow float64, trend string) ([]FibonacciLevel, []FibonacciLevel) {
	ratios := []float64{0, 0.236, 0.382, 0.5, 0.618, 0.786, 1.0}
	extensionRatios := []float64{1.272, 1.618, 2.618}

	rangeVal := swingHigh - swingLow
	retracementLevels := make([]FibonacciLevel, len(ratios))
	extensionLevels := make([]FibonacciLevel, len(extensionRatios))

	if trend == "uptrend" {
		// For uptrend: level = swingLow + ratio * range
		for i, ratio := range ratios {
			level := swingLow + (ratio * rangeVal)
			retracementLevels[i] = FibonacciLevel{
				Ratio: ratio,
				Level: level,
			}
		}
		for i, ratio := range extensionRatios {
			level := swingLow + (ratio * rangeVal)
			extensionLevels[i] = FibonacciLevel{
				Ratio: ratio,
				Level: level,
			}
		}
	} else {
		// For downtrend: level = swingHigh - ratio * range
		for i, ratio := range ratios {
			level := swingHigh - (ratio * rangeVal)
			retracementLevels[i] = FibonacciLevel{
				Ratio: ratio,
				Level: level,
			}
		}
		for i, ratio := range extensionRatios {
			level := swingHigh - (ratio * rangeVal)
			extensionLevels[i] = FibonacciLevel{
				Ratio: ratio,
				Level: level,
			}
		}
	}

	return retracementLevels, extensionLevels
}

// findNearestLevel finds the nearest Fibonacci level to current price
func findNearestLevel(currentPrice float64, retracementLevels, extensionLevels []FibonacciLevel) FibonacciLevel {
	allLevels := append(retracementLevels, extensionLevels...)

	nearest := allLevels[0]
	minDist := abs(currentPrice - nearest.Level)

	for _, level := range allLevels {
		dist := abs(currentPrice - level.Level)
		if dist < minDist {
			minDist = dist
			nearest = level
		}
	}

	nearest.Distance = minDist
	return nearest
}

// computeGoldenPocket computes the golden pocket zone (0.618-0.786)
func computeGoldenPocket(swingHigh, swingLow float64, trend string) GoldenPocket {
	rangeVal := swingHigh - swingLow

	var lower, upper float64
	if trend == "uptrend" {
		lower = swingLow + (0.618 * rangeVal)
		upper = swingLow + (0.786 * rangeVal)
	} else {
		upper = swingHigh - (0.618 * rangeVal)
		lower = swingHigh - (0.786 * rangeVal)
	}

	return GoldenPocket{
		Lower:    lower,
		Upper:    upper,
		IsInZone: false, // Will be set when we know current price
	}
}

// computeCurrentDepth computes how deep we are in the retracement (0-100%)
func computeCurrentDepth(currentPrice, swingHigh, swingLow float64, trend string) float64 {
	rangeVal := swingHigh - swingLow
	if rangeVal == 0 {
		return 0
	}

	if trend == "uptrend" {
		depth := (currentPrice - swingLow) / rangeVal * 100
		if depth < 0 {
			depth = 0
		}
		if depth > 100 {
			depth = 100
		}
		return depth
	}

	depth := (swingHigh - currentPrice) / rangeVal * 100
	if depth < 0 {
		depth = 0
	}
	if depth > 100 {
		depth = 100
	}
	return depth
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// RunFibonacci performs Fibonacci retracement analysis
func RunFibonacci(cfg *config.Config, symbol, exchange, lookback, timeframe string) error {
	// Validate inputs
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if exchange == "" {
		return fmt.Errorf("exchange cannot be empty")
	}
	if lookback == "" {
		return fmt.Errorf("lookback cannot be empty")
	}
	if timeframe == "" {
		return fmt.Errorf("timeframe cannot be empty")
	}

	// Map lookback to Yahoo Finance range
	yahooRange := mapLookback(lookback)

	// Create HTTP and Yahoo clients
	httpClient := client.NewHTTPClient(cfg)
	yahooClient := client.NewYahooClient(httpClient)

	// Fetch candles
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	result, err := yahooClient.GetFullChart(ctx, symbol, "1d", yahooRange)
	if err != nil {
		return fmt.Errorf("fetch chart: %w", err)
	}

	if len(result.Candles) == 0 {
		return fmt.Errorf("no candles retrieved for %s", symbol)
	}

	// Find swings
	swingHigh, swingLow := findSwings(result.Candles)
	if swingHigh == swingLow {
		return fmt.Errorf("cannot compute Fibonacci: swing high equals swing low")
	}

	// Get current price (last candle close)
	currentPrice := result.Candles[len(result.Candles)-1].Close

	// Detect trend
	trend := detectTrend(currentPrice, swingHigh, swingLow, result.Candles)

	// Compute Fibonacci levels
	retracementLevels, extensionLevels := computeFibonacciLevels(swingHigh, swingLow, trend)

	// Find nearest level
	nearestLevel := findNearestLevel(currentPrice, retracementLevels, extensionLevels)

	// Compute golden pocket
	goldenPocket := computeGoldenPocket(swingHigh, swingLow, trend)
	goldenPocket.IsInZone = currentPrice >= goldenPocket.Lower && currentPrice <= goldenPocket.Upper

	// Compute current depth
	currentDepth := computeCurrentDepth(currentPrice, swingHigh, swingLow, trend)

	// Build output
	output := FibonacciRetracementOutput{
		Symbol:              symbol,
		Exchange:            exchange,
		Lookback:            lookback,
		Timeframe:           timeframe,
		Trend:               trend,
		SwingHigh:           swingHigh,
		SwingLow:            swingLow,
		CurrentPrice:        currentPrice,
		RetracementLevels:   retracementLevels,
		ExtensionLevels:     extensionLevels,
		NearestLevel:        nearestLevel,
		GoldenPocket:        goldenPocket,
		CurrentDepthPercent: currentDepth,
		Timestamp:           time.Now().UTC(),
	}

	return utils.PrintJSON(output)
}
