package system

import (
	"context"
	"runtime"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
	"github.com/alorse/trading-cli/pkg/utils"
)

const Version = "0.1.0"

type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"goVersion"`
}

type HealthStatus struct {
	TradingView  string `json:"tradingView"`
	YahooFinance string `json:"yahooFinance"`
	Timestamp    string `json:"timestamp"`
}

func RunVersion() error {
	return utils.PrintJSON(VersionInfo{
		Version:   Version,
		GoVersion: runtime.Version(),
	})
}

func RunHealth(cfg *config.Config) error {
	c := client.NewHTTPClient(cfg)
	tvStatus := "ok"
	yfStatus := "ok"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := c.Get(ctx, "https://scanner.tradingview.com/crypto/scan"); err != nil {
		tvStatus = "error"
	}

	if _, err := c.Get(ctx, "https://query1.finance.yahoo.com/v8/finance/chart/AAPL"); err != nil {
		yfStatus = "error"
	}

	return utils.PrintJSON(HealthStatus{
		TradingView:  tvStatus,
		YahooFinance: yfStatus,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	})
}
