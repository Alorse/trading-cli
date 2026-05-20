package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds application configuration from environment variables.
type Config struct {
	HTTPTimeout time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	HTTPProxy   string
	HTTPSProxy  string
	UserAgent   string

	// PHEMEX private API (testnet by default).
	PhemexAPIID     string
	PhemexAPISecret string
	PhemexAPIURL    string
}

// Load reads configuration from environment variables with sensible defaults.
// Also loads a .env file from cwd if present (key=value lines, # comments).
func Load() *Config {
	loadDotenv(".env")
	return &Config{
		HTTPTimeout:     getEnvDuration("HTTP_TIMEOUT", 30*time.Second),
		MaxRetries:      getEnvInt("MAX_RETRIES", 3),
		RetryDelay:      getEnvDuration("RETRY_DELAY", 1*time.Second),
		HTTPProxy:       os.Getenv("HTTP_PROXY"),
		HTTPSProxy:      os.Getenv("HTTPS_PROXY"),
		UserAgent:       getEnvString("USER_AGENT", "trading-cli/0.1.0"),
		PhemexAPIID:     os.Getenv("PHEMEX_API_ID"),
		PhemexAPISecret: os.Getenv("PHEMEX_API_SECRET"),
		PhemexAPIURL:    getEnvString("PHEMEX_API_URL", "https://testnet-api.phemex.com"),
	}
}

// loadDotenv reads a .env file (best-effort) and sets unset env vars.
func loadDotenv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range splitLines(string(data)) {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		eq := indexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := trimSpace(line[:eq])
		val := trimSpace(line[eq+1:])
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		// Strip surrounding quotes if any.
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		}
		_ = os.Setenv(key, val)
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func getEnvString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n := fallback
	_, _ = fmt.Sscanf(v, "%d", &n)
	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
