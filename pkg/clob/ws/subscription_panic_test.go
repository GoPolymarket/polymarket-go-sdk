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

// TestSubscriptionPanic_SendToClosedChannel tests that sending to a closed
// subscription channel doesn't cause a panic.
func TestSubscriptionPanic_SendToClosedChannel(t *testing.T) {
	sub := newTestSub(nil, nil)

	// Close the subscription
	sub.close()

	// Try to send to closed channel - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic occurred when sending to closed channel: %v", r)
		}
	}()

	// Multiple sends to closed channel
	for i := 0; i < 10; i++ {
		sub.trySend([]byte(`{"test":true}`))
	}
}

// TestSubscriptionPanic_ConcurrentCloseAndSend tests concurrent close and send
// operations to ensure no panics occur.
func TestSubscriptionPanic_ConcurrentCloseAndSend(t *testing.T) {
	sub := newTestSub(nil, nil)

	var wg sync.WaitGroup
	panicOccurred := false
	var panicMu sync.Mutex

	// Goroutines sending messages
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicMu.Lock()
					panicOccurred = true
					panicMu.Unlock()
					t.Errorf("panic in send goroutine %d: %v", n, r)
				}
			}()

			for j := 0; j < 50; j++ {
				sub.trySend([]byte(`{"test":true}`))
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Goroutines closing the subscription
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicMu.Lock()
					panicOccurred = true
					panicMu.Unlock()
					t.Errorf("panic in close goroutine %d: %v", n, r)
				}
			}()

			time.Sleep(25 * time.Millisecond)
			sub.close()
		}(i)
	}

	wg.Wait()

	if panicOccurred {
		t.Fatal("panic occurred during concurrent close and send operations")
	}
}

// TestSubscriptionPanic_FullChannelSend tests that sending to a full channel
// doesn't cause a panic and properly handles lag notification.
func TestSubscriptionPanic_FullChannelSend(t *testing.T) {
	sub := newTestSub(nil, nil)

	// Fill the channel
	for i := 0; i < cap(sub.ch); i++ {
		sub.ch <- []byte(`{"test":true}`)
	}

	// Try to send to full channel - should not panic, should notify lag
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic occurred when sending to full channel: %v", r)
		}
	}()

	sub.trySend([]byte(`{"overflow":true}`))

	// Check that lag was notified
	select {
	case err := <-sub.errCh:
		if _, ok := err.(LaggedError); !ok {
			t.Fatalf("expected LaggedError, got %T", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected lag notification")
	}
}

// TestSubscriptionPanic_MultipleCloseOperations tests that multiple close
// operations don't cause panics.
func TestSubscriptionPanic_MultipleCloseOperations(t *testing.T) {
	sub := newTestSub(nil, nil)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic occurred during multiple close operations: %v", r)
		}
	}()

	// Close multiple times
	for i := 0; i < 10; i++ {
		sub.close()
	}
}

// TestSubscriptionPanic_CloseWhileReading tests closing a subscription while
// another goroutine is reading from it.
func TestSubscriptionPanic_CloseWhileReading(t *testing.T) {
	sub := newTestSub(nil, nil)

	var wg sync.WaitGroup
	panicOccurred := false
	var panicMu sync.Mutex

	// Goroutine reading from channel
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				panicMu.Lock()
				panicOccurred = true
				panicMu.Unlock()
				t.Errorf("panic in reader goroutine: %v", r)
			}
		}()

		for {
			select {
			case _, ok := <-sub.ch:
				if !ok {
					return
				}
			case <-time.After(200 * time.Millisecond):
				return
			}
		}
	}()

	// Send some messages
	for i := 0; i < 5; i++ {
		sub.trySend([]byte(`{"test":true}`))
		time.Sleep(10 * time.Millisecond)
	}

	// Close while reader is active
	sub.close()

	wg.Wait()

	if panicOccurred {
		t.Fatal("panic occurred during close while reading")
	}
}

// TestSubscriptionPanic_DispatchToClosedSubscription tests that dispatching
// events to closed subscriptions doesn't cause panics.
func TestSubscriptionPanic_DispatchToClosedSubscription(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send some events
		for i := 0; i < 10; i++ {
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"price","market":"m1","price_changes":[{"asset_id":"test","price":"0.5"}]}`))
			time.Sleep(10 * time.Millisecond)
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

	// Create subscription
	stream, err := client.SubscribePricesStream(ctx, []string{"test"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Close subscription immediately
	stream.Close()

	// Wait for events to be dispatched to closed subscription
	time.Sleep(200 * time.Millisecond)

	// No panic should occur
}

// TestSubscriptionPanic_ConcurrentDispatchAndClose tests concurrent event
// dispatching and subscription closing.
func TestSubscriptionPanic_ConcurrentDispatchAndClose(t *testing.T) {
	c := newTestClient()

	var wg sync.WaitGroup
	panicOccurred := false
	var panicMu sync.Mutex

	// Create multiple subscriptions
	for i := 0; i < 10; i++ {
		entry := &subscriptionEntry[PriceChangeEvent]{
			id:      string(rune(i)),
			channel: ChannelMarket,
			event:   Price,
			ch:      make(chan PriceChangeEvent, 10),
			errCh:   make(chan error, 5),
		}
		c.priceSubs[entry.id] = entry
	}

	// Goroutines dispatching events
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicMu.Lock()
					panicOccurred = true
					panicMu.Unlock()
					t.Errorf("panic in dispatch goroutine: %v", r)
				}
			}()

			for j := 0; j < 100; j++ {
				event := PriceEvent{Market: "m1", PriceChanges: []PriceChangeEvent{{AssetID: "test", Price: "0.5"}}}
				c.dispatchPrice(event)
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// Goroutines closing subscriptions
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicMu.Lock()
					panicOccurred = true
					panicMu.Unlock()
					t.Errorf("panic in close goroutine: %v", r)
				}
			}()

			time.Sleep(50 * time.Millisecond)
			c.subMu.Lock()
			for _, sub := range c.priceSubs {
				sub.close()
			}
			c.subMu.Unlock()
		}()
	}

	wg.Wait()

	if panicOccurred {
		t.Fatal("panic occurred during concurrent dispatch and close")
	}
}

// TestSubscriptionPanic_NotifyLagOnClosedChannel tests that notifying lag
// on a closed error channel doesn't cause a panic.
func TestSubscriptionPanic_NotifyLagOnClosedChannel(t *testing.T) {
	sub := newTestSub(nil, nil)

	// Close the subscription (which closes errCh)
	sub.close()

	// Try to notify lag - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic occurred when notifying lag on closed channel: %v", r)
		}
	}()

	sub.notifyLag(5)
}
