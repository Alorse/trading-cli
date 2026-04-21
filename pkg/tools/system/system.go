package system

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/alorse/trading-cli/internal/config"
	"github.com/alorse/trading-cli/pkg/client"
)

const Version = "0.1.0"

// VersionInfo holds version details.
type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"goVersion"`
}

// HealthStatus holds connectivity check results.
type HealthStatus struct {
	TradingView  string `json:"tradingView"`
	YahooFinance string `json:"yahooFinance"`
	Timestamp    string `json:"timestamp"`
}

// RunVersion returns version information.
func RunVersion() error {
	info := VersionInfo{
		Version:   Version,
		GoVersion: runtime.Version(),
	}
	return printJSON(info)
}

// RunHealth checks connectivity to data sources.
func RunHealth(cfg *config.Config) error {
	c := client.NewHTTPClient(cfg)
	tvStatus := "ok"
	yfStatus := "ok"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check TradingView
	_, err := c.Get(ctx, "https://scanner.tradingview.com/crypto/scan")
	if err != nil {
		tvStatus = "error"
	}

	// Check Yahoo Finance
	_, err = c.Get(ctx, "https://query1.finance.yahoo.com/v8/finance/chart/AAPL")
	if err != nil {
		yfStatus = "error"
	}

	status := HealthStatus{
		TradingView:  tvStatus,
		YahooFinance: yfStatus,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
	return printJSON(status)
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
