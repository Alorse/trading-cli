package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const tvScannerBase = "https://scanner.tradingview.com"

var ExchangeToScreener = map[string]string{
	"KUCOIN":   "crypto",
	"BINANCE":  "crypto",
	"BITGET":   "crypto",
	"BYBIT":    "crypto",
	"OKX":      "crypto",
	"COINBASE": "crypto",
	"GATEIO":   "crypto",
	"MEXC":     "crypto",
	"HUOBI":    "crypto",
	"BITFINEX": "crypto",
	"NASDAQ":   "america",
	"NYSE":     "america",
}

var DefaultColumns = []string{
	"open", "high", "low", "close", "volume",
	"SMA10", "SMA20", "SMA30", "SMA50", "SMA100", "SMA200",
	"EMA9", "EMA20", "EMA50", "EMA100", "EMA200",
	"BB.upper", "BB.lower",
	"RSI", "RSI[1]",
	"MACD.macd", "MACD.signal",
	"Stoch.K", "Stoch.D",
	"ATR", "ADX", "CCI20", "W.R", "AO", "Mom",
	"P.SAR", "Ichimoku.BLine", "VWAP", "VWMA", "HullMA9",
	"volume.SMA20", "Stoch.RSI.K",
	"Pivot.M.Classic.Middle", "Pivot.M.Classic.R1", "Pivot.M.Classic.S1",
	"Recommend.All", "Recommend.MA", "Recommend.Other",
	"UO", "change", "Volatility.D",
}

type tvScanRequest struct {
	Symbols struct {
		Tickers []string `json:"tickers"`
	} `json:"symbols"`
	Columns []string `json:"columns"`
}

type TVSymbolData struct {
	Symbol string
	Values map[string]interface{}
}

type TradingViewClient struct {
	http *HTTPClient
}

func NewTradingViewClient(http *HTTPClient) *TradingViewClient {
	return &TradingViewClient{http: http}
}

func (c *TradingViewClient) GetMultipleAnalysis(ctx context.Context, screener string, tickers []string, columns []string) ([]TVSymbolData, error) {
	if len(columns) == 0 {
		columns = DefaultColumns
	}

	const batchSize = 200
	var results []TVSymbolData

	for i := 0; i < len(tickers); i += batchSize {
		end := i + batchSize
		if end > len(tickers) {
			end = len(tickers)
		}
		batch := tickers[i:end]

		batchResults, err := c.fetchBatch(ctx, screener, batch, columns)
		if err != nil {
			return nil, err
		}
		results = append(results, batchResults...)
	}

	return results, nil
}

func (c *TradingViewClient) fetchBatch(ctx context.Context, screener string, tickers []string, columns []string) ([]TVSymbolData, error) {
	req := tvScanRequest{Columns: columns}
	req.Symbols.Tickers = tickers

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s/scan", tvScannerBase, screener)
	headers := map[string]string{
		"Content-Type": "application/json",
		"Origin":       "https://www.tradingview.com",
		"Referer":      "https://www.tradingview.com/",
	}

	respBytes, err := c.http.PostWithHeaders(ctx, url, bytes.NewReader(body), headers)
	if err != nil {
		return nil, fmt.Errorf("tradingview scan: %w", err)
	}

	var raw struct {
		Data []struct {
			S string        `json:"s"`
			D []interface{} `json:"d"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]TVSymbolData, 0, len(raw.Data))
	for _, item := range raw.Data {
		values := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if i < len(item.D) {
				values[col] = item.D[i]
			}
		}
		results = append(results, TVSymbolData{
			Symbol: item.S,
			Values: values,
		})
	}

	return results, nil
}

func ScreenerForExchange(exchange string) (string, error) {
	screener, ok := ExchangeToScreener[strings.ToUpper(exchange)]
	if !ok {
		return "", fmt.Errorf("unknown exchange: %s", exchange)
	}
	return screener, nil
}
