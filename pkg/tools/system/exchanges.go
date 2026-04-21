package system

import "github.com/alorse/trading-cli/pkg/models"

// GetExchanges returns all supported exchanges.
func GetExchanges() []models.ExchangeInfo {
	return []models.ExchangeInfo{
		{Name: "KUCOIN", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "BINANCE", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "BYBIT", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "OKX", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "BITGET", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "COINBASE", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "GATEIO", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "MEXC", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "HUOBI", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "BITFINEX", Type: "crypto", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "NASDAQ", Type: "stock", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
		{Name: "NYSE", Type: "stock", Timeframes: []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}},
	}
}

// RunListExchanges prints all supported exchanges.
func RunListExchanges() error {
	return printJSON(GetExchanges())
}
