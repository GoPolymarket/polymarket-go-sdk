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
// --------------- newTestClient helper ---------------

func newTestClient() *clientImpl {
	ready := make(chan struct{})
	close(ready)
	return &clientImpl{
		done:       make(chan struct{}),
		connReady:  ready,
		stateSubs:  make(map[string]*stateSubscription),
		subRefs:    make(map[string]int),
		subDetails: make(map[string]Subscription),
		subs:       make(map[string]*subscriptionEntry),
		subsByKey:  make(map[string]map[string]*subscriptionEntry),
	}
}

// --------------- parseMessages ---------------

func TestParseMessages_Empty(t *testing.T) {
	msgs, err := parseMessages([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}
}

func TestParseMessages_Whitespace(t *testing.T) {
	msgs, err := parseMessages([]byte("   "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}
}

func TestParseMessages_SingleObject(t *testing.T) {
	raw := `{"topic":"crypto_prices","type":"update","timestamp":123,"payload":{"symbol":"BTC"}}`
	msgs, err := parseMessages([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Topic != "crypto_prices" {
		t.Fatalf("expected crypto_prices, got %s", msgs[0].Topic)
	}
}

func TestParseMessages_Array(t *testing.T) {
	raw := `[{"topic":"crypto_prices","type":"update","payload":{}},{"topic":"comments","type":"comment_created","payload":{}}]`
	msgs, err := parseMessages([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
}

func TestParseMessages_InvalidJSON(t *testing.T) {
	_, err := parseMessages([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseMessages_InvalidArray(t *testing.T) {
	_, err := parseMessages([]byte("[not json]"))
	if err == nil {
		t.Fatal("expected error for invalid array")
	}
}

// --------------- subscriptionEntry ---------------

func TestSubscriptionEntry_Matches(t *testing.T) {
	entry := &subscriptionEntry{
		topic:   "crypto_prices",
		msgType: "update",
	}
	msg := RtdsMessage{Topic: "crypto_prices", MsgType: "update"}
	if !entry.matches(msg) {
		t.Fatal("should match")
	}
}

func TestSubscriptionEntry_Matches_WrongTopic(t *testing.T) {
	entry := &subscriptionEntry{
		topic:   "crypto_prices",
		msgType: "update",
	}
	msg := RtdsMessage{Topic: "comments", MsgType: "update"}
	if entry.matches(msg) {
		t.Fatal("should not match wrong topic")
	}
}

func TestSubscriptionEntry_Matches_WrongMsgType(t *testing.T) {
	entry := &subscriptionEntry{
		topic:   "crypto_prices",
		msgType: "update",
	}
	msg := RtdsMessage{Topic: "crypto_prices", MsgType: "snapshot"}
	if entry.matches(msg) {
		t.Fatal("should not match wrong msg type")
	}
}

func TestSubscriptionEntry_Matches_Wildcard(t *testing.T) {
	entry := &subscriptionEntry{
		topic:   "comments",
		msgType: "*",
	}
	msg := RtdsMessage{Topic: "comments", MsgType: "comment_created"}
	if !entry.matches(msg) {
		t.Fatal("wildcard should match any msg type")
	}
}

func TestSubscriptionEntry_Matches_WithFilter(t *testing.T) {
	entry := &subscriptionEntry{
		topic:   "crypto_prices",
		msgType: "update",
		filter: func(msg RtdsMessage) bool {
			return strings.Contains(string(msg.Payload), "BTC")
		},
	}
	msg := RtdsMessage{Topic: "crypto_prices", MsgType: "update", Payload: json.RawMessage(`{"symbol":"BTC"}`)}
	if !entry.matches(msg) {
		t.Fatal("should match with filter")
	}
	msg2 := RtdsMessage{Topic: "crypto_prices", MsgType: "update", Payload: json.RawMessage(`{"symbol":"ETH"}`)}
	if entry.matches(msg2) {
		t.Fatal("should not match with filter")
	}
}

func TestSubscriptionEntry_TrySend(t *testing.T) {
	entry := &subscriptionEntry{
		ch:    make(chan RtdsMessage, 5),
		errCh: make(chan error, 5),
	}
	msg := RtdsMessage{Topic: "test"}
	entry.trySend(msg)
	select {
	case got := <-entry.ch:
		if got.Topic != "test" {
			t.Fatalf("expected test, got %s", got.Topic)
		}
	default:
		t.Fatal("expected message")
	}
}

func TestSubscriptionEntry_TrySend_Closed(t *testing.T) {
	entry := &subscriptionEntry{
		ch:    make(chan RtdsMessage, 5),
		errCh: make(chan error, 5),
	}
	entry.close()
	// Should not panic
	entry.trySend(RtdsMessage{})
}

func TestSubscriptionEntry_TrySend_FullChannel(t *testing.T) {
	entry := &subscriptionEntry{
		topic:   "test",
		msgType: "update",
		ch:      make(chan RtdsMessage, 1),
		errCh:   make(chan error, 5),
	}
	entry.trySend(RtdsMessage{Topic: "test"}) // fills channel
	entry.trySend(RtdsMessage{Topic: "test"}) // should trigger lag

	select {
	case err := <-entry.errCh:
		le, ok := err.(LaggedError)
		if !ok {
			t.Fatalf("expected LaggedError, got %T", err)
		}
		if le.Count != 1 {
			t.Fatalf("expected count 1, got %d", le.Count)
		}
	default:
		t.Fatal("expected lag error")
	}
}

func TestSubscriptionEntry_NotifyLag_Zero(t *testing.T) {
	entry := &subscriptionEntry{
		errCh: make(chan error, 5),
	}
	entry.notifyLag(0)
	select {
	case <-entry.errCh:
		t.Fatal("should not send for count 0")
	default:
	}
}

func TestSubscriptionEntry_Close_Idempotent(t *testing.T) {
	entry := &subscriptionEntry{
		ch:    make(chan RtdsMessage, 1),
		errCh: make(chan error, 1),
	}
	entry.close()
	entry.close() // should not panic
}

// --------------- subscriptionKey ---------------

func TestSubscriptionKey(t *testing.T) {
	key := subscriptionKey("crypto_prices", "update")
	if key != "crypto_prices|update" {
		t.Fatalf("expected crypto_prices|update, got %s", key)
	}
}

// --------------- symbolSet ---------------

func TestSymbolSet_Nil(t *testing.T) {
	if symbolSet(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestSymbolSet_Empty(t *testing.T) {
	if symbolSet([]string{}) != nil {
		t.Fatal("expected nil")
	}
}

func TestSymbolSet_AllEmpty(t *testing.T) {
	if symbolSet([]string{"", "  "}) != nil {
		t.Fatal("expected nil for all-empty")
	}
}

func TestSymbolSet_Normal(t *testing.T) {
	set := symbolSet([]string{"BTC", "eth", " SOL "})
	if len(set) != 3 {
		t.Fatalf("expected 3, got %d", len(set))
	}
	if _, ok := set["btc"]; !ok {
		t.Fatal("expected lowercase btc")
	}
	if _, ok := set["eth"]; !ok {
		t.Fatal("expected eth")
	}
	if _, ok := set["sol"]; !ok {
		t.Fatal("expected trimmed sol")
	}
}

// --------------- shouldReconnect ---------------

func TestShouldReconnect_Disabled(t *testing.T) {
	c := newTestClient()
	c.reconnect = false
	if c.shouldReconnect(0) {
		t.Fatal("should not reconnect when disabled")
	}
}

func TestShouldReconnect_UnlimitedRetries(t *testing.T) {
	c := newTestClient()
	c.reconnect = true
	c.reconnectMax = 0
	if !c.shouldReconnect(100) {
		t.Fatal("should reconnect with unlimited retries")
	}
}

func TestShouldReconnect_WithinMax(t *testing.T) {
	c := newTestClient()
	c.reconnect = true
	c.reconnectMax = 5
	if !c.shouldReconnect(4) {
		t.Fatal("should reconnect within max")
	}
}

func TestShouldReconnect_ExceedsMax(t *testing.T) {
	c := newTestClient()
	c.reconnect = true
	c.reconnectMax = 5
	if c.shouldReconnect(5) {
		t.Fatal("should not reconnect at max")
	}
}

// --------------- ConnectionState ---------------

func TestConnectionState_Disconnected(t *testing.T) {
	c := newTestClient()
	if c.ConnectionState() != ConnectionDisconnected {
		t.Fatal("expected disconnected")
	}
}

func TestConnectionState_Connected(t *testing.T) {
	c := newTestClient()
	c.setState(ConnectionConnected)
	if c.ConnectionState() != ConnectionConnected {
		t.Fatal("expected connected")
	}
}

func TestSetState_BackToDisconnected(t *testing.T) {
	c := newTestClient()
	c.setState(ConnectionConnected)
	c.setState(ConnectionDisconnected)
	if c.ConnectionState() != ConnectionDisconnected {
		t.Fatal("expected disconnected")
	}
}

// --------------- SubscriptionCount ---------------

func TestSubscriptionCount_Empty(t *testing.T) {
	c := newTestClient()
	if c.SubscriptionCount() != 0 {
		t.Fatal("expected 0")
	}
}

// --------------- dispatch ---------------

func TestDispatch_MatchingSub(t *testing.T) {
	c := newTestClient()
	ch := make(chan RtdsMessage, 5)
	c.subs["s1"] = &subscriptionEntry{
		id:      "s1",
		topic:   "crypto_prices",
		msgType: "update",
		ch:      ch,
		errCh:   make(chan error, 5),
	}

	msg := RtdsMessage{Topic: "crypto_prices", MsgType: "update", Payload: json.RawMessage(`{}`)}
	c.dispatch(msg)

	select {
	case got := <-ch:
		if got.Topic != "crypto_prices" {
			t.Fatalf("expected crypto_prices, got %s", got.Topic)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestDispatch_NonMatchingSub(t *testing.T) {
	c := newTestClient()
	ch := make(chan RtdsMessage, 5)
	c.subs["s1"] = &subscriptionEntry{
		id:      "s1",
		topic:   "comments",
		msgType: "update",
		ch:      ch,
		errCh:   make(chan error, 5),
	}

	msg := RtdsMessage{Topic: "crypto_prices", MsgType: "update"}
	c.dispatch(msg)

	select {
	case <-ch:
		t.Fatal("should not receive non-matching message")
	case <-time.After(50 * time.Millisecond):
	}
}

// --------------- Stream ---------------

func TestStream_Close_Nil(t *testing.T) {
	var s *Stream[int]
	if err := s.Close(); err != nil {
		t.Fatalf("nil stream close should not error: %v", err)
	}
}

func TestStream_Close_NilCloseF(t *testing.T) {
	s := &Stream[int]{closeF: nil}
	if err := s.Close(); err != nil {
		t.Fatalf("nil closeF should not error: %v", err)
	}
}

func TestStream_Close_Normal(t *testing.T) {
	called := false
	s := &Stream[int]{closeF: func() error { called = true; return nil }}
	if err := s.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("closeF not called")
	}
}

// --------------- LaggedError ---------------

func TestLaggedError_WithTopic(t *testing.T) {
	e := LaggedError{Count: 5, Topic: "crypto_prices", MsgType: "update"}
	s := e.Error()
	if !strings.Contains(s, "5") || !strings.Contains(s, "crypto_prices") {
		t.Fatalf("unexpected error string: %s", s)
	}
}

func TestLaggedError_WithoutTopic(t *testing.T) {
	e := LaggedError{Count: 2}
	s := e.Error()
	if !strings.Contains(s, "2") || strings.Contains(s, "topic=") {
		t.Fatalf("unexpected error string: %s", s)
	}
}

// --------------- Authenticate / Deauthenticate ---------------

func TestAuthenticate(t *testing.T) {
	c := newTestClient()
	ret := c.Authenticate(nil)
	if ret != c {
		t.Fatal("expected same client")
	}
}

func TestDeauthenticate(t *testing.T) {
	c := newTestClient()
	ret := c.Deauthenticate()
	if ret != c {
		t.Fatal("expected same client")
	}
	if c.auth != nil {
		t.Fatal("auth should be nil")
	}
}

// --------------- setState with state subscribers ---------------

func TestSetState_NotifiesSubscribers(t *testing.T) {
	c := newTestClient()
	ch := make(chan ConnectionStateEvent, 10)
	c.stateSubs["s1"] = &stateSubscription{
		id:    "s1",
		ch:    ch,
		errCh: make(chan error, 5),
	}

	c.setState(ConnectionConnected)

	select {
	case ev := <-ch:
		if ev.State != ConnectionConnected {
			t.Fatalf("expected connected, got %s", ev.State)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

// --------------- stateSubscription ---------------

func TestStateSubscription_TrySend(t *testing.T) {
	sub := &stateSubscription{
		ch:    make(chan ConnectionStateEvent, 5),
		errCh: make(chan error, 5),
	}
	sub.trySend(ConnectionStateEvent{State: ConnectionConnected})
	select {
	case ev := <-sub.ch:
		if ev.State != ConnectionConnected {
			t.Fatalf("expected connected, got %s", ev.State)
		}
	default:
		t.Fatal("expected event")
	}
}

func TestStateSubscription_TrySend_Closed(t *testing.T) {
	sub := &stateSubscription{
		ch:    make(chan ConnectionStateEvent, 5),
		errCh: make(chan error, 5),
	}
	sub.close()
	sub.trySend(ConnectionStateEvent{}) // should not panic
}

func TestStateSubscription_Close_Idempotent(t *testing.T) {
	sub := &stateSubscription{
		ch:    make(chan ConnectionStateEvent, 1),
		errCh: make(chan error, 1),
	}
	first := sub.close()
	if !first {
		t.Fatal("first close should return true")
	}
	second := sub.close()
	if second {
		t.Fatal("second close should return false")
	}
}

func TestStateSubscription_NotifyLag_Zero(t *testing.T) {
	sub := &stateSubscription{
		errCh: make(chan error, 5),
	}
	sub.notifyLag(0)
	select {
	case <-sub.errCh:
		t.Fatal("should not send for count 0")
	default:
	}
}

func TestStateSubscription_NotifyLag_Normal(t *testing.T) {
	sub := &stateSubscription{
		errCh: make(chan error, 5),
	}
	sub.notifyLag(3)
	select {
	case err := <-sub.errCh:
		le, ok := err.(LaggedError)
		if !ok {
			t.Fatalf("expected LaggedError, got %T", err)
		}
		if le.Count != 3 {
			t.Fatalf("expected count 3, got %d", le.Count)
		}
	default:
		t.Fatal("expected error")
	}
}

// --------------- ConnectionStateStream ---------------

func TestConnectionStateStream(t *testing.T) {
	c := newTestClient()
	stream, err := c.ConnectionStateStream(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer stream.Close()

	select {
	case ev := <-stream.C:
		if ev.State != ConnectionDisconnected {
			t.Fatalf("expected initial disconnected, got %s", ev.State)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for initial state")
	}
}

// --------------- closeAllSubscriptions ---------------

func TestCloseAllSubscriptions(t *testing.T) {
	c := newTestClient()
	entry := &subscriptionEntry{
		id:      "s1",
		key:     "test|update",
		topic:   "test",
		msgType: "update",
		ch:      make(chan RtdsMessage, 1),
		errCh:   make(chan error, 1),
	}
	c.subs["s1"] = entry
	c.subsByKey["test|update"] = map[string]*subscriptionEntry{"s1": entry}
	c.subRefs["test|update"] = 1
	c.subDetails["test|update"] = Subscription{Topic: "test", MsgType: "update"}

	c.closeAllSubscriptions()

	if len(c.subs) != 0 {
		t.Fatal("expected empty subs")
	}
	if len(c.subsByKey) != 0 {
		t.Fatal("expected empty subsByKey")
	}
	if len(c.subRefs) != 0 {
		t.Fatal("expected empty subRefs")
	}
}

// --------------- closeStateSubscriptions ---------------

func TestCloseStateSubscriptions(t *testing.T) {
	c := newTestClient()
	sub := &stateSubscription{
		id:    "s1",
		ch:    make(chan ConnectionStateEvent, 1),
		errCh: make(chan error, 1),
	}
	c.stateSubs["s1"] = sub

	c.closeStateSubscriptions()

	if len(c.stateSubs) != 0 {
		t.Fatal("expected empty stateSubs")
	}
}

// --------------- sendSubscriptions ---------------

func TestSendSubscriptions_Empty(t *testing.T) {
	c := newTestClient()
	err := c.sendSubscriptions(SubscribeAction, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendSubscriptions_NoConn(t *testing.T) {
	c := newTestClient()
	err := c.sendSubscriptions(SubscribeAction, []Subscription{{Topic: "test", MsgType: "update"}})
	if err == nil {
		t.Fatal("expected error with no connection")
	}
}

// --------------- writeJSON ---------------

func TestWriteJSON_NoConn(t *testing.T) {
	c := newTestClient()
	err := c.writeJSON(map[string]string{"test": "value"})
	if err == nil {
		t.Fatal("expected error with no connection")
	}
}

// --------------- SubscribeRawStream / UnsubscribeRaw ---------------

func TestSubscribeRawStream_NilSub(t *testing.T) {
	c := newTestClient()
	_, err := c.SubscribeRawStream(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil subscription")
	}
}

func TestUnsubscribeRaw_NilSub(t *testing.T) {
	c := newTestClient()
	err := c.UnsubscribeRaw(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil subscription")
	}
}

func TestSubscribeRaw_EmptyTopic(t *testing.T) {
	c := newTestClient()
	_, err := c.subscribeRaw(Subscription{Topic: "", MsgType: "update"}, nil)
	if err == nil {
		t.Fatal("expected error for empty topic")
	}
}

func TestSubscribeRaw_EmptyMsgType(t *testing.T) {
	c := newTestClient()
	_, err := c.subscribeRaw(Subscription{Topic: "test", MsgType: ""}, nil)
	if err == nil {
		t.Fatal("expected error for empty msg type")
	}
}

// --------------- unsubscribeByID ---------------

func TestUnsubscribeByID_NotFound(t *testing.T) {
	c := newTestClient()
	err := c.unsubscribeByID("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --------------- unsubscribeTopic ---------------

func TestUnsubscribeTopic_NotFound(t *testing.T) {
	c := newTestClient()
	err := c.unsubscribeTopic("nonexistent", "update")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --------------- Subscription MarshalJSON ---------------

func TestSubscriptionMarshalJSON_NilFilters(t *testing.T) {
	sub := Subscription{Topic: "test", MsgType: "update"}
	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(data), "filters") {
		t.Fatalf("should not contain filters: %s", string(data))
	}
}

func TestSubscriptionMarshalJSON_WithClobAuth(t *testing.T) {
	sub := Subscription{
		Topic:    "comments",
		MsgType:  "*",
		ClobAuth: &ClobAuth{Key: "k", Secret: "s", Passphrase: "p"},
	}
	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(data), "clob_auth") {
		t.Fatalf("expected clob_auth: %s", string(data))
	}
}

func TestSubscriptionMarshalJSON_SliceFilters(t *testing.T) {
	sub := Subscription{
		Topic:   string(CryptoPrice),
		MsgType: "update",
		Filters: []string{"btc", "eth"},
	}
	data, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(data), `"filters":["btc","eth"]`) {
		t.Fatalf("expected array filters: %s", string(data))
	}
}

// --------------- signalDone ---------------

func TestSignalDone_Idempotent(t *testing.T) {
	c := newTestClient()
	c.signalDone()
	c.signalDone() // should not panic
}
