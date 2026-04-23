package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	scannerURL = "https://scanner.tradingview.com/crypto/scan"
	pageSize   = 500
)

var exchanges = []string{
	"KUCOIN", "BINANCE", "BYBIT", "OKX", "BITGET",
	"COINBASE", "GATE", "MEXC", "HTX", "BITFINEX",
	"BINGX", "PHEMEX", "KRAKEN",
}

type scanRequest struct {
	Filter  []scanFilter     `json:"filter"`
	Symbols scanSymbolsQuery `json:"symbols"`
	Columns []string         `json:"columns"`
	Options scanOptions      `json:"options"`
	Range   [2]int           `json:"range"`
}

type scanFilter struct {
	Left      string `json:"left"`
	Operation string `json:"operation"`
	Right     string `json:"right"`
}

type scanSymbolsQuery struct {
	Query scanQueryTypes `json:"query"`
}

type scanQueryTypes struct {
	Types []string `json:"types"`
}

type scanOptions struct {
	Lang string `json:"lang"`
}

type scanResponse struct {
	TotalCount int `json:"totalCount"`
	Data       []struct {
		S string `json:"s"`
	} `json:"data"`
}

func fetchSymbols(exchange string) ([]string, error) {
	var allSymbols []string
	offset := 0

	for {
		reqBody := scanRequest{
			Filter: []scanFilter{
				{Left: "exchange", Operation: "equal", Right: exchange},
				{Left: "type", Operation: "equal", Right: "spot"},
			},
			Symbols: scanSymbolsQuery{Query: scanQueryTypes{Types: []string{}}},
			Columns: []string{"name"},
			Options: scanOptions{Lang: "en"},
			Range:   [2]int{offset, offset + pageSize},
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}

		req, err := http.NewRequest("POST", scannerURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "https://www.tradingview.com")
		req.Header.Set("Referer", "https://www.tradingview.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		var scanResp scanResponse
		if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}

		if len(scanResp.Data) == 0 {
			break
		}

		for _, item := range scanResp.Data {
			allSymbols = append(allSymbols, item.S)
		}

		fmt.Printf("  %s: fetched %d/%d\n", exchange, len(allSymbols), scanResp.TotalCount)

		if len(allSymbols) >= scanResp.TotalCount {
			break
		}

		offset += pageSize
	}

	sort.Strings(allSymbols)
	return allSymbols, nil
}

func main() {
	dataDir := os.Args[1]
	if dataDir == "" {
		dataDir = "pkg/tools/screener/data"
	}

	// Verify directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "data directory not found: %s\n", dataDir)
		os.Exit(1)
	}

	for _, exchange := range exchanges {
		fmt.Printf("Fetching %s...\n", exchange)

		symbols, err := fetchSymbols(exchange)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", exchange, err)
			os.Exit(1)
		}

		filename := fmt.Sprintf("%s/%s.txt", strings.TrimRight(dataDir, "/"), strings.ToLower(exchange))
		content := strings.Join(symbols, "\n") + "\n"

		if err := os.WriteFile(filename, []byte(content), fs.FileMode(0644)); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", filename, err)
			os.Exit(1)
		}

		fmt.Printf("  Wrote %d symbols to %s\n\n", len(symbols), filename)
	}

	fmt.Println("Done!")
}
