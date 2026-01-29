package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func mockWSServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		handler(conn)
	}))
}

func TestClientConnection(t *testing.T) {
	// Simple test: connect and receive one message
	s := mockWSServer(t, func(c *websocket.Conn) {
		// Wait for subscription request
		_, _, _ = c.ReadMessage()

		// Send a dummy event
		err := c.WriteJSON(map[string]string{
			"event_type": "price",
			"asset_id":   "123",
			"price":      "0.5",
		})
		if err != nil {
			return
		}
		// Keep alive
		time.Sleep(1 * time.Second)
	})
	defer s.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)
	if client.ConnectionState(ChannelMarket) != ConnectionConnected {
		t.Errorf("expected connected, got %v", client.ConnectionState(ChannelMarket))
	}

	sub, err := client.SubscribePrices(context.Background(), []string{"123"})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	select {
	case event := <-sub:
		if event.AssetID != "123" {
			t.Errorf("expected asset 123, got %s", event.AssetID)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for event")
	}
}

func TestClientReadTimeout(t *testing.T) {
	// Save original timeout and restore after test
	originalTimeout := ReadTimeout
	ReadTimeout = 100 * time.Millisecond
	defer func() { ReadTimeout = originalTimeout }()

	connections := make(chan struct{}, 10)

	s := mockWSServer(t, func(c *websocket.Conn) {
		connections <- struct{}{}
		// Hang: don't send anything. Client should timeout after 100ms.
		select {}
	})
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	// Set reconnect max to allow multiple reconnects
	// We can't easily set reconnect options via NewClient without env vars,
	// but defaults are usually fine (reconnect=true).
	
	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	// 1. First connection
	select {
	case <-connections:
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for first connection")
	}

	// 2. Client should timeout (100ms) + reconnect delay (default is 2s, which is too long for this test)
	// We need to override reconnect delay? 
	// The clientImpl reads CLOB_WS_RECONNECT_DELAY_MS from env.
	// But it reads it in NewClient. We can't set it easily now.
	// However, we can verify that the connection drops.
	
	time.Sleep(200 * time.Millisecond) // Wait for timeout

	// The client should have closed the connection by now.
	// We check if it reconnects.
	// Since default reconnect delay is 2s, we might need to wait > 2s.
	// That's acceptable for a test.
	
	select {
	case <-connections:
		// Reconnected!
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for reconnection")
	}
}
