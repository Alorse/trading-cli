package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
)

// ---------------------------------------------------------------------------
// Tool registration
// ---------------------------------------------------------------------------

// registerTool adds a single tool definition and its handler to the server,
// wrapping the handler with middleware (timeout, semaphore, panic recovery).
func registerTool(s *Server, def ToolDef, handler ToolHandler) {
	s.tools = append(s.tools, def)
	s.handlers[def.Name] = wrapHandler(handler)
}

// registerTools registers all 25 trading tools with the server.
func registerTools(s *Server) {
	// ── Screeners ──────────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "top_gainers",
		Description: "Top performing assets by price change on an exchange.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name (e.g. BINANCE, NASDAQ)"},
			{Name: "timeframe", Type: "string", Description: "Timeframe (e.g. 1m, 5m, 15m, 1h, 4h, 1D, 1W)", Default: "1D"},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, topGainersHandler)

	registerTool(s, ToolDef{
		Name:        "top_losers",
		Description: "Worst performing assets by price change on an exchange.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "1D"},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, topLosersHandler)

	registerTool(s, ToolDef{
		Name:        "bollinger_scan",
		Description: "Scan for Bollinger Band squeezes across symbols on an exchange.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "4h"},
			{Name: "bbw_threshold", Type: "number", Description: "Bollinger Band Width threshold", Default: 0.04},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, bollingerScanHandler)

	registerTool(s, ToolDef{
		Name:        "rating_filter",
		Description: "Filter symbols by Bollinger Band rating on an exchange.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "4h"},
			{Name: "rating", Type: "integer", Description: "Bollinger Band rating (-3 to 3)", Default: float64(2)},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, ratingFilterHandler)

	// ── Analysis ───────────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "coin_analysis",
		Description: "Full technical analysis for a specific symbol.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required, e.g. BTCUSDT, AAPL)"},
			{Name: "exchange", Type: "string", Description: "Exchange name", Default: "BINANCE"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "4h"},
		}, []string{"symbol"}),
	}, coinAnalysisHandler)

	// ── Patterns ───────────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "consecutive_candles_scan",
		Description: "Scan for consecutive bullish/bearish candle patterns across symbols.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "4h"},
			{Name: "pattern_type", Type: "string", Description: "Pattern type (bullish/bearish)", Default: "bullish"},
			{Name: "min_growth", Type: "number", Description: "Minimum growth percentage", Default: 2.0},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, consecutiveCandlesHandler)

	registerTool(s, ToolDef{
		Name:        "advanced_candle_pattern",
		Description: "Multi-timeframe candle pattern scoring with size increase detection.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name"},
			{Name: "base_timeframe", Type: "string", Description: "Base timeframe", Default: "4h"},
			{Name: "min_size_increase", Type: "number", Description: "Minimum size increase percentage", Default: 10.0},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, advancedCandleHandler)

	// ── Volume ─────────────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "volume_breakout_scanner",
		Description: "Scan for volume + price breakout signals across symbols.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "4h"},
			{Name: "volume_multiplier", Type: "number", Description: "Volume multiplier threshold", Default: 2.0},
			{Name: "price_change_min", Type: "number", Description: "Minimum price change percentage", Default: 3.0},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, volumeBreakoutHandler)

	registerTool(s, ToolDef{
		Name:        "volume_confirmation_analysis",
		Description: "Deep volume analysis confirming price movements for a specific symbol.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "exchange", Type: "string", Description: "Exchange name", Default: "BINANCE"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "4h"},
		}, []string{"symbol"}),
	}, volumeConfirmationHandler)

	registerTool(s, ToolDef{
		Name:        "smart_volume_scanner",
		Description: "Volume scanner with RSI filtering and recommendations.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "exchange", Type: "string", Description: "Exchange name"},
			{Name: "min_volume_ratio", Type: "number", Description: "Minimum volume ratio", Default: 2.0},
			{Name: "min_price_change", Type: "number", Description: "Minimum price change percentage", Default: 2.0},
			{Name: "rsi_range", Type: "string", Description: "RSI range (any/oversold/neutral/overbought)", Default: "any"},
			{Name: "limit", Type: "integer", Description: "Number of results", Default: float64(10)},
			{Name: "futures", Type: "boolean", Description: "Use futures/perpetual symbols", Default: false},
		}, nil),
	}, smartVolumeHandler)

	// ── Multi-timeframe ────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "multi_timeframe_analysis",
		Description: "Cross-timeframe alignment analysis for a specific symbol.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "exchange", Type: "string", Description: "Exchange name", Default: "BINANCE"},
		}, []string{"symbol"}),
	}, multiTimeframeHandler)

	// ── Sentiment & News ───────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "market_sentiment",
		Description: "Reddit sentiment analysis for a symbol or category.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "category", Type: "string", Description: "Category (all/crypto/stocks)", Default: "all"},
			{Name: "limit", Type: "integer", Description: "Number of posts to analyze", Default: float64(50)},
		}, []string{"symbol"}),
	}, marketSentimentHandler)

	registerTool(s, ToolDef{
		Name:        "financial_news",
		Description: "RSS financial news headlines for a symbol or category.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol to filter news"},
			{Name: "category", Type: "string", Description: "Category (all/crypto/stocks)", Default: "all"},
			{Name: "limit", Type: "integer", Description: "Number of articles", Default: float64(10)},
		}, nil),
	}, financialNewsHandler)

	registerTool(s, ToolDef{
		Name:        "combined_analysis",
		Description: "Technical + sentiment + news confluence analysis for a symbol.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "exchange", Type: "string", Description: "Exchange name", Default: "BINANCE"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "4h"},
			{Name: "category", Type: "string", Description: "Category (all/crypto/stocks)", Default: "all"},
		}, []string{"symbol"}),
	}, combinedAnalysisHandler)

	// ── Backtesting ────────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "backtest_strategy",
		Description: "Backtest a single trading strategy on historical data.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "strategy", Type: "string", Description: "Strategy name (required)"},
			{Name: "period", Type: "string", Description: "Period (1y, 6mo, 3mo, 1mo)", Default: "1y"},
			{Name: "interval", Type: "string", Description: "Interval (1d, 1h, 15m)", Default: "1d"},
			{Name: "initial_capital", Type: "number", Description: "Initial capital", Default: 10000.0},
			{Name: "commission_pct", Type: "number", Description: "Commission percentage", Default: 0.1},
			{Name: "slippage_pct", Type: "number", Description: "Slippage percentage", Default: 0.05},
			{Name: "include_trade_log", Type: "boolean", Description: "Include trade log", Default: false},
			{Name: "include_equity_curve", Type: "boolean", Description: "Include equity curve", Default: false},
		}, []string{"symbol", "strategy"}),
	}, backtestStrategyHandler)

	registerTool(s, ToolDef{
		Name:        "compare_strategies",
		Description: "Compare all available strategies on a symbol.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "period", Type: "string", Description: "Period", Default: "1y"},
			{Name: "interval", Type: "string", Description: "Interval", Default: "1d"},
			{Name: "initial_capital", Type: "number", Description: "Initial capital", Default: 10000.0},
		}, []string{"symbol"}),
	}, compareStrategiesHandler)

	registerTool(s, ToolDef{
		Name:        "walk_forward_backtest",
		Description: "Walk-forward backtest for overfitting detection.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "strategy", Type: "string", Description: "Strategy name (required)"},
			{Name: "period", Type: "string", Description: "Period", Default: "2y"},
			{Name: "interval", Type: "string", Description: "Interval", Default: "1d"},
			{Name: "initial_capital", Type: "number", Description: "Initial capital", Default: 10000.0},
			{Name: "commission_pct", Type: "number", Description: "Commission percentage", Default: 0.1},
			{Name: "slippage_pct", Type: "number", Description: "Slippage percentage", Default: 0.05},
			{Name: "n_splits", Type: "integer", Description: "Number of splits", Default: float64(3)},
			{Name: "train_ratio", Type: "number", Description: "Training ratio (0-1)", Default: 0.7},
		}, []string{"symbol", "strategy"}),
	}, walkForwardHandler)

	// ── Yahoo Finance ──────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "yahoo_price",
		Description: "Real-time price quote from Yahoo Finance.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required, e.g. AAPL, BTC-USD)"},
		}, []string{"symbol"}),
	}, yahooPriceHandler)

	registerTool(s, ToolDef{
		Name:        "market_snapshot",
		Description: "Global market overview snapshot.",
		InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
	}, marketSnapshotHandler)

	// ── Trade Planning ─────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "trade_plan",
		Description: "Generate a complete trade plan with entry/exit and risk management.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "exchange", Type: "string", Description: "Exchange name", Default: "NASDAQ"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "1D"},
		}, []string{"symbol"}),
	}, tradePlanHandler)

	registerTool(s, ToolDef{
		Name:        "fibonacci_retracement",
		Description: "Fibonacci retracement analysis for a symbol.",
		InputSchema: inputSchema([]PropertyDesc{
			{Name: "symbol", Type: "string", Description: "Symbol (required)"},
			{Name: "exchange", Type: "string", Description: "Exchange name", Default: "BINANCE"},
			{Name: "lookback", Type: "string", Description: "Lookback period (1M, 3M, 6M, 52W, ALL)", Default: "52W"},
			{Name: "timeframe", Type: "string", Description: "Timeframe", Default: "1D"},
		}, []string{"symbol"}),
	}, fibonacciHandler)

	// ── System ─────────────────────────────────────────────────────────────
	registerTool(s, ToolDef{
		Name:        "list_exchanges",
		Description: "List all available exchanges.",
		InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
	}, listExchangesHandler)

	registerTool(s, ToolDef{
		Name:        "version",
		Description: "Show trading-cli version information.",
		InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
	}, versionHandler)

	registerTool(s, ToolDef{
		Name:        "health",
		Description: "Check API connectivity and service health.",
		InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
	}, healthHandler)
}

