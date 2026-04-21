package indicators

// MACDResult holds the MACD, Signal, and Histogram values.
type MACDResult struct {
	MACD      []float64
	Signal    []float64
	Histogram []float64
}

// MACD calculates the Moving Average Convergence Divergence.
// Standard parameters: fast=12, slow=26, signal=9.
// MACD = EMA(fast) - EMA(slow)
// Signal = EMA(signal) of MACD
// Histogram = MACD - Signal
// All slices have same length as prices; leading zeros where not enough data.
func MACD(prices []float64, fast, slow, signal int) MACDResult {
	result := MACDResult{
		MACD:      make([]float64, len(prices)),
		Signal:    make([]float64, len(prices)),
		Histogram: make([]float64, len(prices)),
	}

	if len(prices) == 0 {
		return result
	}

	// Calculate EMAs
	fastEMA := EMA(prices, fast)
	slowEMA := EMA(prices, slow)

	// Calculate MACD line (difference of EMAs)
	for i := 0; i < len(prices); i++ {
		if i >= slow-1 {
			result.MACD[i] = fastEMA[i] - slowEMA[i]
		}
	}

	// Calculate Signal line (EMA of MACD)
	signalEMA := EMA(result.MACD, signal)
	for i := 0; i < len(prices); i++ {
		result.Signal[i] = signalEMA[i]
	}

	// Calculate Histogram (MACD - Signal)
	for i := 0; i < len(prices); i++ {
		result.Histogram[i] = result.MACD[i] - result.Signal[i]
	}

	return result
}
