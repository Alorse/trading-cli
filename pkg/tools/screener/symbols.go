package screener

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"strings"
)

//go:embed data/*.txt
var symbolsFS embed.FS

// LoadSymbols reads symbol data from the embedded data/symbols/{exchange}.txt
// Returns a slice of symbol strings, skipping blank lines and comments (lines starting with #)
// Returns an error if the file is not found
func LoadSymbols(exchange string) ([]string, error) {
	filename := strings.ToLower(exchange) + ".txt"

	data, err := symbolsFS.ReadFile("data/" + filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open symbols file for %s: %w", exchange, err)
	}

	var symbols []string
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		symbols = append(symbols, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading symbols file: %w", err)
	}

	return symbols, nil
}

// FormatTicker returns a formatted ticker string in the form "EXCHANGE:SYMBOL"
// If the symbol already has an exchange prefix, it returns it as-is
func FormatTicker(exchange, symbol string) string {
	// Check if symbol already has exchange prefix (format: "EXCHANGE:SYMBOL")
	if strings.Contains(symbol, ":") {
		return symbol
	}

	// Add exchange prefix
	return strings.ToUpper(exchange) + ":" + symbol
}
