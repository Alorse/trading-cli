package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds application configuration from environment variables.
type Config struct {
	HTTPTimeout    time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	HTTPProxy      string
	HTTPSProxy     string
	UserAgent      string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		HTTPTimeout:    getEnvDuration("HTTP_TIMEOUT", 30*time.Second),
		MaxRetries:     getEnvInt("MAX_RETRIES", 3),
		RetryDelay:     getEnvDuration("RETRY_DELAY", 1*time.Second),
		HTTPProxy:      os.Getenv("HTTP_PROXY"),
		HTTPSProxy:     os.Getenv("HTTPS_PROXY"),
		UserAgent:      getEnvString("USER_AGENT", "trading-cli/0.1.0"),
	}
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
