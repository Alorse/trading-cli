package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alorse/trading-cli/internal/config"
)

type HTTPClient struct {
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
	userAgent  string
}

func NewHTTPClient(cfg *config.Config) *HTTPClient {
	transport := &http.Transport{}
	if cfg.HTTPProxy != "" {
		proxyURL, err := url.Parse(cfg.HTTPProxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
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

func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	return c.doRequestWithHeaders(ctx, "GET", url, nil, nil)
}

func (c *HTTPClient) Post(ctx context.Context, url string, body io.Reader) ([]byte, error) {
	return c.doRequestWithHeaders(ctx, "POST", url, body, nil)
}

func (c *HTTPClient) PostWithHeaders(ctx context.Context, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	return c.doRequestWithHeaders(ctx, "POST", url, body, headers)
}

func (c *HTTPClient) GetWithHeaders(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return c.doRequestWithHeaders(ctx, "GET", url, nil, headers)
}

func (c *HTTPClient) doRequestWithHeaders(ctx context.Context, method, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	// Buffer body upfront so retries can replay it.
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}

		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
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