// ---------------------------------------------------------------------------
// wrapHandler — middleware for every tool handler
// ---------------------------------------------------------------------------

// wrapHandler decorates a ToolHandler with:
//   - Concurrency semaphore (using s.toolSem)
//   - 60s timeout context
//   - Panic recovery
//   - Post-processing (ensures non-empty response)
func wrapHandler(inner ToolHandler) ToolHandler {
	return func(ctx context.Context, args map[string]any) (result []ContentBlock, structured interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[mcp] panic in tool handler: %v\n%s", r, debug.Stack())
				err = fmt.Errorf("internal error: %v", r)
			}
		}()

		timeoutCtx, cancel := context.WithTimeout(ctx, toolTimeout)
		defer cancel()

		result, structured, err = inner(timeoutCtx, args)
		if err != nil {
			return nil, nil, err
		}

		if len(result) == 0 {
			result = []ContentBlock{
				{Type: "text", Text: "Tool executed successfully with no output."},
			}
		}

		return result, structured, nil
	}
}

// ---------------------------------------------------------------------------
// Schema builder helper
// ---------------------------------------------------------------------------

// PropertyDesc describes a single tool input property for declarative registration.
type PropertyDesc struct {
	Name        string
	Type        string
	Description string
	Default     any
}

// inputSchema builds an InputSchema from a slice of PropertyDesc and an optional
// list of required field names.
func inputSchema(props []PropertyDesc, required []string) InputSchema {
	properties := make(map[string]Property, len(props))
	for _, p := range props {
		properties[p.Name] = Property{
			Type:        p.Type,
			Description: p.Description,
			Default:     p.Default,
		}
	}
	return InputSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// ---------------------------------------------------------------------------
// Argument helper functions
// ---------------------------------------------------------------------------

// argString extracts a string argument from the args map.
// Returns empty string if the key is missing or not a string.
func argString(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// argInt extracts an integer argument from the args map with a default value.
// Handles float64 (JSON numbers), int, and json.Number representations.
func argInt(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case json.Number:
			if i, err := n.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return def
}

// argFloat extracts a float64 argument from the args map with a default value.
// Handles float64, int, and json.Number representations.
func argFloat(args map[string]any, key string, def float64) float64 {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		case json.Number:
			if f, err := n.Float64(); err == nil {
				return f
			}
		}
	}
	return def
}

// argBool extracts a boolean argument from the args map with a default value.
func argBool(args map[string]any, key string, def bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

// ---------------------------------------------------------------------------
// Error and response helpers
// ---------------------------------------------------------------------------

// toolExecutionError wraps an error with a tool label for clearer error messages.
func toolExecutionError(label string, err error) error {
	return fmt.Errorf("%s: %w", label, err)
}

// buildAnnotatedContentBlocks builds a response with a summary text block
// followed by a structured JSON data block. The structured data is also
// returned as the structuredContent field.
func buildAnnotatedContentBlocks(summary string, data any) ([]ContentBlock, interface{}, error) {
	blocks := []ContentBlock{
		{Type: "text", Text: summary},
	}

	if data != nil {
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, nil, toolExecutionError("buildAnnotatedContentBlocks", err)
		}
		blocks = append(blocks, ContentBlock{
			Type: "text",
			Text: string(jsonBytes),
		})
	}

	return blocks, data, nil
}
