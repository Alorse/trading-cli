package plan

import (
	"context"
	"fmt"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

// TradeSetup represents the entry, stop loss, and targets for a trade
type TradeSetup struct {
	Entry    float64 `json:"entry"`
	StopLoss float64 `json:"stopLoss"`
	Target1  float64 `json:"target1"`
	Target2  float64 `json:"target2"`
	Target3  float64 `json:"target3"`
	RR1      float64 `json:"rr1"`
	RR2      float64 `json:"rr2"`
	RR3      float64 `json:"rr3"`
}

// TradePlanOutput represents the complete trade plan JSON output
type TradePlanOutput struct {
	Symbol         string       `json:"symbol"`
	Exchange       string       `json:"exchange"`
	Timeframe      string       `json:"timeframe"`
	StockScore     int          `json:"stockScore"`
	Grade          string       `json:"grade"`
	TradeQuality   int          `json:"tradeQuality"`
	Setup          TradeSetup   `json:"setup"`
	Recommendation string       `json:"recommendation"`
	Details        TradeDetails `json:"details"`
	Timestamp      time.Time    `json:"timestamp"`
}

// TradeDetails provides breakdown of scores and analysis
type TradeDetails struct {
	ScoreComponents  ScoreComponents  `json:"scoreComponents"`
	QualityBreakdown QualityBreakdown `json:"qualityBreakdown"`
	Indicators       IndicatorValues  `json:"indicators"`
}

// ScoreComponents breaks down the stock score
type ScoreComponents struct {
	EMAAlignment int `json:"emaAlignment"`
	RSIScore     int `json:"rsiScore"`
	MACDScore    int `json:"macdScore"`
	VolumeScore  int `json:"volumeScore"`
	ADXScore     int `json:"adxScore"`
}

// QualityBreakdown breaks down quality score by component
type QualityBreakdown struct {
	Structure  int `json:"structure"`
	RewardRisk int `json:"rewardRisk"`
	Volume     int `json:"volume"`
	StopSize   int `json:"stopSize"`
	Liquidity  int `json:"liquidity"`
}

// IndicatorValues holds raw indicator values
type IndicatorValues struct {
	Close       float64 `json:"close"`
	EMA50       float64 `json:"ema50"`
	EMA200      float64 `json:"ema200"`
	RSI         float64 `json:"rsi"`
	MACD        float64 `json:"macd"`
	MACDSignal  float64 `json:"macdSignal"`
	VolumeRatio float64 `json:"volumeRatio"`
	ATR         float64 `json:"atr"`
	ADX         float64 `json:"adx"`
	Recommend   float64 `json:"recommend"`
	BBLower     float64 `json:"bbLower"`
	BBUpper     float64 `json:"bbUpper"`
	PivotR1     float64 `json:"pivotR1"`
}

// computeStockScore calculates the stock score (0-100) based on indicators
func computeStockScore(values map[string]interface{}) (int, ScoreComponents) {
	score := 0
	components := ScoreComponents{}

	// Get indicator values
	close := screener.GetFloatFromInterface(values, "close")
	ema50 := screener.GetFloatFromInterface(values, "EMA50")
	ema200 := screener.GetFloatFromInterface(values, "EMA200")
	rsi := screener.GetFloatFromInterface(values, "RSI")
	macd := screener.GetFloatFromInterface(values, "MACD.macd")
	macdSignal := screener.GetFloatFromInterface(values, "MACD.signal")
	volumeRatio := 1.0
	if vol := screener.GetFloatFromInterface(values, "volume.SMA20"); vol > 0 {
		volumeRatio = screener.GetFloatFromInterface(values, "volume") / vol
	}
	adx := screener.GetFloatFromInterface(values, "ADX")
	bbLower := screener.GetFloatFromInterface(values, "BB.lower")
	bbUpper := screener.GetFloatFromInterface(values, "BB.upper")
	recommend := screener.GetFloatFromInterface(values, "Recommend.All")

	// EMA alignment (20 pts)
	if close > ema50 {
		components.EMAAlignment += 10
	}
	if close > ema200 {
		components.EMAAlignment += 10
	}
	score += components.EMAAlignment

	// RSI score (20 pts)
	if rsi > 50 {
		components.RSIScore += 10
	}
	if rsi >= 40 && rsi <= 70 {
		components.RSIScore += 10
	}
	score += components.RSIScore

	// MACD score (20 pts)
	if macd > macdSignal {
		components.MACDScore += 10
	}
	// Change > 0 (assuming positive MACD is bullish)
	if macd > 0 {
		components.MACDScore += 10
	}
	score += components.MACDScore

	// Volume + Bollinger Bands (20 pts)
	if volumeRatio > 1.0 {
		components.VolumeScore += 10
	}
	if close > bbLower && close < bbUpper {
		components.VolumeScore += 10
	}
	score += components.VolumeScore

	// ADX + Recommend (20 pts)
	if adx > 20 {
		components.ADXScore += 10
	}
	if recommend > 0 {
		components.ADXScore += 10
	}
	score += components.ADXScore

	return score, components
}

// gradeScore returns a letter grade for the stock score
func gradeScore(score int) string {
	if score >= 80 {
		return "A"
	}
	if score >= 70 {
		return "B"
	}
	if score >= 60 {
		return "C"
	}
	if score >= 50 {
		return "D"
	}
	return "F"
}

// computeTradeQuality calculates trade quality (0-100)
func computeTradeQuality(score int, rr2, volumeRatio, stopLossPct, rsi float64) (int, QualityBreakdown) {
	quality := 0
	breakdown := QualityBreakdown{}

	// Structure (30): score >= 70 → 30, >= 50 → 20, else 10
	if score >= 70 {
		breakdown.Structure = 30
	} else if score >= 50 {
		breakdown.Structure = 20
	} else {
		breakdown.Structure = 10
	}
	quality += breakdown.Structure

	// Reward/Risk (30): rr2 >= 2 → 30, >= 1.5 → 20, else 10
	if rr2 >= 2.0 {
		breakdown.RewardRisk = 30
	} else if rr2 >= 1.5 {
		breakdown.RewardRisk = 20
	} else {
		breakdown.RewardRisk = 10
	}
	quality += breakdown.RewardRisk

	// Volume (20): volumeRatio >= 1.5 → 20, >= 1 → 10, else 0
	if volumeRatio >= 1.5 {
		breakdown.Volume = 20
	} else if volumeRatio >= 1.0 {
		breakdown.Volume = 10
	} else {
		breakdown.Volume = 0
	}
	quality += breakdown.Volume

	// Stop Size (10): < 5% → 10, < 10% → 5, else 0
	if stopLossPct < 5.0 {
		breakdown.StopSize = 10
	} else if stopLossPct < 10.0 {
		breakdown.StopSize = 5
	} else {
		breakdown.StopSize = 0
	}
	quality += breakdown.StopSize

	// Liquidity (10): RSI between 30-70 → 10, else 0
	if rsi >= 30 && rsi <= 70 {
		breakdown.Liquidity = 10
	} else {
		breakdown.Liquidity = 0
	}
	quality += breakdown.Liquidity

	return quality, breakdown
}

// getRecommendation returns trade recommendation based on scores
func getRecommendation(score, quality int, rr2 float64) string {
	if score >= 70 && quality >= 65 && rr2 >= 2.0 {
		return "QUALIFIED"
	}
	if score >= 70 && quality >= 50 {
		return "CONDITIONAL"
	}
	if score >= 55 {
		return "WATCHLIST"
	}
	return "AVOID"
}

// RunTradePlan generates a complete trade plan with entry, targets, and quality assessment
func RunTradePlan(cfg *config.Config, symbol, exchange, timeframe string) error {
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

	// Create HTTP client and TradingView client
	httpClient := client.NewHTTPClient(cfg)
	tvClient := client.NewTradingViewClient(httpClient)

	// Fetch analysis data
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
	close := screener.GetFloatFromInterface(values, "close")
	atr := screener.GetFloatFromInterface(values, "ATR")
	pivotR1 := screener.GetFloatFromInterface(values, "Pivot.M.Classic.R1")
	rsi := screener.GetFloatFromInterface(values, "RSI")
	volumeRatio := 1.0
	if vol := screener.GetFloatFromInterface(values, "volume.SMA20"); vol > 0 {
		volumeRatio = screener.GetFloatFromInterface(values, "volume") / vol
	}

	// Compute stock score
	score, components := computeStockScore(values)
	grade := gradeScore(score)

	// Compute trade setup
	entry := close
	stopLoss := close - (2 * atr)
	target1 := close + (2 * atr)
	target2 := close + (4 * atr)

	// target3: use Pivot.R1 if > close, else close + 6*ATR
	target3 := close + (6 * atr)
	if pivotR1 > close {
		target3 = pivotR1
	}

	// Compute risk/reward ratios
	risk := entry - stopLoss
	rr1 := 1.0
	rr2 := 2.0
	rr3 := 3.0
	if risk > 0 {
		rr1 = (target1 - entry) / risk
		rr2 = (target2 - entry) / risk
		rr3 = (target3 - entry) / risk
	}

	setup := TradeSetup{
		Entry:    entry,
		StopLoss: stopLoss,
		Target1:  target1,
		Target2:  target2,
		Target3:  target3,
		RR1:      rr1,
		RR2:      rr2,
		RR3:      rr3,
	}

	// Compute trade quality
	stopLossPct := (entry - stopLoss) / entry * 100
	quality, breakdown := computeTradeQuality(score, rr2, volumeRatio, stopLossPct, rsi)

	// Get recommendation
	recommendation := getRecommendation(score, quality, rr2)

	// Build output
	output := TradePlanOutput{
		Symbol:         ticker,
		Exchange:       exchange,
		Timeframe:      timeframe,
		StockScore:     score,
		Grade:          grade,
		TradeQuality:   quality,
		Setup:          setup,
		Recommendation: recommendation,
		Details: TradeDetails{
			ScoreComponents:  components,
			QualityBreakdown: breakdown,
			Indicators: IndicatorValues{
				Close:       close,
				EMA50:       screener.GetFloatFromInterface(values, "EMA50"),
				EMA200:      screener.GetFloatFromInterface(values, "EMA200"),
				RSI:         rsi,
				MACD:        screener.GetFloatFromInterface(values, "MACD.macd"),
				MACDSignal:  screener.GetFloatFromInterface(values, "MACD.signal"),
				VolumeRatio: volumeRatio,
				ATR:         atr,
				ADX:         screener.GetFloatFromInterface(values, "ADX"),
				Recommend:   screener.GetFloatFromInterface(values, "Recommend.All"),
				BBLower:     screener.GetFloatFromInterface(values, "BB.lower"),
				BBUpper:     screener.GetFloatFromInterface(values, "BB.upper"),
				PivotR1:     pivotR1,
			},
		},
		Timestamp: time.Now().UTC(),
	}

	return utils.PrintJSON(output)
}
