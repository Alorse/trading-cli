package backtest

import (
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/indicators"
)

// RSIStrategy implements a simple RSI-based strategy.
// Buy when RSI(14) < 40, Sell when RSI(14) > 60.
type RSIStrategy struct{}

func (s *RSIStrategy) Name() string {
	return "rsi"
}

func (s *RSIStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	// Extract closes
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	// Calculate RSI(14)
	rsiValues := indicators.RSI(closes, 14)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 0; i < len(candles); i++ {
		if rsiValues[i] == 0 {
			continue
		}

		if !inPosition && rsiValues[i] < 40 {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && rsiValues[i] > 60 {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// BollingerStrategy implements a Bollinger Bands strategy.
// Buy when close < BB(20,2).Lower, Sell when close > BB(20,2).Middle.
type BollingerStrategy struct{}

func (s *BollingerStrategy) Name() string {
	return "bollinger"
}

func (s *BollingerStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	bb := indicators.BollingerBands(closes, 20, 2.0)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 0; i < len(candles); i++ {
		if bb.Lower[i] == 0 {
			continue
		}

		if !inPosition && closes[i] < bb.Lower[i] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && closes[i] > bb.Middle[i] {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// MACDStrategy implements a MACD crossover strategy.
// Buy when MACD crosses above Signal, Sell when MACD crosses below Signal.
type MACDStrategy struct{}

func (s *MACDStrategy) Name() string {
	return "macd"
}

func (s *MACDStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	macdResult := indicators.MACD(closes, 12, 26, 9)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 1; i < len(candles); i++ {
		// Ensure we have valid MACD and Signal values
		if macdResult.MACD[i] == 0 && macdResult.Signal[i] == 0 {
			continue
		}

		// Check for crossover: MACD crosses above Signal
		if !inPosition && macdResult.MACD[i] > macdResult.Signal[i] && macdResult.MACD[i-1] <= macdResult.Signal[i-1] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && macdResult.MACD[i] < macdResult.Signal[i] && macdResult.MACD[i-1] >= macdResult.Signal[i-1] {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// EMACrossStrategy implements an EMA crossover strategy.
// Buy when EMA(20) crosses above EMA(50), Sell when EMA(20) crosses below EMA(50).
type EMACrossStrategy struct{}

func (s *EMACrossStrategy) Name() string {
	return "ema-cross"
}

func (s *EMACrossStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	ema20 := indicators.EMA(closes, 20)
	ema50 := indicators.EMA(closes, 50)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 1; i < len(candles); i++ {
		if ema20[i] == 0 || ema50[i] == 0 {
			continue
		}

		// EMA(20) crosses above EMA(50)
		if !inPosition && ema20[i] > ema50[i] && ema20[i-1] <= ema50[i-1] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && ema20[i] < ema50[i] && ema20[i-1] >= ema50[i-1] {
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// SupertrendStrategy implements a Supertrend strategy.
// Buy when direction flips to +1, Sell when direction flips to -1.
// Uses period=10, multiplier=3.0.
type SupertrendStrategy struct{}

func (s *SupertrendStrategy) Name() string {
	return "supertrend"
}

func (s *SupertrendStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))

	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}

	st := indicators.Supertrend(highs, lows, closes, 10, 3.0)

	signals := make([]int, len(candles))
	var lastDirection int

	for i := 0; i < len(candles); i++ {
		currentDirection := st.Direction[i]

		if currentDirection == 0 {
			continue
		}

		// Buy when direction flips to +1
		if currentDirection == 1 && lastDirection != 1 {
			signals[i] = 1
		} else if currentDirection == -1 && lastDirection != -1 {
			// Sell when direction flips to -1
			signals[i] = -1
		}

		lastDirection = currentDirection
	}

	return signals
}

// DonchianStrategy implements a Donchian breakout strategy.
// Buy when high > previous period upper, Sell when close < current lower.
// Uses period=20.
type DonchianStrategy struct{}

func (s *DonchianStrategy) Name() string {
	return "donchian"
}

func (s *DonchianStrategy) GenerateSignals(candles []client.YahooOHLCV) []int {
	if len(candles) == 0 {
		return []int{}
	}

	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))

	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
	}

	dc := indicators.DonchianChannel(highs, lows, 20)

	signals := make([]int, len(candles))
	var inPosition bool

	for i := 1; i < len(candles); i++ {
		if dc.Upper[i] == 0 || dc.Lower[i] == 0 {
			continue
		}

		// Buy when high breaks above previous upper band
		if !inPosition && highs[i] > dc.Upper[i-1] {
			signals[i] = 1 // Buy
			inPosition = true
		} else if inPosition && candles[i].Close < dc.Lower[i] {
			// Sell when close falls below current lower band
			signals[i] = -1 // Sell
			inPosition = false
		}
	}

	return signals
}

// GetStrategy returns the appropriate strategy for the given name.
func GetStrategy(name string) Strategy {
	switch name {
	case "rsi":
		return &RSIStrategy{}
	case "bollinger":
		return &BollingerStrategy{}
	case "macd":
		return &MACDStrategy{}
	case "ema-cross":
		return &EMACrossStrategy{}
	case "supertrend":
		return &SupertrendStrategy{}
	case "donchian":
		return &DonchianStrategy{}
	default:
		return nil
	}
}
