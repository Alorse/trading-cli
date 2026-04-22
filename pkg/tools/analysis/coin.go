package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/tools/screener"
	"github.com/alorse/trading-cli/pkg/utils"
)

// CoinAnalysisOutput represents the comprehensive coin analysis output
type CoinAnalysisOutput struct {
	Symbol             string              `json:"symbol"`
	Exchange           string              `json:"exchange"`
	Timeframe          string              `json:"timeframe"`
	Price              PriceData           `json:"price"`
	RSI                RSIData             `json:"rsi"`
	MACD               MACDData            `json:"macd"`
	SMA                map[string]float64  `json:"sma"`
	EMA                map[string]float64  `json:"ema"`
	BollingerBands     BollingerBandsData  `json:"bollingerBands"`
	ATR                float64             `json:"atr"`
	ADX                float64             `json:"adx"`
	Volume             VolumeData          `json:"volume"`
	Stochastic         StochasticData      `json:"stochastic"`
	CCI                CCIData             `json:"cci"`
	WilliamsR          WilliamsRData       `json:"williamsR"`
	AwesomeOscillator  float64             `json:"awesomeOscillator"`
	Momentum           MomentumData        `json:"momentum"`
	ParabolicSAR       float64             `json:"parabolicSAR"`
	Ichimoku           IchimokuData        `json:"ichimoku"`
	HullMA             float64             `json:"hullMA"`
	StochasticRSI      StochasticRSIData   `json:"stochasticRSI"`
	UltimateOscillator float64             `json:"ultimateOscillator"`
	VWAP               float64             `json:"vwap"`
	VWMA               float64             `json:"vwma"`
	Recommendation     RecommendationData  `json:"recommendation"`
	MarketStructure    MarketStructureData `json:"marketStructure"`
	Timestamp          string              `json:"timestamp"`
}

type PriceData struct {
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	ChangePercent float64 `json:"changePercent"`
	Volume        float64 `json:"volume"`
}

type RSIData struct {
	Value    float64 `json:"value"`
	Signal   string  `json:"signal"`
	Previous float64 `json:"previous"`
}

