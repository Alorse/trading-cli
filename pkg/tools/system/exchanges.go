package system

import (
	"github.com/alorse/trading-cli/pkg/models"
	"github.com/alorse/trading-cli/pkg/utils"
)

var supportedTimeframes = []string{"5m", "15m", "1h", "4h", "1D", "1W", "1M"}

func GetExchanges() []models.ExchangeInfo {
	tf := supportedTimeframes
	return []models.ExchangeInfo{
		{Name: "KUCOIN", Type: "crypto", Timeframes: tf},
		{Name: "BINANCE", Type: "crypto", Timeframes: tf},
		{Name: "BYBIT", Type: "crypto", Timeframes: tf},
		{Name: "OKX", Type: "crypto", Timeframes: tf},
		{Name: "BITGET", Type: "crypto", Timeframes: tf},
		{Name: "COINBASE", Type: "crypto", Timeframes: tf},
		{Name: "GATE", Type: "crypto", Timeframes: tf},
		{Name: "MEXC", Type: "crypto", Timeframes: tf},
		{Name: "HTX", Type: "crypto", Timeframes: tf},
		{Name: "BITFINEX", Type: "crypto", Timeframes: tf},
		{Name: "BINGX", Type: "crypto", Timeframes: tf},
		{Name: "PHEMEX", Type: "crypto", Timeframes: tf},
		{Name: "KRAKEN", Type: "crypto", Timeframes: tf},
		{Name: "NASDAQ", Type: "stock", Timeframes: tf},
		{Name: "NYSE", Type: "stock", Timeframes: tf},
	}
}

func RunListExchanges() error {
	return utils.PrintJSON(GetExchanges())
}
