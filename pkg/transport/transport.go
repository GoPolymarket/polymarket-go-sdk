// Package transport provides a robust HTTP client wrapper optimized for Polymarket's API.
// It handles automatic L2 HMAC authentication, builder attribution, retries with
// exponential backoff, and unified error handling.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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

// Doer defines the interface for executing an HTTP request.
// It matches the standard *http.Client's Do method.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a high-level wrapper around an HTTP Doer.
// It adds Polymarket-specific functionality like automatic HMAC signing
// and transparent request retries for ephemeral server errors.
type Client struct {
	httpClient     Doer
	baseURL        string
	userAgent      string
	signer         auth.Signer
	apiKey         *auth.APIKey
	builder        *auth.BuilderConfig
	useServerTime  bool
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreaker
}

// NewClient creates a new transport client.
// If httpClient is nil, http.DefaultClient will be used.
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

// NewClientWithRateLimiter creates a new transport client with rate limiting enabled.
func NewClientWithRateLimiter(httpClient Doer, baseURL string, requestsPerSecond int) *Client {
	client := NewClient(httpClient, baseURL)
	client.rateLimiter = NewRateLimiter(requestsPerSecond)
	client.rateLimiter.Start()
	return client
}

// NewClientWithCircuitBreaker creates a new transport client with circuit breaker enabled.
func NewClientWithCircuitBreaker(httpClient Doer, baseURL string, config CircuitBreakerConfig) *Client {
	client := NewClient(httpClient, baseURL)
	client.circuitBreaker = NewCircuitBreaker(config)
	return client
}

// NewClientWithResilience creates a new transport client with both rate limiting and circuit breaker.
func NewClientWithResilience(httpClient Doer, baseURL string, requestsPerSecond int, cbConfig CircuitBreakerConfig) *Client {
	client := NewClient(httpClient, baseURL)
	client.rateLimiter = NewRateLimiter(requestsPerSecond)
	client.rateLimiter.Start()
	client.circuitBreaker = NewCircuitBreaker(cbConfig)
	return client
}

// SetRateLimiter sets the rate limiter for the client.
func (c *Client) SetRateLimiter(rl *RateLimiter) {
	c.rateLimiter = rl
}

// SetCircuitBreaker sets the circuit breaker for the client.
func (c *Client) SetCircuitBreaker(cb *CircuitBreaker) {
	c.circuitBreaker = cb
}

// CloneWithBaseURL creates a new client sharing the same underlying HTTP Doer
// but targeting a different base URL (e.g., for specialized sub-services).
// All settings including rate limiter, circuit breaker, and auth are preserved.
func (c *Client) CloneWithBaseURL(baseURL string) *Client {
	if c == nil {
		return NewClient(nil, baseURL)
	}
	clone := NewClient(c.httpClient, baseURL)
	clone.userAgent = c.userAgent
	clone.useServerTime = c.useServerTime
	clone.signer = c.signer
	clone.apiKey = c.apiKey
	clone.builder = c.builder
	clone.rateLimiter = c.rateLimiter
	clone.circuitBreaker = c.circuitBreaker
	return clone
}

// SetUserAgent sets the User-Agent header value for all subsequent requests.
func (c *Client) SetUserAgent(userAgent string) {
	if userAgent != "" {
		c.userAgent = userAgent
	}
}

// SetAuth configures the client with credentials for Layer 2 HMAC authentication.
func (c *Client) SetAuth(signer auth.Signer, apiKey *auth.APIKey) {
	c.signer = signer
	c.apiKey = apiKey
}

// SetBuilderConfig configures the client for builder attribution headers.
func (c *Client) SetBuilderConfig(config *auth.BuilderConfig) {
	c.builder = config
}

// SetUseServerTime enables or disables server-time synchronization for signatures.
func (c *Client) SetUseServerTime(use bool) {
	c.useServerTime = use
}

// Call is the core method for executing HTTP requests.
// It handles payload serialization, authentication header injection, and retry logic.
// Retryable errors include HTTP 429 (Rate Limit) and 5xx (Server Error).
func (c *Client) Call(ctx context.Context, method, path string, query url.Values, body interface{}, dest interface{}, headers map[string]string) error {
	// Apply circuit breaker if configured
	if c.circuitBreaker != nil {
		return c.circuitBreaker.CallWithFailurePredicate(func() error {
			// Apply rate limiting only after breaker allows the request.
			if c.rateLimiter != nil {
				if err := c.rateLimiter.Wait(ctx); err != nil {
					return fmt.Errorf("rate limiter: %w", err)
				}
			}
			return c.doCall(ctx, method, path, query, body, dest, headers)
		}, shouldCountCircuitBreakerFailure)
	}

	// Apply rate limiting if configured (no circuit breaker).
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter: %w", err)
		}
	}

	return c.doCall(ctx, method, path, query, body, dest, headers)
}

func shouldCountCircuitBreakerFailure(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var apiErr *types.Error
	if errors.As(err, &apiErr) {
		if apiErr.Status >= 400 && apiErr.Status < 500 {
			return false
		}
	}
	var statusErr interface{ StatusCode() int }
	if errors.As(err, &statusErr) {
		status := statusErr.StatusCode()
		if status >= 400 && status < 500 {
			return false
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return false
	}
	return true
}

type httpStatusError struct {
	status int
	body   string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("server error %d: %s", e.status, e.body)
}

func (e *httpStatusError) StatusCode() int {
	return e.status
}

// doCall performs the actual HTTP request without rate limiting or circuit breaker.
func (c *Client) doCall(ctx context.Context, method, path string, query url.Values, body interface{}, dest interface{}, headers map[string]string) error {
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
				lastErr = &httpStatusError{status: resp.StatusCode, body: string(respBytes)}
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

// Get performs a GET request with automatic L2 authentication if credentials are provided.
func (c *Client) Get(ctx context.Context, path string, query url.Values, dest interface{}) error {
	return c.Call(ctx, http.MethodGet, path, query, nil, dest, nil)
}

// Post performs a POST request with automatic L2 authentication if credentials are provided.
func (c *Client) Post(ctx context.Context, path string, body interface{}, dest interface{}) error {
	return c.Call(ctx, http.MethodPost, path, nil, body, dest, nil)
}

// Delete performs a DELETE request with automatic L2 authentication if credentials are provided.
func (c *Client) Delete(ctx context.Context, path string, body interface{}, dest interface{}) error {
	return c.Call(ctx, http.MethodDelete, path, nil, body, dest, nil)
}

// CallWithHeaders executes an HTTP request with custom headers and automatic L2 authentication.
func (c *Client) CallWithHeaders(ctx context.Context, method, path string, query url.Values, body interface{}, dest interface{}, headers map[string]string) error {
	return c.Call(ctx, method, path, query, body, dest, headers)
}
