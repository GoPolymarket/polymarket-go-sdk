package ws

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
)

// --------------- normalizeWSURLs ---------------

func TestNormalizeWSURLs_Empty(t *testing.T) {
	market, user, base := normalizeWSURLs("")
	if base != ProdBaseURL {
		t.Fatalf("expected ProdBaseURL, got %s", base)
	}
	if market != ProdBaseURL+"/ws/market" {
		t.Fatalf("expected market URL, got %s", market)
	}
	if user != ProdBaseURL+"/ws/user" {
		t.Fatalf("expected user URL, got %s", user)
	}
}

func TestNormalizeWSURLs_MarketSuffix(t *testing.T) {
	market, user, base := normalizeWSURLs("wss://example.com/ws/market")
	if base != "wss://example.com" {
		t.Fatalf("expected base wss://example.com, got %s", base)
	}
	if market != "wss://example.com/ws/market" {
		t.Fatalf("expected market URL, got %s", market)
	}
	if user != "wss://example.com/ws/user" {
		t.Fatalf("expected user URL, got %s", user)
	}
}

func TestNormalizeWSURLs_UserSuffix(t *testing.T) {
	market, user, base := normalizeWSURLs("wss://example.com/ws/user")
	if base != "wss://example.com" {
		t.Fatalf("expected base, got %s", base)
	}
	if market != "wss://example.com/ws/market" {
		t.Fatalf("expected market URL, got %s", market)
	}
	if user != "wss://example.com/ws/user" {
		t.Fatalf("expected user URL, got %s", user)
	}
}

func TestNormalizeWSURLs_PlainURL(t *testing.T) {
	market, user, base := normalizeWSURLs("wss://example.com")
	if base != "wss://example.com" {
		t.Fatalf("expected base, got %s", base)
	}
	if market != "wss://example.com/ws/market" {
		t.Fatalf("got %s", market)
	}
	if user != "wss://example.com/ws/user" {
		t.Fatalf("got %s", user)
	}
}

func TestNormalizeWSURLs_TrailingSlash(t *testing.T) {
	market, _, base := normalizeWSURLs("wss://example.com/")
	if base != "wss://example.com" {
		t.Fatalf("expected trimmed base, got %s", base)
	}
	if !strings.HasSuffix(market, "/ws/market") {
		t.Fatalf("expected /ws/market suffix, got %s", market)
	}
}

// --------------- makeIDSet ---------------

func TestMakeIDSet_Nil(t *testing.T) {
	if makeIDSet(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestMakeIDSet_Empty(t *testing.T) {
	if makeIDSet([]string{}) != nil {
		t.Fatal("expected nil")
	}
}

func TestMakeIDSet_AllEmpty(t *testing.T) {
	if makeIDSet([]string{"", ""}) != nil {
		t.Fatal("expected nil for all-empty")
	}
}

func TestMakeIDSet_Normal(t *testing.T) {
	set := makeIDSet([]string{"a", "b", "a"})
	if len(set) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(set))
	}
	if _, ok := set["a"]; !ok {
		t.Fatal("missing a")
	}
	if _, ok := set["b"]; !ok {
		t.Fatal("missing b")
	}
}

func TestMakeIDSet_FiltersEmpty(t *testing.T) {
	set := makeIDSet([]string{"a", "", "b"})
	if len(set) != 2 {
		t.Fatalf("expected 2, got %d", len(set))
	}
}

// --------------- subscriptionEntry ---------------

func newTestSub(assets map[string]struct{}, markets map[string]struct{}) *subscriptionEntry[json.RawMessage] {
	return &subscriptionEntry[json.RawMessage]{
		id:      "test-1",
		channel: ChannelMarket,
		event:   Price,
		assets:  assets,
		markets: markets,
		ch:      make(chan json.RawMessage, 10),
		errCh:   make(chan error, 5),
	}
}

func TestSubscriptionEntry_MatchesAsset_EmptyMatchesAll(t *testing.T) {
	sub := newTestSub(nil, nil)
	if !sub.matchesAsset("anything") {
		t.Fatal("empty assets should match all")
	}
}

