package models

// OHLCV represents a single candlestick bar.
type OHLCV struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// IndicatorValues holds computed technical indicator values.
type IndicatorValues struct {
	RSI        float64 `json:"rsi,omitempty"`
	MACDLine   float64 `json:"macdLine,omitempty"`
	MACDSignal float64 `json:"macdSignal,omitempty"`
	MACDHist   float64 `json:"macdHist,omitempty"`
	SMA20      float64 `json:"sma20,omitempty"`
	SMA50      float64 `json:"sma50,omitempty"`
	SMA200     float64 `json:"sma200,omitempty"`
	EMA9       float64 `json:"ema9,omitempty"`
	EMA20      float64 `json:"ema20,omitempty"`
	EMA50      float64 `json:"ema50,omitempty"`
	EMA200     float64 `json:"ema200,omitempty"`
	BBUpper    float64 `json:"bbUpper,omitempty"`
	BBMiddle   float64 `json:"bbMiddle,omitempty"`
	BBLower    float64 `json:"bbLower,omitempty"`
	ATR        float64 `json:"atr,omitempty"`
	ADX        float64 `json:"adx,omitempty"`
	CCI        float64 `json:"cci,omitempty"`
	StochK     float64 `json:"stochK,omitempty"`
	StochD     float64 `json:"stochD,omitempty"`
	WilliamsR  float64 `json:"williamsR,omitempty"`
	OBV        float64 `json:"obv,omitempty"`
	VWAP       float64 `json:"vwap,omitempty"`
	MFI        float64 `json:"mfi,omitempty"`
}

// BacktestTrade represents a single executed trade.
type BacktestTrade struct {
	EntryDate  string  `json:"entryDate"`
	ExitDate   string  `json:"exitDate"`
	EntryPrice float64 `json:"entryPrice"`
	ExitPrice  float64 `json:"exitPrice"`
	Shares     float64 `json:"shares"`
	PnL        float64 `json:"pnl"`
	PnLPct     float64 `json:"pnlPct"`
	Type       string  `json:"type"` // "long"
}

// BacktestResult holds the outcome of a backtest run.
type BacktestResult struct {
	Strategy       string          `json:"strategy"`
	Symbol         string          `json:"symbol"`
	Period         string          `json:"period"`
	InitialCapital float64         `json:"initialCapital"`
	FinalCapital   float64         `json:"finalCapital"`
	TotalReturn    float64         `json:"totalReturnPct"`
	TotalTrades    int             `json:"totalTrades"`
	WinningTrades  int             `json:"winningTrades"`
	LosingTrades   int             `json:"losingTrades"`
	WinRate        float64         `json:"winRatePct"`
	AvgGain        float64         `json:"avgGainPct"`
	AvgLoss        float64         `json:"avgLossPct"`
	MaxDrawdown    float64         `json:"maxDrawdownPct"`
	ProfitFactor   float64         `json:"profitFactor"`
	SharpeRatio    float64         `json:"sharpeRatio"`
	CalmarRatio    float64         `json:"calmarRatio"`
	Expectancy     float64         `json:"expectancyPct"`
	BestTrade      float64         `json:"bestTradePct"`
	WorstTrade     float64         `json:"worstTradePct"`
	TradeLog       []BacktestTrade `json:"tradeLog,omitempty"`
	EquityCurve    []EquityPoint   `json:"equityCurve,omitempty"`
}

// EquityPoint represents a point in the equity curve.
type EquityPoint struct {
	Date   string  `json:"date"`
	Equity float64 `json:"equity"`
}
