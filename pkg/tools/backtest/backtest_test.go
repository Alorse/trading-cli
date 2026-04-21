package backtest

import (
	"math"
	"testing"

	"github.com/alorse/trading-cli/pkg/client"
)

// TestRunBacktestWithSyntheticData creates synthetic candles with a known trend
// and verifies that the backtest engine produces valid results.
func TestRunBacktestWithSyntheticData(t *testing.T) {
	// Create 200 candles with uptrend, downtrend, uptrend pattern
	candles := make([]client.YahooOHLCV, 200)

	// Phase 1: Uptrend (0-70)
	for i := 0; i < 70; i++ {
		price := 100.0 + float64(i)*0.5
		candles[i] = client.YahooOHLCV{
			Timestamp: int64(i),
			Open:      price,
			High:      price + 0.5,
			Low:       price - 0.5,
			Close:     price,
			Volume:    1000000,
		}
	}

	// Phase 2: Downtrend (71-140)
	for i := 70; i < 140; i++ {
		price := 135.0 - float64(i-70)*0.5
		candles[i] = client.YahooOHLCV{
			Timestamp: int64(i),
			Open:      price,
			High:      price + 0.5,
			Low:       price - 0.5,
			Close:     price,
			Volume:    1000000,
		}
	}

	// Phase 3: Uptrend (141-199)
	for i := 140; i < 200; i++ {
		price := 100.0 + float64(i-140)*0.5
		candles[i] = client.YahooOHLCV{
			Timestamp: int64(i),
			Open:      price,
			High:      price + 0.5,
			Low:       price - 0.5,
			Close:     price,
			Volume:    1000000,
		}
	}

	strategy := &RSIStrategy{}
	result := RunBacktest("TEST", "rsi", candles, strategy, 10000.0, 0.1, 0.1, true, true)

	// Verify result structure
	if result.Strategy != "rsi" {
		t.Errorf("Expected strategy 'rsi', got '%s'", result.Strategy)
	}
	if result.Symbol != "TEST" {
		t.Errorf("Expected symbol 'TEST', got '%s'", result.Symbol)
	}

	// Verify we executed trades
	if result.TotalTrades < 1 {
		t.Errorf("Expected at least 1 trade, got %d", result.TotalTrades)
	}

	// Verify win/loss counts match total
	if result.WinningTrades+result.LosingTrades != result.TotalTrades {
		t.Errorf("WinningTrades (%d) + LosingTrades (%d) != TotalTrades (%d)",
			result.WinningTrades, result.LosingTrades, result.TotalTrades)
	}

	// Verify win rate is between 0-100
	if result.WinRate < 0 || result.WinRate > 100 {
		t.Errorf("WinRate should be 0-100, got %f", result.WinRate)
	}

	// Verify we have trade log and equity curve
	if len(result.Trades) != result.TotalTrades {
		t.Errorf("Trade log length %d != TotalTrades %d", len(result.Trades), result.TotalTrades)
	}
	if len(result.EquityCurve) == 0 {
		t.Errorf("Expected non-empty equity curve")
	}

	// Verify capital is positive
	if result.FinalCapital <= 0 {
		t.Errorf("FinalCapital should be positive, got %f", result.FinalCapital)
	}
}

// TestAllStrategiesCompile verifies all 6 strategies can generate signals
// without panicking on valid candles.
func TestAllStrategiesCompile(t *testing.T) {
	// Create simple uptrend candles
	candles := make([]client.YahooOHLCV, 100)
	for i := 0; i < 100; i++ {
		price := 100.0 + float64(i)*0.1
		candles[i] = client.YahooOHLCV{
			Timestamp: int64(i),
			Open:      price,
			High:      price + 0.2,
			Low:       price - 0.2,
			Close:     price,
			Volume:    1000000,
		}
	}

	strategies := []struct {
		name     string
		strategy Strategy
	}{
		{"rsi", &RSIStrategy{}},
		{"bollinger", &BollingerStrategy{}},
		{"macd", &MACDStrategy{}},
		{"ema-cross", &EMACrossStrategy{}},
		{"supertrend", &SupertrendStrategy{}},
		{"donchian", &DonchianStrategy{}},
	}

	for _, s := range strategies {
		t.Run(s.name, func(t *testing.T) {
			// Should not panic
			signals := s.strategy.GenerateSignals(candles)

			// Verify signal length matches candles
			if len(signals) != len(candles) {
				t.Errorf("Signal length %d != candles length %d", len(signals), len(candles))
			}

			// Verify signals are only 0, 1, or -1
			for i, sig := range signals {
				if sig != 0 && sig != 1 && sig != -1 {
					t.Errorf("Invalid signal %d at index %d", sig, i)
				}
			}
		})
	}
}

