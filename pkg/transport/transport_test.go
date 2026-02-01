package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

// MockDoer implements Doer for testing
type MockDoer struct {
	DoFunc func(req *http.Request) (*http.Response, error)
	calls  []*http.Request
}

func (m *MockDoer) Do(req *http.Request) (*http.Response, error) {
	m.calls = append(m.calls, req)
	return m.DoFunc(req)
}

func TestClient_Call_Retry(t *testing.T) {
	t.Run("Success on first try", func(t *testing.T) {
		attempts := 0
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				attempts++
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
				}, nil
			},
		}

		client := NewClient(mock, "http://example.com")
		err := client.Call(context.Background(), "GET", "/test", nil, nil, nil, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("Retry on 429 then success", func(t *testing.T) {
		attempts := 0
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				attempts++
				if attempts == 1 {
					return &http.Response{
						StatusCode: 429,
						Body:       io.NopCloser(strings.NewReader(`{"error":"too many requests"}`)),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
				}, nil
			},
		}

		client := NewClient(mock, "http://example.com")
		err := client.Call(context.Background(), "GET", "/test", nil, nil, nil, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("Max retries exceeded", func(t *testing.T) {
		attempts := 0
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				attempts++
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader(`{"error":"server error"}`)),
				}, nil
			},
		}

		client := NewClient(mock, "http://example.com")
		err := client.Call(context.Background(), "GET", "/test", nil, nil, nil, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if attempts != 4 {
			t.Errorf("expected 4 attempts, got %d", attempts)
		}
	})
}

func TestClient_Call_CircuitBreakerIgnoresClientErrors(t *testing.T) {
	t.Run("4xx does not trip breaker", func(t *testing.T) {
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 400,
					Body:       io.NopCloser(strings.NewReader(`{"message":"bad request"}`)),
				}, nil
			},
		}

		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     1,
			ResetTimeout:    time.Second,
			HalfOpenMaxReqs: 1,
		})
		client := NewClient(mock, "http://example.com")
		client.SetCircuitBreaker(cb)

		err := client.Call(context.Background(), "GET", "/bad", nil, nil, nil, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var apiErr *types.Error
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *types.Error, got %T", err)
		}
		if cb.Failures() != 0 {
			t.Errorf("Failures() = %d, want 0", cb.Failures())
		}
		if cb.State() != StateClosed {
			t.Errorf("State() = %v, want %v", cb.State(), StateClosed)
		}
	})

	t.Run("context cancellation does not trip breaker", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, ctx.Err()
			},
		}

		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     1,
			ResetTimeout:    time.Second,
			HalfOpenMaxReqs: 1,
		})
		client := NewClient(mock, "http://example.com")
		client.SetCircuitBreaker(cb)

		err := client.Call(ctx, "GET", "/canceled", nil, nil, nil, nil)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
		if cb.Failures() != 0 {
			t.Errorf("Failures() = %d, want 0", cb.Failures())
		}
		if cb.State() != StateClosed {
			t.Errorf("State() = %v, want %v", cb.State(), StateClosed)
		}
	})
}

func TestClientHelpers(t *testing.T) {
	ctx := context.Background()

	t.Run("Get", func(t *testing.T) {
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"id":"1"}`)),
				}, nil
			},
		}
		client := NewClient(mock, "http://example.com")
		var resp map[string]string
		err := client.Get(ctx, "/get", nil, &resp)
		if err != nil || resp["id"] != "1" {
			t.Errorf("Get failed: %v", err)
		}
	})

	t.Run("Post", func(t *testing.T) {
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
				}, nil
			},
		}
		client := NewClient(mock, "http://example.com")
		err := client.Post(ctx, "/post", nil, nil)
		if err != nil {
			t.Errorf("Post failed: %v", err)
		}
	})

	t.Run("Clone", func(t *testing.T) {
		client := NewClient(http.DefaultClient, "http://example.com")
		clone := client.CloneWithBaseURL("http://new.com")
		if clone.baseURL != "http://new.com" {
			t.Errorf("clone failed")
		}
	})

	t.Run("Clone preserves resilience settings", func(t *testing.T) {
		// Create client with rate limiter and circuit breaker
		client := NewClientWithResilience(http.DefaultClient, "http://example.com", 10, DefaultCircuitBreakerConfig())
		client.SetUserAgent("test-agent")
		client.SetUseServerTime(true)

		// Clone with new base URL
		clone := client.CloneWithBaseURL("http://new.com")

		// Verify base URL changed
		if clone.baseURL != "http://new.com" {
			t.Errorf("expected baseURL http://new.com, got %s", clone.baseURL)
		}

		// Verify resilience settings preserved
		if clone.rateLimiter == nil {
			t.Error("rate limiter not preserved in clone")
		}
		if clone.circuitBreaker == nil {
			t.Error("circuit breaker not preserved in clone")
		}
		if clone.rateLimiter != client.rateLimiter {
			t.Error("rate limiter should be shared between original and clone")
		}
		if clone.circuitBreaker != client.circuitBreaker {
			t.Error("circuit breaker should be shared between original and clone")
		}

		// Verify other settings preserved
		if clone.userAgent != client.userAgent {
			t.Errorf("userAgent not preserved: expected %s, got %s", client.userAgent, clone.userAgent)
		}
		if clone.useServerTime != client.useServerTime {
			t.Error("useServerTime not preserved")
		}
	})

	t.Run("Setters", func(t *testing.T) {
		client := NewClient(http.DefaultClient, "http://example.com")
		client.SetUserAgent("ua")
		client.SetUseServerTime(true)
		client.SetAuth(nil, nil)
		client.SetBuilderConfig(nil)
	})

	t.Run("CallWithHeaders", func(t *testing.T) {
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("X-Test") != "val" {
					return nil, fmt.Errorf("missing header")
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{}`)),
				}, nil
			},
		}
		client := NewClient(mock, "http://example.com")
		err := client.CallWithHeaders(ctx, "GET", "/", nil, nil, nil, map[string]string{"X-Test": "val"})
		if err != nil {
			t.Errorf("CallWithHeaders failed: %v", err)
		}
	})

	t.Run("ServerTime", func(t *testing.T) {
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`12345`)),
				}, nil
			},
		}
		client := NewClient(mock, "http://example.com")
		ts, err := client.serverTime(ctx)
		if err != nil || ts != 12345 {
			t.Errorf("serverTime failed: %v", err)
		}
	})
}

func TestMarshalBody(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected string
	}{
		{map[string]string{"a": "b"}, `{"a":"b"}`},
		{"string", "string"},
		{[]byte("bytes"), "bytes"},
		{nil, ""},
	}

	for _, c := range cases {
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				var body []byte
				if req.Body != nil {
					body, _ = io.ReadAll(req.Body)
				}
				if string(body) != c.expected {
					t.Errorf("expected %q, got %q", c.expected, string(body))
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}"))}, nil
			},
		}
		client := NewClient(mock, "http://example.com")
		_ = client.Post(context.Background(), "/", c.input, nil)
	}
}
