package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alorse/trading-cli/internal/config"
)

// HTTPClient wraps http.Client with retry logic and configurable timeout.
type HTTPClient struct {
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
	userAgent  string
}

// NewHTTPClient creates a new HTTP client from config.
func NewHTTPClient(cfg *config.Config) *HTTPClient {
	transport := &http.Transport{}
	if cfg.HTTPProxy != "" {
		proxyURL, _ := url.Parse(cfg.HTTPProxy)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout:   cfg.HTTPTimeout,
			Transport: transport,
		},
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
		userAgent:  cfg.UserAgent,
	}
}

// Get performs a GET request with retry logic.
func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	return c.doRequest(ctx, "GET", url, nil)
}

// Post performs a POST request with retry logic.
func (c *HTTPClient) Post(ctx context.Context, url string, body io.Reader) ([]byte, error) {
	return c.doRequest(ctx, "POST", url, body)
}

// PostWithHeaders performs a POST request with custom headers and retry logic.
func (c *HTTPClient) PostWithHeaders(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	return c.doRequestWithHeaders(ctx, "POST", url, body, headers)
}

// GetWithHeaders performs a GET request with custom headers and retry logic.
func (c *HTTPClient) GetWithHeaders(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return c.doRequestWithHeaders(ctx, "GET", url, nil, headers)
}

func (c *HTTPClient) doRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	return c.doRequestWithHeaders(ctx, method, url, body, nil)
}

func (c *HTTPClient) doRequestWithHeaders(ctx context.Context, method, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("User-Agent", c.userAgent)
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("client error: %d: %s", resp.StatusCode, string(data))
		}

		return data, nil
	}

	return nil, fmt.Errorf("after %d retries: %w", c.maxRetries, lastErr)
}
