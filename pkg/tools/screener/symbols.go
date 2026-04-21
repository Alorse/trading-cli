package screener

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadSymbols reads symbol data from data/symbols/{exchange}.txt
// Returns a slice of symbol strings, skipping blank lines and comments (lines starting with #)
// Returns an error if the file is not found
func LoadSymbols(exchange string) ([]string, error) {
	// Construct path to symbols file relative to current working directory
	filename := strings.ToLower(exchange) + ".txt"
	filepath := filepath.Join("data", "symbols", filename)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open symbols file for %s: %w", exchange, err)
	}
	defer file.Close()

	var symbols []string
	scanner := bufio.NewScanner(file)

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
