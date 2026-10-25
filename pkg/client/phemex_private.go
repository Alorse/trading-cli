package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// PhemexPrivateClient performs authenticated requests against the PHEMEX REST API.
// Signature scheme (per PHEMEX docs):
//
//	sig = HMAC_SHA256(secret, URLPath + QueryString + Expiry + Body)
//
// where the secret is base64-url-decoded before being used as the HMAC key.
type PhemexPrivateClient struct {
	baseURL string
	apiID   string
	secret  []byte
	http    *http.Client
}

// NewPhemexPrivateClient builds a private client. If apiSecret is base64-url-encoded
// it will be decoded; otherwise it is used as-is (covers both formats robustly).
func NewPhemexPrivateClient(baseURL, apiID, apiSecret string) *PhemexPrivateClient {
	secretBytes := decodeSecret(apiSecret)
	return &PhemexPrivateClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiID:   apiID,
		secret:  secretBytes,
		http:    &http.Client{Timeout: 20 * time.Second},
	}
}

func decodeSecret(s string) []byte {
	// Try url-safe base64 (without padding), then standard, then raw bytes fallback.
	for _, enc := range []*base64.Encoding{
		base64.RawURLEncoding, base64.URLEncoding,
		base64.RawStdEncoding, base64.StdEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b
		}
	}
	return []byte(s)
}

// BaseURL returns the base URL in use (helpful to confirm testnet vs prod).
func (c *PhemexPrivateClient) BaseURL() string { return c.baseURL }

// Do executes a signed request. method = GET/POST/PUT/DELETE.
// urlPath = "/g-accounts/positions" (no host, must start with /).
// query = map of query params; will be sorted and url-encoded.
// body = optional JSON-serializable struct (nil for none).
// Returns the raw response bytes.
func (c *PhemexPrivateClient) Do(ctx context.Context, method, urlPath string, query map[string]string, body interface{}) ([]byte, error) {
	if c.apiID == "" || len(c.secret) == 0 {
		return nil, fmt.Errorf("phemex: PHEMEX_API_ID/PHEMEX_API_SECRET not set")
	}

	// Build sorted query string (no leading ?).
	qs := buildQuery(query)

	// Body.
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
	}
	bodyStr := string(bodyBytes)

	// Expiry: now + 60s.
	expiry := fmt.Sprintf("%d", time.Now().Unix()+60)

	// Sign: HMAC-SHA256(secret, path + qs + expiry + body)
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(urlPath))
	mac.Write([]byte(qs))
	mac.Write([]byte(expiry))
	mac.Write([]byte(bodyStr))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Build URL.
	fullURL := c.baseURL + urlPath
	if qs != "" {
		fullURL += "?" + qs
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("x-phemex-access-token", c.apiID)
	req.Header.Set("x-phemex-request-expiry", expiry)
	req.Header.Set("x-phemex-request-signature", signature)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("phemex http: %w", err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return respBytes, fmt.Errorf("phemex http %d: %s", resp.StatusCode, string(respBytes))
	}

	// PHEMEX returns {"code":0,"msg":"OK","data":...} on success; non-zero code = logical error.
	var envelope struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &envelope); err == nil && envelope.Code != 0 {
		return respBytes, fmt.Errorf("phemex error code=%d msg=%s", envelope.Code, envelope.Msg)
	}
	return respBytes, nil
}

// buildQuery returns "k1=v1&k2=v2" with keys sorted, no leading ?.
func buildQuery(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(params[k])
	}
	return sb.String()
}