func TestSubscriptionEntry_MatchesAsset_Specific(t *testing.T) {
	sub := newTestSub(map[string]struct{}{"abc": {}}, nil)
	if !sub.matchesAsset("abc") {
		t.Fatal("should match abc")
	}
	if sub.matchesAsset("xyz") {
		t.Fatal("should not match xyz")
	}
}

func TestSubscriptionEntry_MatchesAnyAsset_EmptyMatchesAll(t *testing.T) {
	sub := newTestSub(nil, nil)
	if !sub.matchesAnyAsset([]string{"x"}) {
		t.Fatal("empty assets should match all")
	}
}

func TestSubscriptionEntry_MatchesAnyAsset_Partial(t *testing.T) {
	sub := newTestSub(map[string]struct{}{"a": {}, "b": {}}, nil)
	if !sub.matchesAnyAsset([]string{"c", "a"}) {
		t.Fatal("should match partial")
	}
}

func TestSubscriptionEntry_MatchesAnyAsset_NoMatch(t *testing.T) {
	sub := newTestSub(map[string]struct{}{"a": {}}, nil)
	if sub.matchesAnyAsset([]string{"x", "y"}) {
		t.Fatal("should not match")
	}
}

func TestSubscriptionEntry_MatchesMarket_EmptyMatchesAll(t *testing.T) {
	sub := newTestSub(nil, nil)
	if !sub.matchesMarket("any") {
		t.Fatal("empty markets should match all")
	}
}

func TestSubscriptionEntry_MatchesMarket_Specific(t *testing.T) {
	sub := newTestSub(nil, map[string]struct{}{"m1": {}})
	if !sub.matchesMarket("m1") {
		t.Fatal("should match m1")
	}
	if sub.matchesMarket("m2") {
		t.Fatal("should not match m2")
	}
}

func TestSubscriptionEntry_TrySend(t *testing.T) {
	sub := newTestSub(nil, nil)
	msg := json.RawMessage(`{"test":true}`)
	sub.trySend(msg)
	select {
	case got := <-sub.ch:
		if string(got) != `{"test":true}` {
			t.Fatalf("unexpected msg: %s", got)
		}
	default:
		t.Fatal("expected message in channel")
	}
}

func TestSubscriptionEntry_TrySend_Closed(t *testing.T) {
	sub := newTestSub(nil, nil)
	sub.close()
	// Should not panic
	sub.trySend(json.RawMessage(`{}`))
}

func TestSubscriptionEntry_Close_Idempotent(t *testing.T) {
	sub := newTestSub(nil, nil)
	first := sub.close()
	if !first {
		t.Fatal("first close should return true")
	}
	second := sub.close()
	if second {
		t.Fatal("second close should return false")
	}
}

func TestSubscriptionEntry_NotifyLag_ZeroCount(t *testing.T) {
	sub := newTestSub(nil, nil)
	sub.notifyLag(0)
	select {
	case <-sub.errCh:
		t.Fatal("should not send for count 0")
	default:
	}
}

