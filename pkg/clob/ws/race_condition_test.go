package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestRaceCondition_ConcurrentGetConn tests concurrent access to getConn
func TestRaceCondition_ConcurrentGetConn(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	impl := client.(*clientImpl)

	// Concurrent reads of connection
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := impl.getConn(ChannelMarket)
			_ = conn // Use the connection
		}()
	}
	wg.Wait()
}

// TestRaceCondition_ConcurrentCloseAndRead tests closing connection while reads are in progress
func TestRaceCondition_ConcurrentCloseAndRead(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send periodic messages
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for i := 0; i < 20; i++ {
			<-ticker.C
			conn.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"price","asset_id":"test","price":"0.5"}`))
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	impl := client.(*clientImpl)

	// Start concurrent operations
	var wg sync.WaitGroup

	// Concurrent getConn calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				conn := impl.getConn(ChannelMarket)
				_ = conn
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	// Concurrent closeConn calls
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		impl.closeConn(ChannelMarket)
	}()

	wg.Wait()
	client.Close()
}

// TestRaceCondition_ConcurrentWriteJSON tests concurrent writes to WebSocket
func TestRaceCondition_ConcurrentWriteJSON(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read and discard messages
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
	defer client.Close()

	impl := client.(*clientImpl)

	// Concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			req := NewMarketSubscription([]string{"asset1"})
			_ = impl.writeJSON(ChannelMarket, req)
		}(i)
	}
	wg.Wait()
}

// TestRaceCondition_ConcurrentSubscriptionAccess tests concurrent access to subscription maps
func TestRaceCondition_ConcurrentSubscriptionAccess(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send periodic events
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for i := 0; i < 50; i++ {
			<-ticker.C
			conn.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"price","asset_id":"test","price":"0.5"}`))
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent subscribe operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			stream, err := client.SubscribePricesStream(ctx, []string{"test"})
			if err != nil {
				return
			}
			time.Sleep(20 * time.Millisecond)
			stream.Close()
		}(i)
	}

	wg.Wait()
}

// TestRaceCondition_ConcurrentStateAccess tests concurrent access to connection state
func TestRaceCondition_ConcurrentStateAccess(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(wsURL, nil, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	impl := client.(*clientImpl)

	var wg sync.WaitGroup

	// Concurrent state reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = impl.ConnectionState(ChannelMarket)
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// Concurrent state writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				impl.setConnState(ChannelMarket, ConnectionConnected, 0)
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
}

// TestRaceCondition_ConcurrentRefCounting tests concurrent access to ref counting maps
func TestRaceCondition_ConcurrentRefCounting(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read and discard
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
	defer client.Close()

	impl := client.(*clientImpl)

	var wg sync.WaitGroup

	// Concurrent add/remove operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			assets := []string{"asset1", "asset2"}
			impl.addMarketRefs(assets, false)
			time.Sleep(5 * time.Millisecond)
			impl.removeMarketRefs(assets)
		}(i)
	}

	wg.Wait()
}
