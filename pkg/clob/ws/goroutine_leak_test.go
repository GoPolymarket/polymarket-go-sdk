package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/goleak"
)

// TestWebSocketGoroutineLeaks tests that goroutines are properly cleaned up
// during reconnection scenarios to prevent goroutine leaks.
func TestWebSocketGoroutineLeaks_Reconnection(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	// Create a mock WebSocket server that closes connections immediately
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Close immediately to trigger reconnection
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Setenv("CLOB_WS_RECONNECT_DELAY_MS", "10")
	t.Setenv("CLOB_WS_RECONNECT_MAX_DELAY_MS", "50")
	t.Setenv("CLOB_WS_RECONNECT_MAX", "2")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Set short timeouts to speed up test
	impl := client.(*clientImpl)
	impl.setReadTimeout(100 * time.Millisecond)

	// Wait for reconnection attempts
	time.Sleep(200 * time.Millisecond)

	// Close the client
	if err := client.Close(); err != nil {
		t.Fatalf("failed to close client: %v", err)
	}

	// Give goroutines time to exit
	time.Sleep(100 * time.Millisecond)

	// goleak.VerifyNone will check for leaked goroutines at test end
}

// TestWebSocketGoroutineLeaks_MultipleReconnections tests that multiple
// reconnection cycles don't accumulate goroutines.
func TestWebSocketGoroutineLeaks_MultipleReconnections(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var connectionCount int32
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		atomic.AddInt32(&connectionCount, 1)
		// Close after a short delay
		time.AfterFunc(20*time.Millisecond, func() {
			conn.Close()
		})
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Setenv("CLOB_WS_RECONNECT_DELAY_MS", "10")
	t.Setenv("CLOB_WS_RECONNECT_MAX_DELAY_MS", "50")
	t.Setenv("CLOB_WS_RECONNECT_MAX", "5")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	impl := client.(*clientImpl)
	impl.setReadTimeout(50 * time.Millisecond)

	// Wait for multiple reconnection cycles
	time.Sleep(500 * time.Millisecond)

	if err := client.Close(); err != nil {
		t.Fatalf("failed to close client: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	count := atomic.LoadInt32(&connectionCount)
	if count < 2 {
		t.Logf("Warning: only %d connections made, expected multiple reconnections", count)
	}
}

// TestWebSocketGoroutineLeaks_CloseWhileReading tests that closing the client
// while a read is in progress doesn't leak goroutines.
func TestWebSocketGoroutineLeaks_CloseWhileReading(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	done := make(chan struct{})
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Keep connection open but don't send data
		// This simulates a hanging read
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}))
	defer server.Close()
	defer close(done)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	impl := client.(*clientImpl)
	impl.setReadTimeout(100 * time.Millisecond)

	// Give readLoop time to start
	time.Sleep(50 * time.Millisecond)

	// Close while read is potentially in progress
	if err := client.Close(); err != nil {
		t.Fatalf("failed to close client: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

// TestWebSocketGoroutineLeaks_PingLoopCleanup tests that pingLoop goroutines
// are properly cleaned up when connections are closed.
func TestWebSocketGoroutineLeaks_PingLoopCleanup(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Keep connection alive and respond to pings
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if string(msg) == "PING" {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("PONG"))
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Wait for ping loop to start and send some pings
	time.Sleep(100 * time.Millisecond)

	if err := client.Close(); err != nil {
		t.Fatalf("failed to close client: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

// TestWebSocketGoroutineLeaks_SubscriptionCleanup tests that subscription
// goroutines are properly cleaned up.
func TestWebSocketGoroutineLeaks_SubscriptionCleanup(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read subscription requests and keep connection alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Create multiple subscriptions
	ctx := context.Background()
	stream1, err := client.SubscribeOrderbookStream(ctx, []string{"asset1"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	stream2, err := client.SubscribePricesStream(ctx, []string{"asset2"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Close streams
	stream1.Close()
	stream2.Close()

	// Close client
	if err := client.Close(); err != nil {
		t.Fatalf("failed to close client: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}
