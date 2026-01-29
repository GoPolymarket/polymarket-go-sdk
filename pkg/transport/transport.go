package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

const (
	defaultMaxRetries = 3
	defaultMinWait    = 100 * time.Millisecond
	defaultMaxWait    = 2 * time.Second
)

// Doer matches http.Client's Do method.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a wrapper around http.Client that handles common API tasks.
type Client struct {
	httpClient    Doer
	baseURL       string
	userAgent     string
	signer        auth.Signer
	apiKey        *auth.APIKey
	builder       *auth.BuilderConfig
	useServerTime bool
}

// NewClient creates a new transport client.
func NewClient(httpClient Doer, baseURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	// Ensure base URL doesn't have a trailing slash for consistency
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		userAgent:  "github.com/GoPolymarket/polymarket-go-sdk/1.0",
	}
}

// CloneWithBaseURL creates a new client with the same HTTP doer and user agent,
// but a different base URL. Auth/builder config are intentionally not copied.
func (c *Client) CloneWithBaseURL(baseURL string) *Client {
	if c == nil {
		return NewClient(nil, baseURL)
	}
	clone := NewClient(c.httpClient, baseURL)
	clone.userAgent = c.userAgent
	clone.useServerTime = c.useServerTime
	return clone
}

// SetUserAgent overrides the default user agent.
func (c *Client) SetUserAgent(userAgent string) {
	if userAgent != "" {
		c.userAgent = userAgent
	}
}

// SetAuth sets the credentials for L2 authentication.
func (c *Client) SetAuth(signer auth.Signer, apiKey *auth.APIKey) {
	c.signer = signer
	c.apiKey = apiKey
}

// SetBuilderConfig sets the builder config for extra builder headers.
func (c *Client) SetBuilderConfig(config *auth.BuilderConfig) {
	c.builder = config
}

// SetUseServerTime enables or disables server-time signing.
func (c *Client) SetUseServerTime(use bool) {
	c.useServerTime = use
}

// Call executes an HTTP request.
func (c *Client) Call(ctx context.Context, method, path string, query url.Values, body interface{}, dest interface{}, headers map[string]string) error {
	u := c.baseURL + "/" + strings.TrimLeft(path, "/")

	// Append query parameters
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	payload, serialized, err := MarshalBody(body)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= defaultMaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms...
			wait := defaultMinWait * time.Duration(1<<uint(attempt-1))
			if wait > defaultMaxWait {
				wait = defaultMaxWait
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		var reqBody io.Reader
		if len(payload) > 0 {
			reqBody = bytes.NewBuffer(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")
		if len(payload) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}

		// Set custom headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// L2 Authentication (only if no custom auth headers provided)
		// If custom POLY_SIGNATURE is provided, skip auto-L2 auth
		if c.apiKey != nil && c.signer != nil && req.Header.Get(auth.HeaderPolySignature) == "" {
			ts := time.Now().Unix()
			if c.useServerTime {
				serverTime, err := c.serverTime(ctx)
				if err != nil {
					lastErr = fmt.Errorf("failed to get server time: %w", err)
					continue
				}
				ts = serverTime
			}
			signPath := "/" + strings.TrimLeft(path, "/")

			message := fmt.Sprintf("%d%s%s", ts, method, signPath)
			if serialized != nil && *serialized != "" {
				message += strings.ReplaceAll(*serialized, "'", "\"")
			}

			sig, err := auth.SignHMAC(c.apiKey.Secret, message)
			if err != nil {
				return fmt.Errorf("failed to sign request: %w", err)
			}

			req.Header.Set(auth.HeaderPolyAddress, c.signer.Address().Hex())
			req.Header.Set(auth.HeaderPolyAPIKey, c.apiKey.Key)
			req.Header.Set(auth.HeaderPolyPassphrase, c.apiKey.Passphrase)
			req.Header.Set(auth.HeaderPolyTimestamp, fmt.Sprintf("%d", ts))
			req.Header.Set(auth.HeaderPolySignature, sig)

			if c.builder != nil && c.builder.IsValid() {
				builderHeaders, err := c.builder.Headers(ctx, method, signPath, serialized, ts)
				if err != nil {
					return fmt.Errorf("failed to build builder headers: %w", err)
				}
				for k, values := range builderHeaders {
					if len(values) == 0 || req.Header.Get(k) != "" {
						continue
					}
					req.Header.Set(k, values[0])
				}
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Read response body
		respBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", readErr)
			continue
		}

		// Check for error status codes
		if resp.StatusCode >= 400 {
			// Check if retryable (429 or 5xx)
			if resp.StatusCode == 429 || resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBytes))
				continue
			}

			var apiErr types.Error
			if err := json.Unmarshal(respBytes, &apiErr); err == nil && (apiErr.Message != "" || apiErr.Code != "") {
				apiErr.Status = resp.StatusCode
				apiErr.Path = path
				return &apiErr
			}
			// Fallback for unknown error formats
			return &types.Error{
				Status:  resp.StatusCode,
				Message: string(respBytes),
				Path:    path,
			}
		}

		// Unmarshal success response
		if dest != nil {
			if err := json.Unmarshal(respBytes, dest); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		return nil
	}

	return lastErr
}

func (c *Client) serverTime(ctx context.Context) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/time", nil)
	if err != nil {
		return 0, fmt.Errorf("create server time request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("server time request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read server time response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("server time status %d", resp.StatusCode)
	}

	var ts int64
	if err := json.Unmarshal(body, &ts); err == nil && ts > 0 {
		return ts, nil
	}

	var payload struct {
		Timestamp  int64  `json:"timestamp"`
		ServerTime string `json:"server_time"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Timestamp > 0 {
			return payload.Timestamp, nil
		}
		if payload.ServerTime != "" {
			if parsed, parseErr := strconv.ParseInt(payload.ServerTime, 10, 64); parseErr == nil {
				return parsed, nil
			}
		}
	}

	return 0, fmt.Errorf("invalid server time response")
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, query url.Values, dest interface{}) error {
	return c.Call(ctx, http.MethodGet, path, query, nil, dest, nil)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}, dest interface{}) error {
	return c.Call(ctx, http.MethodPost, path, nil, body, dest, nil)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, body interface{}, dest interface{}) error {
	return c.Call(ctx, http.MethodDelete, path, nil, body, dest, nil)
}

// CallWithHeaders executes an HTTP request with custom headers.
func (c *Client) CallWithHeaders(ctx context.Context, method, path string, query url.Values, body interface{}, dest interface{}, headers map[string]string) error {
	return c.Call(ctx, method, path, query, body, dest, headers)
}
