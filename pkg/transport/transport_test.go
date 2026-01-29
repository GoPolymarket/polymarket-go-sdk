package transport

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockDoer implements Doer for testing
type MockDoer struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockDoer) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestClient_Call_Retry(t *testing.T) {
	// Override defaults for faster tests
	// We can't easily override constants in Go, so we'll just rely on the fact
	// that we mock the time or just accept a small delay.
	// Since defaultMinWait is 100ms, a few retries will take ~300ms, which is acceptable.

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

	t.Run("Retry on 500 then success", func(t *testing.T) {
		attempts := 0
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				attempts++
				if attempts < 3 {
					return &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader(`{"error":"server error"}`)),
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
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
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
		// defaultMaxRetries is 3, so attempts should be 0, 1, 2, 3 -> 4 attempts total?
		// Loop: attempt 0..3 (inclusive) = 4 iterations.
		if attempts != 4 {
			t.Errorf("expected 4 attempts, got %d", attempts)
		}
	})

	t.Run("Post body is preserved on retry", func(t *testing.T) {
		attempts := 0
		mock := &MockDoer{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				attempts++
				// Check body
				body, _ := io.ReadAll(req.Body)
				if string(body) != `{"foo":"bar"}` {
					return nil, io.ErrUnexpectedEOF // Fail if body doesn't match
				}

				if attempts == 1 {
					return &http.Response{
						StatusCode: 502,
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
				}, nil
			},
		}

		client := NewClient(mock, "http://example.com")
		payload := map[string]string{"foo": "bar"}
		err := client.Call(context.Background(), "POST", "/test", nil, payload, nil, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})
}