type MACDData struct {
	Line      float64 `json:"line"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

type BollingerBandsData struct {
	Upper    float64 `json:"upper"`
	Middle   float64 `json:"middle"`
	Lower    float64 `json:"lower"`
	Width    float64 `json:"width"`
	Position string  `json:"position"`
}

type VolumeData struct {
	Current float64 `json:"current"`
	Avg20   float64 `json:"avg20"`
	Ratio   float64 `json:"ratio"`
}

type StochasticData struct {
	K float64 `json:"k"`
	D float64 `json:"d"`
}

type RecommendationData struct {
	All   float64 `json:"all"`
	MA    float64 `json:"ma"`
	Other float64 `json:"other"`
}

type MarketStructureData struct {
	Trend             string  `json:"trend"`
	TrendScore        int     `json:"trendScore"`
	MomentumAlignment float64 `json:"momentumAlignment"`
}

type CCIData struct {
	Value  float64 `json:"value"`
	Signal string  `json:"signal"`
}

type WilliamsRData struct {
	Value float64 `json:"value"`
}

type MomentumData struct {
	Value float64 `json:"value"`
}

type IchimokuData struct {
	BaseLine float64 `json:"baseLine"`
}

type StochasticRSIData struct {
	K float64 `json:"k"`
}

// getFloat safely extracts a float64 from an interface{} map value
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

// computeRSISignal determines the RSI signal based on value
func computeRSISignal(rsi float64) string {
	if rsi > 70 {
		return "overbought"
	}
	if rsi < 30 {
		return "oversold"
	}
	return "neutral"
}

// computeBBPosition determines position relative to Bollinger Bands
func computeBBPosition(close, upper, lower float64) string {
	if close > upper {
		return "above"
	}
	if close < lower {
		return "below"
	}
	return "inside"
}

// computeTrendScore calculates trend score based on multiple conditions (0-5)
func computeTrendScore(close, sma20, sma50, ema50, ema200, rsi float64) int {
	score := 0

	// Condition 1: close > SMA20
	if close > sma20 {
		score++
	}

	// Condition 2: close > SMA50
	if close > sma50 {
		score++
	}

	// Condition 3: close > EMA50
	if close > ema50 {
		score++
	}

	// Condition 4: close > EMA200
	if close > ema200 {
		score++
	}

	// Condition 5: RSI > 50
	if rsi > 50 {
		score++
	}

	return score
}

// computeTrendFloat determines trend based on EMA levels
func computeTrendFloat(close, ema50, ema200 float64) string {
	if close > ema50 && ema50 > ema200 {
		return "bullish"
	}
	if close < ema50 && ema50 < ema200 {
		return "bearish"
	}
	return "neutral"
}

func computeCCISignal(cci float64) string {
	if cci > 100 {
		return "overbought"
	}
	if cci < -100 {
		return "oversold"
	}
	return "neutral"
}

// BuildCoinAnalysisOutput constructs the analysis output from raw values
func BuildCoinAnalysisOutput(symbol, exchange, timeframe string, values map[string]interface{}) *CoinAnalysisOutput {
	close := getFloat(values, "close")
	open := getFloat(values, "open")
	high := getFloat(values, "high")
	low := getFloat(values, "low")
	volume := getFloat(values, "volume")
	avgVolume := getFloat(values, "average_volume_10d_calc")
	change := getFloat(values, "change")

	rsi := getFloat(values, "RSI")
	rsiPrev := getFloat(values, "RSI[1]")

	macdLine := getFloat(values, "MACD.macd")
	macdSignal := getFloat(values, "MACD.signal")
	macdHistogram := macdLine - macdSignal

	sma10 := getFloat(values, "SMA10")
	sma20 := getFloat(values, "SMA20")
	sma50 := getFloat(values, "SMA50")
	sma100 := getFloat(values, "SMA100")
	sma200 := getFloat(values, "SMA200")

	ema9 := getFloat(values, "EMA9")
	ema20 := getFloat(values, "EMA20")
	ema50 := getFloat(values, "EMA50")
	ema100 := getFloat(values, "EMA100")
	ema200 := getFloat(values, "EMA200")

	bbUpper := getFloat(values, "BB.upper")
	bbLower := getFloat(values, "BB.lower")

	atr := getFloat(values, "ATR")
	adx := getFloat(values, "ADX")

	stochK := getFloat(values, "Stoch.K")
	stochD := getFloat(values, "Stoch.D")

	cci := getFloat(values, "CCI20")
	williamsR := getFloat(values, "W.R")
	ao := getFloat(values, "AO")
	mom := getFloat(values, "Mom")
	psar := getFloat(values, "P.SAR")
	ichimokuBase := getFloat(values, "Ichimoku.BLine")
	hullMA := getFloat(values, "HullMA9")
	stochRSIK := getFloat(values, "Stoch.RSI.K")
	uo := getFloat(values, "UO")
	vwap := getFloat(values, "VWAP")
	vwma := getFloat(values, "VWMA")

	recAll := getFloat(values, "Recommend.All")
	recMA := getFloat(values, "Recommend.MA")
	recOther := getFloat(values, "Recommend.Other")

	volumeRatio := 1.0
	if avgVolume > 0 {
		volumeRatio = volume / avgVolume
	}

	bbWidth := 0.0
	if sma20 > 0 {
		bbWidth = (bbUpper - bbLower) / sma20
	}

	trendScore := computeTrendScore(close, sma20, sma50, ema50, ema200, rsi)
	trend := computeTrendFloat(close, ema50, ema200)

	output := &CoinAnalysisOutput{
		Symbol:    symbol,
		Exchange:  exchange,
		Timeframe: timeframe,
		Price: PriceData{
			Open:          open,
			High:          high,
			Low:           low,
			Close:         close,
			ChangePercent: change,
			Volume:        volume,
		},
		RSI: RSIData{
			Value:    rsi,
			Signal:   computeRSISignal(rsi),
			Previous: rsiPrev,
		},
		MACD: MACDData{
			Line:      macdLine,
			Signal:    macdSignal,
			Histogram: macdHistogram,
		},
		SMA: map[string]float64{
			"10":  sma10,
			"20":  sma20,
			"50":  sma50,
			"100": sma100,
			"200": sma200,
		},
		EMA: map[string]float64{
			"9":   ema9,
			"20":  ema20,
			"50":  ema50,
			"100": ema100,
			"200": ema200,
		},
		BollingerBands: BollingerBandsData{
			Upper:    bbUpper,
			Middle:   sma20,
			Lower:    bbLower,
			Width:    bbWidth,
			Position: computeBBPosition(close, bbUpper, bbLower),
		},
		ATR: atr,
		ADX: adx,
		Volume: VolumeData{
			Current: volume,
			Avg20:   avgVolume,
			Ratio:   volumeRatio,
		},
		Stochastic: StochasticData{
			K: stochK,
			D: stochD,
		},
		CCI: CCIData{
			Value:  cci,
			Signal: computeCCISignal(cci),
		},
		WilliamsR: WilliamsRData{
			Value: williamsR,
		},
		AwesomeOscillator: ao,
		Momentum: MomentumData{
			Value: mom,
		},
		ParabolicSAR: psar,
		Ichimoku: IchimokuData{
			BaseLine: ichimokuBase,
		},
		HullMA: hullMA,
		StochasticRSI: StochasticRSIData{
			K: stochRSIK,
		},
		UltimateOscillator: uo,
		VWAP:               vwap,
		VWMA:               vwma,
		Recommendation: RecommendationData{
			All:   recAll,
			MA:    recMA,
			Other: recOther,
		},
		MarketStructure: MarketStructureData{
			Trend:             trend,
			TrendScore:        trendScore,
			MomentumAlignment: float64(trendScore) / 5.0,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return output
}

// RunCoinAnalysis performs comprehensive analysis on a single cryptocurrency
func RunCoinAnalysis(cfg *config.Config, symbol, exchange, timeframe string) error {
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

	// Build and output analysis
	output := BuildCoinAnalysisOutput(ticker, exchange, timeframe, results[0].Values)
	return utils.PrintJSON(output)
}
