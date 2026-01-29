package rtds

import (
	"context"
	"encoding/json"
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
			return
		}
		defer conn.Close()
		handler(conn)
	}))
}

func TestRtdsConnection(t *testing.T) {
	s := mockWSServer(t, func(c *websocket.Conn) {
		// Wait for sub
		_, _, _ = c.ReadMessage()
		// Keep alive
		select {}
	})
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	client, err := NewClient(wsURL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	// Wait for connect
	time.Sleep(100 * time.Millisecond)

	_, err = client.SubscribeCryptoPrices(context.Background(), []string{"BTC"})
	if err != nil {
		t.Errorf("Subscribe failed: %v", err)
	}
}

func TestRtdsReconnectLogic(t *testing.T) {
	client := &clientImpl{
		reconnect:    true,
		reconnectMax: 3,
	}
	if !client.shouldReconnect(1) {
		t.Errorf("should reconnect on attempt 1")
	}
}

func TestRtdsMessageUnmarshal(t *testing.T) {
	raw := `{"topic":"crypto_prices","type":"price","payload":{"symbol":"BTC","value":"50000"}}`
	var msg RtdsMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}