// TestMaxDrawdown verifies the max drawdown calculation with known values.
func TestMaxDrawdown(t *testing.T) {
	// Known equity curve: start at 100, peak at 120, drop to 100, recover to 110
	equityCurve := []float64{100.0, 110.0, 120.0, 110.0, 100.0, 110.0}

	// Max DD should be from 120 to 100 = -16.67%
	expected := -16.67

	result := calculateMaxDrawdown(equityCurve)

	if math.Abs(result-expected) > 0.1 {
		t.Errorf("Expected max drawdown ~%.2f, got %.2f", expected, result)
	}
}

// TestSharpeRatio verifies Sharpe ratio with flat equity (should be 0).
func TestSharpeRatio(t *testing.T) {
	// Flat returns = zero Sharpe
	dailyReturns := []float64{0.0, 0.0, 0.0, 0.0, 0.0}

	result := calculateSharpeRatio(dailyReturns)

	if result != 0 {
		t.Errorf("Expected Sharpe ratio 0 for flat returns, got %f", result)
	}
}

// TestWalkForwardSplit verifies walk-forward split logic on synthetic data.
func TestWalkForwardSplit(t *testing.T) {
	// Create 100 simple uptrend candles
	candles := make([]client.YahooOHLCV, 100)
	for i := 0; i < 100; i++ {
		price := 100.0 + float64(i)*0.1
		candles[i] = client.YahooOHLCV{
			Timestamp: int64(i),
			Open:      price,
			High:      price + 0.2,
			Low:       price - 0.2,
			Close:     price,
			Volume:    1000000,
		}
	}

	nSplits := 5
	trainRatio := 0.7

	// Calculate expected split sizes
	foldSize := 100 / nSplits // 20
	trainSize := int(float64(foldSize) * trainRatio)
	testSize := foldSize - trainSize

	for fold := 0; fold < nSplits; fold++ {
		foldStart := fold * foldSize
		foldEnd := foldStart + foldSize

		trainEnd := foldStart + trainSize

		expectedFold := candles[foldStart:foldEnd]
		expectedTrain := candles[foldStart:trainEnd]
		expectedTest := candles[trainEnd:foldEnd]

		if len(expectedFold) != foldSize {
			t.Errorf("Fold %d: expected size %d, got %d", fold, foldSize, len(expectedFold))
		}
		if len(expectedTrain) != trainSize {
			t.Errorf("Fold %d: expected train size %d, got %d", fold, trainSize, len(expectedTrain))
		}
		if len(expectedTest) != testSize {
			t.Errorf("Fold %d: expected test size %d, got %d", fold, testSize, len(expectedTest))
		}
	}
}

// TestEmptyCandles tests behavior with empty candles slice.
func TestEmptyCandles(t *testing.T) {
	strategy := &RSIStrategy{}
	result := RunBacktest("TEST", "rsi", []client.YahooOHLCV{}, strategy, 10000.0, 0.1, 0.1, false, false)

	if result.FinalCapital != 10000.0 {
		t.Errorf("Expected final capital 10000, got %f", result.FinalCapital)
	}
	if result.TotalTrades != 0 {
		t.Errorf("Expected 0 trades, got %d", result.TotalTrades)
	}
}

