package system

import (
	"github.com/alorse/trading-cli/pkg/models"
	"github.com/alorse/trading-cli/pkg/utils"
)

func GetExchanges() []models.ExchangeInfo {
	return []models.ExchangeInfo{
		{Name: "KUCOIN", Type: "crypto"},
		{Name: "BINANCE", Type: "crypto"},
		{Name: "BYBIT", Type: "crypto"},
		{Name: "OKX", Type: "crypto"},
		{Name: "BITGET", Type: "crypto"},
		{Name: "COINBASE", Type: "crypto"},
		{Name: "GATE", Type: "crypto"},
		{Name: "MEXC", Type: "crypto"},
		{Name: "HTX", Type: "crypto"},
		{Name: "BITFINEX", Type: "crypto"},
		{Name: "BINGX", Type: "crypto"},
		{Name: "PHEMEX", Type: "crypto"},
		{Name: "KRAKEN", Type: "crypto"},
		{Name: "NASDAQ", Type: "stock"},
		{Name: "NYSE", Type: "stock"},
	}
}

func RunListExchanges() error {
	return utils.PrintJSON(GetExchanges())
}
