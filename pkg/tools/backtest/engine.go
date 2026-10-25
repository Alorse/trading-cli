package backtest

import (
	"math"

	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

// Trade represents a single trade (entry and exit).
type Trade struct {
	EntryIndex int     `json:"entryIndex"`
	ExitIndex  int     `json:"exitIndex"`
	EntryPrice float64 `json:"entryPrice"`
	ExitPrice  float64 `json:"exitPrice"`
	ReturnPct  float64 `json:"returnPct"`
	IsWin      bool    `json:"isWin"`
	ExitReason string  `json:"exitReason,omitempty"` // "stop", "target", "signal", "end"
}

// BacktestResult holds comprehensive backtest metrics and results.
type BacktestResult struct {
	Strategy      string    `json:"strategy"`
	Symbol        string    `json:"symbol"`
	Period        string    `json:"period"`
	TotalTrades   int       `json:"totalTrades"`
	WinningTrades int       `json:"winningTrades"`
	LosingTrades  int       `json:"losingTrades"`
	WinRate       float64   `json:"winRate"`
	FinalCapital  float64   `json:"finalCapital"`
	TotalReturn   float64   `json:"totalReturn"`
	AvgGain       float64   `json:"avgGain"`
	AvgLoss       float64   `json:"avgLoss"`
	MaxDrawdown   float64   `json:"maxDrawdown"`
	ProfitFactor  float64   `json:"profitFactor"`
	SharpeRatio   float64   `json:"sharpeRatio"`
	CalmarRatio   float64   `json:"calmarRatio"`
	Expectancy    float64   `json:"expectancy"`
	BestTrade     float64   `json:"bestTrade"`
	WorstTrade    float64   `json:"worstTrade"`
	Trades        []Trade   `json:"trades,omitempty"`
	EquityCurve   []float64 `json:"equityCurve,omitempty"`
}

// Strategy interface that all strategies must implement.
type Strategy interface {
	Name() string
	GenerateSignals(candles []client.YahooOHLCV) []int // +1 buy, -1 sell, 0 hold
}

// StrategyWithStops is an optional interface a Strategy may implement
// to provide ATR-based stop loss and take profit levels at entry time.
// If sl or tp is 0, the engine treats that level as disabled.
type StrategyWithStops interface {
	Strategy
	EntryStops(candles []client.YahooOHLCV, entryIdx int) (sl float64, tp float64)
}

// RunBacktest executes a backtest strategy on OHLCV candles.
// Commission and slippage are percentages (e.g. 0.1 = 0.1%).
// Cost per trade = (commissionPct + slippagePct) / 100 * 2
func RunBacktest(symbol, strategyName string, candles []client.YahooOHLCV, strategy Strategy, initialCapital, commissionPct, slippagePct float64, includeTradeLog, includeEquityCurve bool) BacktestResult {
	result := BacktestResult{
		Strategy: strategyName,
		Symbol:   symbol,
		Period:   "",
	}

	if len(candles) == 0 {
		result.FinalCapital = initialCapital
		return result
	}

	// Generate signals
	signals := strategy.GenerateSignals(candles)
	if len(signals) != len(candles) {
		// Invalid signals, return empty result
		result.FinalCapital = initialCapital
		return result
	}

	// Calculate transaction cost
	costPerTrade := (commissionPct + slippagePct) / 100 * 2

	// Check if strategy provides SL/TP via the optional interface.
	stopsProvider, hasStops := strategy.(StrategyWithStops)

	// Execute trades
	capital := initialCapital
	equityCurve := []float64{capital}
	var trades []Trade
	var inPosition bool
	var entryPrice float64
	var entryIndex int
	var currentSL, currentTP float64
	var dailyReturns []float64

	closeTrade := func(exitIdx int, exitPrice float64, reason string) {
		inPosition = false
		returnBeforeCost := (exitPrice - entryPrice) / entryPrice
		transactionCost := costPerTrade
		returnAfterCost := returnBeforeCost - transactionCost

		returnPct := returnAfterCost * 100
		tradeGain := capital * returnAfterCost
		capital += tradeGain

		trades = append(trades, Trade{
			EntryIndex: entryIndex,
			ExitIndex:  exitIdx,
			EntryPrice: entryPrice,
			ExitPrice:  exitPrice,
			ReturnPct:  returnPct,
			IsWin:      returnPct > 0,
			ExitReason: reason,
		})
		dailyReturns = append(dailyReturns, returnAfterCost)
		equityCurve = append(equityCurve, capital)

		currentSL = 0
		currentTP = 0
	}

	for i := 0; i < len(candles); i++ {
		signal := signals[i]
		closePrice := candles[i].Close

		// If in position, check SL/TP intrabar first (skip on the very entry bar).
		if inPosition && i > entryIndex {
			high := candles[i].High
			low := candles[i].Low

			// Stop loss has priority over take profit if both touched in same bar (conservative).
			if currentSL > 0 && low <= currentSL {
				closeTrade(i, currentSL, "stop")
			} else if currentTP > 0 && high >= currentTP {
				closeTrade(i, currentTP, "target")
			}
		}

		// SELL signal (only if still in position after SL/TP check).
		if signal == -1 && inPosition {
			closeTrade(i, closePrice, "signal")
		}

		// BUY signal
		if signal == 1 && !inPosition {
			inPosition = true
			entryPrice = closePrice
			entryIndex = i
			currentSL = 0
			currentTP = 0
			if hasStops {
				sl, tp := stopsProvider.EntryStops(candles, i)
				if sl > 0 && sl < entryPrice {
					currentSL = sl
				}
				if tp > 0 && tp > entryPrice {
					currentTP = tp
				}
			}
		}
	}

	// Close any open position at last candle
	if inPosition {
		closeTrade(len(candles)-1, candles[len(candles)-1].Close, "end")
	}

	result.FinalCapital = capital
	result.TotalTrades = len(trades)
	result.TotalReturn = ((capital - initialCapital) / initialCapital) * 100

	// Calculate win/loss statistics
	if len(trades) > 0 {
		var gainSum, lossSum, gainCount, lossCount float64
		var gains, losses []float64
		bestReturn := math.Inf(-1)
		worstReturn := math.Inf(1)

		for _, trade := range trades {
			if trade.IsWin {
				result.WinningTrades++
				gainSum += trade.ReturnPct
				gainCount++
				gains = append(gains, trade.ReturnPct)
			} else {
				result.LosingTrades++
				lossSum += trade.ReturnPct
				lossCount++
				losses = append(losses, trade.ReturnPct)
			}

			if trade.ReturnPct > bestReturn {
				bestReturn = trade.ReturnPct
			}
			if trade.ReturnPct < worstReturn {
				worstReturn = trade.ReturnPct
			}
		}

		result.WinRate = (float64(result.WinningTrades) / float64(result.TotalTrades)) * 100
		result.AvgGain = utils.SafeDivide(gainSum, gainCount)
		result.AvgLoss = utils.SafeDivide(lossSum, lossCount)
		result.BestTrade = bestReturn
		result.WorstTrade = worstReturn

		// Calculate Profit Factor
		if len(losses) > 0 {
			totalGains := 0.0
			totalLosses := 0.0
			for _, g := range gains {
				totalGains += g
			}
			for _, l := range losses {
				totalLosses += utils.Abs(l)
			}
			result.ProfitFactor = utils.SafeDivide(totalGains, totalLosses)
		}

		// Calculate Expectancy
		result.Expectancy = (result.WinRate/100)*result.AvgGain - ((1 - (result.WinRate / 100)) * utils.Abs(result.AvgLoss))
	}

	// Calculate Max Drawdown
	result.MaxDrawdown = calculateMaxDrawdown(equityCurve)

	// Calculate Sharpe Ratio
	if len(dailyReturns) > 0 {
		result.SharpeRatio = calculateSharpeRatio(dailyReturns)
	}

	// Calculate Calmar Ratio
	if result.MaxDrawdown != 0 && result.MaxDrawdown < 0 {
		result.CalmarRatio = result.TotalReturn / utils.Abs(result.MaxDrawdown)
	}

	if includeTradeLog {
		result.Trades = trades
	}
	if includeEquityCurve {
		result.EquityCurve = equityCurve
	}

	return result
}

// calculateMaxDrawdown computes peak-to-trough percentage decline.
func calculateMaxDrawdown(equityCurve []float64) float64 {
	if len(equityCurve) == 0 {
		return 0
	}

	maxDrawdown := 0.0
	peak := equityCurve[0]

	for i := 1; i < len(equityCurve); i++ {
		if equityCurve[i] > peak {
			peak = equityCurve[i]
		}
		if peak != 0 {
			dd := ((equityCurve[i] - peak) / peak) * 100
			if dd < maxDrawdown {
				maxDrawdown = dd
			}
		}
	}

	return maxDrawdown
}

// calculateSharpeRatio computes annualized Sharpe ratio.
// Assumes daily returns, risk-free rate = 4% annual = 4/252 per day.
// Sharpe = (avg_return - riskFreeDaily) / stddev * sqrt(252)
func calculateSharpeRatio(dailyReturns []float64) float64 {
	if len(dailyReturns) == 0 {
		return 0
	}

	riskFreeDaily := 0.04 / 252

	// Calculate mean return
	sum := 0.0
	for _, r := range dailyReturns {
		sum += r
	}
	meanReturn := sum / float64(len(dailyReturns))

	// Calculate standard deviation
	variance := 0.0
	for _, r := range dailyReturns {
		diff := r - meanReturn
		variance += diff * diff
	}
	variance /= float64(len(dailyReturns))
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	return ((meanReturn - riskFreeDaily) / stdDev) * math.Sqrt(252)
}