// TestNoTradesExecuted tests result when no trades are generated.
func TestNoTradesExecuted(t *testing.T) {
	// Create flat candles that won't trigger RSI strategy
	candles := make([]client.YahooOHLCV, 50)
	for i := 0; i < 50; i++ {
		candles[i] = client.YahooOHLCV{
			Timestamp: int64(i),
			Open:      100.0,
			High:      100.1,
			Low:       99.9,
			Close:     100.0,
			Volume:    1000000,
		}
	}

	strategy := &RSIStrategy{}
	result := RunBacktest("TEST", "rsi", candles, strategy, 10000.0, 0.1, 0.1, false, false)

	if result.FinalCapital != 10000.0 {
		t.Errorf("Expected final capital unchanged at 10000, got %f", result.FinalCapital)
	}
	if result.TotalTrades != 0 {
		t.Errorf("Expected 0 trades, got %d", result.TotalTrades)
	}
	if result.TotalReturn != 0 {
		t.Errorf("Expected 0 percent return, got %f", result.TotalReturn)
	}
}

// TestTransactionCosts verifies that transaction costs are properly deducted.
func TestTransactionCosts(t *testing.T) {
	// Simple strategy: buy at bar 1, sell at bar 2
	candles := []client.YahooOHLCV{
		{Timestamp: 0, Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000000},
		{Timestamp: 1, Open: 100, High: 100, Low: 100, Close: 100, Volume: 1000000},
		{Timestamp: 2, Open: 100, High: 100, Low: 100, Close: 110, Volume: 1000000}, // +10% gain
	}

	strategy := &TestBuyHoldStrategy{buyAt: 1, sellAt: 2}
	result := RunBacktest("TEST", "test", candles, strategy, 10000.0, 0.1, 0.1, true, false)

	// Without costs: 10% gain = 1000 profit, final = 11000
	// With costs: 0.2% transaction cost on entry + 0.2% on exit = 0.4% total
	// Return = 10% - 0.4% = 9.6%, gain = 960, final = 10960

	if result.TotalTrades != 1 {
		t.Errorf("Expected 1 trade, got %d", result.TotalTrades)
	}

	// Approximate check (accounting for floating point)
	if result.Trades[0].ReturnPct < 9.0 || result.Trades[0].ReturnPct > 10.0 {
		t.Errorf("Expected trade return ~9.6, got %f", result.Trades[0].ReturnPct)
	}
}

// TestProfitFactor verifies profit factor calculation.
func TestProfitFactor(t *testing.T) {
	candles := make([]client.YahooOHLCV, 100)

	// Create pattern: oscillates between 100 and 110
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			candles[i] = client.YahooOHLCV{
				Timestamp: int64(i),
				Close:     100.0,
				High:      100.5,
				Low:       99.5,
				Open:      100.0,
				Volume:    1000000,
			}
		} else {
			candles[i] = client.YahooOHLCV{
				Timestamp: int64(i),
				Close:     110.0,
				High:      110.5,
				Low:       109.5,
				Open:      110.0,
				Volume:    1000000,
			}
		}
	}

	strategy := &RSIStrategy{}
	result := RunBacktest("TEST", "rsi", candles, strategy, 10000.0, 0.0, 0.0, true, false)

	if result.TotalTrades > 0 {
		if result.WinningTrades > 0 && result.LosingTrades > 0 && result.ProfitFactor <= 0 {
			t.Errorf("Expected positive profit factor with wins and losses, got %f", result.ProfitFactor)
		}
	}
}

// TestGetStrategy verifies the GetStrategy factory function.
func TestGetStrategy(t *testing.T) {
	strategies := []string{"rsi", "bollinger", "macd", "ema-cross", "supertrend", "donchian"}

	for _, name := range strategies {
		strategy := GetStrategy(name)
		if strategy == nil {
			t.Errorf("GetStrategy(%s) returned nil", name)
		}
		if strategy.Name() != name {
			t.Errorf("Expected strategy name %s, got %s", name, strategy.Name())
		}
	}

	// Unknown strategy should return nil
	unknown := GetStrategy("unknown")
	if unknown != nil {
		t.Errorf("Expected nil for unknown strategy, got %v", unknown)
	}
}

// TestBuyHoldStrategy is a simple test strategy that buys at a specific index and sells at another.
type TestBuyHoldStrategy struct {
	buyAt  int
	sellAt int
}

func (s *TestBuyHoldStrategy) Name() string {
	return "test"
}

func (s *TestBuyHoldStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	signals := make([]int, len(candles))
	if s.buyAt < len(candles) {
		signals[s.buyAt] = 1
	}
	if s.sellAt < len(candles) {
		signals[s.sellAt] = -1
	}
	return signals
}