func TestSubscriptionEntry_NotifyLag_Normal(t *testing.T) {
	sub := newTestSub(nil, nil)
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
		t.Fatal("expected error in channel")
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

func TestLaggedError_WithEventType(t *testing.T) {
	e := LaggedError{Count: 5, Channel: ChannelMarket, EventType: Price}
	s := e.Error()
	if !strings.Contains(s, "5") || !strings.Contains(s, "price") {
		t.Fatalf("unexpected error string: %s", s)
	}
}

func TestLaggedError_WithoutEventType(t *testing.T) {
	e := LaggedError{Count: 2}
	s := e.Error()
	if !strings.Contains(s, "2") {
		t.Fatalf("unexpected error string: %s", s)
	}
}

// --------------- Subscription builders ---------------

func TestNewMarketSubscription(t *testing.T) {
	req := NewMarketSubscription([]string{"a1", "a2"})
	if req.Type != ChannelMarket {
		t.Fatalf("expected market channel, got %s", req.Type)
	}
	if req.Operation != OperationSubscribe {
		t.Fatalf("expected subscribe, got %s", req.Operation)
	}
	if len(req.AssetIDs) != 2 {
		t.Fatalf("expected 2 asset IDs, got %d", len(req.AssetIDs))
	}
	if req.InitialDump == nil || !*req.InitialDump {
		t.Fatal("expected initial_dump true")
	}
}

func TestNewMarketUnsubscribe(t *testing.T) {
	req := NewMarketUnsubscribe([]string{"a1"})
	if req.Operation != OperationUnsubscribe {
		t.Fatalf("expected unsubscribe, got %s", req.Operation)
	}
	if req.InitialDump != nil {
		t.Fatal("unsubscribe should not have initial_dump")
	}
}

func TestNewUserSubscription(t *testing.T) {
	req := NewUserSubscription([]string{"m1"})
	if req.Type != ChannelSubscribe {
		t.Fatalf("expected subscribe channel, got %s", req.Type)
	}
	if req.Operation != OperationSubscribe {
		t.Fatalf("expected subscribe, got %s", req.Operation)
	}
	if len(req.Markets) != 1 {
		t.Fatalf("expected 1 market, got %d", len(req.Markets))
	}
}

func TestNewUserUnsubscribe(t *testing.T) {
	req := NewUserUnsubscribe([]string{"m1"})
	if req.Type != ChannelUser {
		t.Fatalf("expected user channel, got %s", req.Type)
	}
	if req.Operation != OperationUnsubscribe {
		t.Fatalf("expected unsubscribe, got %s", req.Operation)
	}
}

func TestWithCustomFeatures_Nil(t *testing.T) {
	var req *SubscriptionRequest
	got := req.WithCustomFeatures(true)
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestWithCustomFeatures_Normal(t *testing.T) {
	req := NewMarketSubscription([]string{"a1"})
	got := req.WithCustomFeatures(true)
	if got != req {
		t.Fatal("expected same pointer")
	}
	if req.CustomFeatureEnabled == nil || !*req.CustomFeatureEnabled {
		t.Fatal("expected custom features enabled")
	}
}

// --------------- snapshotSubs / closeSubMap ---------------

func TestSnapshotSubs(t *testing.T) {
	m := map[string]*subscriptionEntry[int]{
		"a": {id: "a", ch: make(chan int, 1), errCh: make(chan error, 1)},
		"b": {id: "b", ch: make(chan int, 1), errCh: make(chan error, 1)},
	}
	snap := snapshotSubs(m)
	if len(snap) != 2 {
		t.Fatalf("expected 2, got %d", len(snap))
	}
}

func TestCloseSubMap(t *testing.T) {
	m := map[string]*subscriptionEntry[int]{
		"a": {id: "a", ch: make(chan int, 1), errCh: make(chan error, 1)},
	}
	closeSubMap(m)
	if len(m) != 0 {
		t.Fatal("expected empty map")
	}
}

// --------------- newTestClient helper ---------------

func newTestClient() *clientImpl {
	return &clientImpl{
		done:               make(chan struct{}),
		marketRefs:         make(map[string]int),
		userRefs:           make(map[string]int),
		marketState:        ConnectionDisconnected,
		userState:          ConnectionDisconnected,
		orderbookSubs:      make(map[string]*subscriptionEntry[OrderbookEvent]),
		priceSubs:          make(map[string]*subscriptionEntry[PriceChangeEvent]),
		midpointSubs:       make(map[string]*subscriptionEntry[MidpointEvent]),
		lastTradeSubs:      make(map[string]*subscriptionEntry[LastTradePriceEvent]),
		tickSizeSubs:       make(map[string]*subscriptionEntry[TickSizeChangeEvent]),
		bestBidAskSubs:     make(map[string]*subscriptionEntry[BestBidAskEvent]),
		newMarketSubs:      make(map[string]*subscriptionEntry[NewMarketEvent]),
		marketResolvedSubs: make(map[string]*subscriptionEntry[MarketResolvedEvent]),
		tradeSubs:          make(map[string]*subscriptionEntry[TradeEvent]),
		orderSubs:          make(map[string]*subscriptionEntry[OrderEvent]),
		stateSubs:          make(map[string]*subscriptionEntry[ConnectionStateEvent]),
		orderbookCh:        make(chan OrderbookEvent, 100),
		priceCh:            make(chan PriceEvent, 100),
		midpointCh:         make(chan MidpointEvent, 100),
		lastTradeCh:        make(chan LastTradePriceEvent, 100),
		tickSizeCh:         make(chan TickSizeChangeEvent, 100),
		bestBidAskCh:       make(chan BestBidAskEvent, 100),
		newMarketCh:        make(chan NewMarketEvent, 100),
		marketResolvedCh:   make(chan MarketResolvedEvent, 100),
		tradeCh:            make(chan TradeEvent, 100),
		orderCh:            make(chan OrderEvent, 100),
	}
}

// --------------- processEvent ---------------

func TestProcessEvent_Price(t *testing.T) {
	c := newTestClient()
	ch := make(chan PriceChangeEvent, 5)
	c.priceSubs["p1"] = &subscriptionEntry[PriceChangeEvent]{
		id: "p1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{
		"event_type": "price",
		"market":     "m1",
		"price_changes": []interface{}{
			map[string]interface{}{"asset_id": "tok1", "price": "0.55"},
		},
	}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for price event")
	}
}

func TestProcessEvent_PriceChange(t *testing.T) {
	c := newTestClient()
	ch := make(chan PriceChangeEvent, 5)
	c.priceSubs["p1"] = &subscriptionEntry[PriceChangeEvent]{
		id: "p1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{
		"event_type": "price_change",
		"market":     "m1",
		"price_changes": []interface{}{
			map[string]interface{}{"asset_id": "tok2", "price": "0.60"},
		},
	}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok2" {
			t.Fatalf("expected tok2, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_Book(t *testing.T) {
	c := newTestClient()
	ch := make(chan OrderbookEvent, 5)
	c.orderbookSubs["ob1"] = &subscriptionEntry[OrderbookEvent]{
		id: "ob1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{
		"event_type": "book",
		"asset_id":   "tok1",
		"market":     "m1",
		"bids":       []interface{}{map[string]interface{}{"price": "0.5", "size": "10"}},
		"asks":       []interface{}{map[string]interface{}{"price": "0.6", "size": "10"}},
		"timestamp":  "1700000000",
	}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_BookGeneratesMidpoint(t *testing.T) {
	c := newTestClient()
	midCh := make(chan MidpointEvent, 5)
	c.midpointSubs["mid1"] = &subscriptionEntry[MidpointEvent]{
		id: "mid1", ch: midCh, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{
		"event_type": "book",
		"asset_id":   "tok1",
		"bids":       []interface{}{map[string]interface{}{"price": "0.4", "size": "10"}},
		"asks":       []interface{}{map[string]interface{}{"price": "0.6", "size": "10"}},
	}
	c.processEvent(raw)

	select {
	case ev := <-midCh:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
		if ev.Midpoint != "0.5" {
			t.Fatalf("expected midpoint 0.5, got %s", ev.Midpoint)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for midpoint")
	}
}

func TestProcessEvent_LastTradePrice(t *testing.T) {
	c := newTestClient()
	ch := make(chan LastTradePriceEvent, 5)
	c.lastTradeSubs["ltp1"] = &subscriptionEntry[LastTradePriceEvent]{
		id: "ltp1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{"event_type": "last_trade_price", "asset_id": "tok1", "price": "0.55"}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_TickSizeChange(t *testing.T) {
	c := newTestClient()
	ch := make(chan TickSizeChangeEvent, 5)
	c.tickSizeSubs["ts1"] = &subscriptionEntry[TickSizeChangeEvent]{
		id: "ts1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{"event_type": "tick_size_change", "asset_id": "tok1", "tick_size": "0.01"}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_BestBidAsk(t *testing.T) {
	c := newTestClient()
	ch := make(chan BestBidAskEvent, 5)
	c.bestBidAskSubs["bba1"] = &subscriptionEntry[BestBidAskEvent]{
		id: "bba1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{"event_type": "best_bid_ask", "asset_id": "tok1", "best_bid": "0.5", "best_ask": "0.6"}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_Trade(t *testing.T) {
	c := newTestClient()
	ch := make(chan TradeEvent, 5)
	c.tradeSubs["tr1"] = &subscriptionEntry[TradeEvent]{
		id: "tr1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{"event_type": "trade", "asset_id": "tok1", "side": "BUY", "size": "10", "price": "0.5"}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_Order(t *testing.T) {
	c := newTestClient()
	ch := make(chan OrderEvent, 5)
	c.orderSubs["or1"] = &subscriptionEntry[OrderEvent]{
		id: "or1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{"event_type": "order", "asset_id": "tok1", "side": "SELL", "size": "5"}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.AssetID != "tok1" {
			t.Fatalf("expected tok1, got %s", ev.AssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_NewMarket(t *testing.T) {
	c := newTestClient()
	ch := make(chan NewMarketEvent, 5)
	c.newMarketSubs["mk1"] = &subscriptionEntry[NewMarketEvent]{
		id: "mk1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{"event_type": "new_market", "market": "m1", "assets_ids": []interface{}{"a1", "a2"}}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.Market != "m1" {
			t.Fatalf("expected m1, got %s", ev.Market)
		}
		if len(ev.AssetIDs) != 2 {
			t.Fatalf("expected 2 asset IDs, got %d", len(ev.AssetIDs))
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_NewMarket_AltAssetIDs(t *testing.T) {
	c := newTestClient()
	ch := make(chan NewMarketEvent, 5)
	c.newMarketSubs["mk1"] = &subscriptionEntry[NewMarketEvent]{
		id: "mk1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{"event_type": "new_market", "market": "m1", "asset_ids": []interface{}{"a1"}}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if len(ev.AssetIDs) != 1 || ev.AssetIDs[0] != "a1" {
			t.Fatalf("expected fallback asset_ids, got %v", ev.AssetIDs)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_MarketResolved(t *testing.T) {
	c := newTestClient()
	ch := make(chan MarketResolvedEvent, 5)
	c.marketResolvedSubs["mr1"] = &subscriptionEntry[MarketResolvedEvent]{
		id: "mr1", ch: ch, errCh: make(chan error, 5),
	}

	raw := map[string]interface{}{
		"event_type":       "market_resolved",
		"market":           "m1",
		"assets_ids":       []interface{}{"a1"},
		"winning_asset_id": "a1",
		"winning_outcome":  "Yes",
	}
	c.processEvent(raw)

	select {
	case ev := <-ch:
		if ev.WinningAssetID != "a1" {
			t.Fatalf("expected a1, got %s", ev.WinningAssetID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestProcessEvent_Unknown(t *testing.T) {
	c := newTestClient()
	// Should not panic on unknown event type
	raw := map[string]interface{}{"event_type": "unknown_type", "data": "test"}
	c.processEvent(raw)
}

// --------------- ConnectionState ---------------

func TestConnectionState_Market(t *testing.T) {
	c := newTestClient()
	c.stateMu.Lock()
	c.marketState = ConnectionConnected
	c.stateMu.Unlock()
	if c.ConnectionState(ChannelMarket) != ConnectionConnected {
		t.Fatal("expected connected")
	}
}

func TestConnectionState_User(t *testing.T) {
	c := newTestClient()
	c.stateMu.Lock()
	c.userState = ConnectionConnected
	c.stateMu.Unlock()
	if c.ConnectionState(ChannelUser) != ConnectionConnected {
		t.Fatal("expected connected")
	}
}

func TestConnectionState_Unknown(t *testing.T) {
	c := newTestClient()
	if c.ConnectionState("unknown") != ConnectionDisconnected {
		t.Fatal("expected disconnected for unknown channel")
	}
}

func TestConnectionState_EmptyDefault(t *testing.T) {
	c := newTestClient()
	c.stateMu.Lock()
	c.marketState = ""
	c.stateMu.Unlock()
	if c.ConnectionState(ChannelMarket) != ConnectionDisconnected {
		t.Fatal("expected disconnected for empty state")
	}
}

// --------------- trySendGlobal ---------------

func TestTrySendGlobal_NilChannel(t *testing.T) {
	trySendGlobal[int](nil, 42)
}

func TestTrySendGlobal_Normal(t *testing.T) {
	ch := make(chan int, 1)
	trySendGlobal(ch, 42)
	select {
	case v := <-ch:
		if v != 42 {
			t.Fatalf("expected 42, got %d", v)
		}
	default:
		t.Fatal("expected value in channel")
	}
}

func TestTrySendGlobal_FullChannel(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	trySendGlobal(ch, 2) // should not block
}

// --------------- authPayload ---------------

func TestAuthPayload_NilApiKey(t *testing.T) {
	c := newTestClient()
	if c.authPayload() != nil {
		t.Fatal("expected nil for nil apiKey")
	}
}

func TestAuthPayload_EmptyFields(t *testing.T) {
	c := newTestClient()
	c.apiKey = &auth.APIKey{}
	if c.authPayload() != nil {
		t.Fatal("expected nil for empty apiKey")
	}
}

func TestAuthPayload_Valid(t *testing.T) {
	c := newTestClient()
	c.apiKey = &auth.APIKey{Key: "k", Secret: "s", Passphrase: "p"}
	p := c.authPayload()
	if p == nil {
		t.Fatal("expected non-nil payload")
	}
	if p.APIKey != "k" || p.Secret != "s" || p.Passphrase != "p" {
		t.Fatalf("unexpected payload: %+v", p)
	}
}

// --------------- resolveAuth ---------------

func TestResolveAuth_Explicit(t *testing.T) {
	c := newTestClient()
	explicit := &AuthPayload{APIKey: "explicit", Secret: "s", Passphrase: "p"}
	got := c.resolveAuth(explicit)
	if got == nil || got.APIKey != "explicit" {
		t.Fatal("expected explicit key")
	}
	// Should be a copy
	if got == explicit {
		t.Fatal("expected copy, not same pointer")
	}
}

func TestResolveAuth_FromApiKey(t *testing.T) {
	c := newTestClient()
	c.apiKey = &auth.APIKey{Key: "api", Secret: "s", Passphrase: "p"}
	got := c.resolveAuth(nil)
	if got == nil || got.APIKey != "api" {
		t.Fatal("expected apiKey")
	}
}

func TestResolveAuth_FromLastAuth(t *testing.T) {
	c := newTestClient()
	c.lastAuth = &AuthPayload{APIKey: "last", Secret: "s", Passphrase: "p"}
	got := c.resolveAuth(nil)
	if got == nil || got.APIKey != "last" {
		t.Fatal("expected lastAuth")
	}
}

func TestResolveAuth_AllNil(t *testing.T) {
	c := newTestClient()
	got := c.resolveAuth(nil)
	if got != nil {
		t.Fatal("expected nil")
	}
}

// --------------- Authenticate / Deauthenticate ---------------

func TestAuthenticate(t *testing.T) {
	c := newTestClient()
	key := &auth.APIKey{Key: "k", Secret: "s", Passphrase: "p"}
	ret := c.Authenticate(nil, key)
	if ret != c {
		t.Fatal("expected same client returned")
	}
	if c.apiKey == nil || c.apiKey.Key != "k" {
		t.Fatal("apiKey not set")
	}
}

func TestDeauthenticate(t *testing.T) {
	c := newTestClient()
	c.apiKey = &auth.APIKey{Key: "k", Secret: "s", Passphrase: "p"}
	ret := c.Deauthenticate()
	if ret != c {
		t.Fatal("expected same client returned")
	}
	if c.apiKey != nil {
		t.Fatal("apiKey should be nil")
	}
}

// --------------- addMarketRefs / removeMarketRefs ---------------

func TestAddMarketRefs(t *testing.T) {
	c := newTestClient()
	newAssets := c.addMarketRefs([]string{"a1", "a2"}, false)
	if len(newAssets) != 2 {
		t.Fatalf("expected 2 new, got %d", len(newAssets))
	}
	// Adding same again should return empty
	newAssets = c.addMarketRefs([]string{"a1"}, false)
	if len(newAssets) != 0 {
		t.Fatalf("expected 0 new, got %d", len(newAssets))
	}
	if c.marketRefs["a1"] != 2 {
		t.Fatalf("expected ref count 2, got %d", c.marketRefs["a1"])
	}
}

func TestAddMarketRefs_FiltersEmpty(t *testing.T) {
	c := newTestClient()
	newAssets := c.addMarketRefs([]string{"", "a1", ""}, false)
	if len(newAssets) != 1 || newAssets[0] != "a1" {
		t.Fatalf("expected [a1], got %v", newAssets)
	}
}

func TestRemoveMarketRefs(t *testing.T) {
	c := newTestClient()
	c.addMarketRefs([]string{"a1", "a2"}, false)
	c.addMarketRefs([]string{"a1"}, false) // a1 has ref count 2

	toUnsub := c.removeMarketRefs([]string{"a1"})
	if len(toUnsub) != 0 {
		t.Fatalf("expected 0 unsub (still has ref), got %v", toUnsub)
	}
	toUnsub = c.removeMarketRefs([]string{"a1", "a2"})
	if len(toUnsub) != 2 {
		t.Fatalf("expected 2 unsub, got %v", toUnsub)
	}
}

// --------------- addUserRefs / removeUserRefs ---------------

func TestAddUserRefs(t *testing.T) {
	c := newTestClient()
	ap := &AuthPayload{APIKey: "k", Secret: "s", Passphrase: "p"}
	newMarkets := c.addUserRefs([]string{"m1"}, ap)
	if len(newMarkets) != 1 {
		t.Fatalf("expected 1 new, got %d", len(newMarkets))
	}
	if c.lastAuth == nil || c.lastAuth.APIKey != "k" {
		t.Fatal("lastAuth not set")
	}
}

func TestRemoveUserRefs(t *testing.T) {
	c := newTestClient()
	ap := &AuthPayload{APIKey: "k", Secret: "s", Passphrase: "p"}
	c.addUserRefs([]string{"m1"}, ap)
	toUnsub := c.removeUserRefs([]string{"m1"})
	if len(toUnsub) != 1 {
		t.Fatalf("expected 1 unsub, got %v", toUnsub)
	}
}

// --------------- applySubscription validation ---------------

func TestApplySubscription_NilRequest(t *testing.T) {
	c := newTestClient()
	err := c.applySubscription(nil, OperationSubscribe)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestApplySubscription_NoTypeNoIDs(t *testing.T) {
	c := newTestClient()
	err := c.applySubscription(&SubscriptionRequest{}, OperationSubscribe)
	if err == nil || !strings.Contains(err.Error(), "type is required") {
		t.Fatalf("expected type required error, got %v", err)
	}
}

func TestApplySubscription_InferMarketType(t *testing.T) {
	c := newTestClient()
	req := &SubscriptionRequest{
		Operation: OperationSubscribe,
		AssetIDs:  []string{"a1"},
	}
	// Will fail at ensureConn (no real WS), but should pass validation
	err := c.applySubscription(req, OperationSubscribe)
	if err != nil && strings.Contains(err.Error(), "type is required") {
		t.Fatalf("should have inferred market type: %v", err)
	}
}

func TestApplySubscription_InferUserType(t *testing.T) {
	c := newTestClient()
	c.apiKey = &auth.APIKey{Key: "k", Secret: "s", Passphrase: "p"}
	req := &SubscriptionRequest{
		Operation: OperationSubscribe,
		Markets:   []string{"m1"},
	}
	err := c.applySubscription(req, OperationSubscribe)
	if err != nil && strings.Contains(err.Error(), "type is required") {
		t.Fatalf("should have inferred user type: %v", err)
	}
}

func TestApplySubscription_MarketMissingAssets(t *testing.T) {
	c := newTestClient()
	req := &SubscriptionRequest{Type: ChannelMarket}
	err := c.applySubscription(req, OperationSubscribe)
	if err == nil || !strings.Contains(err.Error(), "assetIDs required") {
		t.Fatalf("expected assetIDs required, got %v", err)
	}
}

func TestApplySubscription_UserMissingMarkets(t *testing.T) {
	c := newTestClient()
	req := &SubscriptionRequest{Type: ChannelUser}
	err := c.applySubscription(req, OperationSubscribe)
	if err == nil || !strings.Contains(err.Error(), "markets required") {
		t.Fatalf("expected markets required, got %v", err)
	}
}

func TestApplySubscription_UnknownChannel(t *testing.T) {
	c := newTestClient()
	req := &SubscriptionRequest{Type: "unknown"}
	err := c.applySubscription(req, OperationSubscribe)
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("expected unknown channel error, got %v", err)
	}
}

// --------------- setConnState / ConnectionStateStream ---------------

func TestSetConnState(t *testing.T) {
	c := newTestClient()
	ch := make(chan ConnectionStateEvent, 10)
	c.stateSubs["s1"] = &subscriptionEntry[ConnectionStateEvent]{
		id: "s1", ch: ch, errCh: make(chan error, 5),
	}

	c.setConnState(ChannelMarket, ConnectionConnected, 0)

	if c.ConnectionState(ChannelMarket) != ConnectionConnected {
		t.Fatal("expected connected")
	}

	select {
	case ev := <-ch:
		if ev.State != ConnectionConnected || ev.Channel != ChannelMarket {
			t.Fatalf("unexpected event: %+v", ev)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for state event")
	}
}

// --------------- Concurrent safety ---------------

func TestConcurrentTrySend(t *testing.T) {
	sub := newTestSub(nil, nil)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			sub.trySend(json.RawMessage(`{"n":1}`))
		}(i)
	}
	wg.Wait()
	sub.close()
}

func TestConcurrentClose(t *testing.T) {
	sub := newTestSub(nil, nil)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub.close()
		}()
	}
	wg.Wait()
}

// --------------- Deadlock regression test ---------------

func TestTrySend_FullChannel_NoDeadlock(t *testing.T) {
	// Regression test for reentrant RLock deadlock in trySend -> notifyLag.
	// trySend holds RLock and calls notifyLagLocked (no lock) instead of
	// notifyLag (which would re-acquire RLock and potentially deadlock).
	sub := &subscriptionEntry[int]{
		id:      "deadlock-test",
		channel: ChannelMarket,
		event:   Price,
		ch:      make(chan int, 1), // small buffer to trigger full path
		errCh:   make(chan error, 5),
	}

	// Fill the channel
	sub.ch <- 42

	// This should complete without deadlock
	done := make(chan struct{})
	go func() {
		defer close(done)
		sub.trySend(99) // channel full → notifyLagLocked path
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(2 * time.Second):
		t.Fatal("trySend deadlocked on full channel")
	}

	// Verify lag error was sent
	select {
	case err := <-sub.errCh:
		if _, ok := err.(LaggedError); !ok {
			t.Fatalf("expected LaggedError, got %T", err)
		}
	default:
		t.Fatal("expected lag notification")
	}
}

func TestTrySend_ConcurrentCloseAndFullChannel(t *testing.T) {
	// Test concurrent close while trySend hits full channel path.
	// This would deadlock with the old reentrant RLock implementation.
	sub := &subscriptionEntry[int]{
		id:      "concurrent-test",
		channel: ChannelMarket,
		event:   Price,
		ch:      make(chan int, 1),
		errCh:   make(chan error, 5),
	}

	sub.ch <- 1 // fill

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub.trySend(99)
		}()
	}
	// Close concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		sub.close()
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected in concurrent trySend + close")
	}
}

// --------------- PriceChangeEvent.AssetID naming ---------------

func TestPriceChangeEvent_AssetID_JSONRoundtrip(t *testing.T) {
	original := PriceChangeEvent{
		AssetID: "0x123abc",
		Price:   "0.55",
		BestBid: "0.54",
		BestAsk: "0.56",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify JSON uses snake_case
	if !strings.Contains(string(data), `"asset_id"`) {
		t.Fatalf("expected asset_id in JSON, got %s", data)
	}

	var decoded PriceChangeEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.AssetID != original.AssetID {
		t.Fatalf("expected %s, got %s", original.AssetID, decoded.AssetID)
	}
}
