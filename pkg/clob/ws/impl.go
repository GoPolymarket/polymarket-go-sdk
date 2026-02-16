package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/logger"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

const (
	ProdBaseURL        = "wss://ws-subscriptions-clob.polymarket.com"
	DefaultReadTimeout = 60 * time.Second
)

type clientImpl struct {
	baseURL      string
	marketURL    string
	userURL      string
	conn         *websocket.Conn
	userConn     *websocket.Conn
	signer       auth.Signer
	apiKey       *auth.APIKey
	mu           sync.Mutex
	userMu       sync.Mutex
	marketInitMu sync.Mutex
	userInitMu   sync.Mutex
	done         chan struct{}
	closeOnce    sync.Once
	closing      atomic.Bool
	// Per-connection context cancellation for goroutine lifecycle management
	marketCtx      context.Context
	marketCancel   context.CancelFunc
	userCtx        context.Context
	userCancel     context.CancelFunc
	goroutineCtxMu sync.Mutex
	// Subscription state
	debug               bool
	disablePing         bool
	reconnect           bool
	reconnectMax        int
	reconnectDelay      time.Duration
	reconnectMaxDelay   time.Duration
	reconnectMultiplier float64
	heartbeatInterval   time.Duration
	heartbeatTimeout    time.Duration
	readTimeout         atomic.Int64 // stored as nanoseconds

	lastPongMarket atomic.Int64
	lastPongUser   atomic.Int64

	subMu          sync.Mutex
	marketRefs     map[string]int
	userRefs       map[string]int
	lastAuth       *AuthPayload
	customFeatures bool
	nextSubID      uint64

	// Connection state
	stateMu     sync.Mutex
	marketState ConnectionState
	userState   ConnectionState

	// Stream subscriptions
	orderbookSubs      map[string]*subscriptionEntry[OrderbookEvent]
	priceSubs          map[string]*subscriptionEntry[PriceChangeEvent]
	midpointSubs       map[string]*subscriptionEntry[MidpointEvent]
	lastTradeSubs      map[string]*subscriptionEntry[LastTradePriceEvent]
	tickSizeSubs       map[string]*subscriptionEntry[TickSizeChangeEvent]
	bestBidAskSubs     map[string]*subscriptionEntry[BestBidAskEvent]
	newMarketSubs      map[string]*subscriptionEntry[NewMarketEvent]
	marketResolvedSubs map[string]*subscriptionEntry[MarketResolvedEvent]
	tradeSubs          map[string]*subscriptionEntry[TradeEvent]
	orderSubs          map[string]*subscriptionEntry[OrderEvent]
	stateSubs          map[string]*subscriptionEntry[ConnectionStateEvent]

	// Channels
	orderbookCh      chan OrderbookEvent
	priceCh          chan PriceEvent
	midpointCh       chan MidpointEvent
	lastTradeCh      chan LastTradePriceEvent
	tickSizeCh       chan TickSizeChangeEvent
	bestBidAskCh     chan BestBidAskEvent
	newMarketCh      chan NewMarketEvent
	marketResolvedCh chan MarketResolvedEvent
	tradeCh          chan TradeEvent
	orderCh          chan OrderEvent

	// Callbacks or listeners could be added here
}

func NewClient(url string, signer auth.Signer, apiKey *auth.APIKey) (Client, error) {
	marketURL, userURL, baseURL := normalizeWSURLs(url)

	reconnect := true
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT")); raw != "" {
		reconnect = raw != "0" && strings.ToLower(raw) != "false"
	}
	reconnectDelay := 2 * time.Second
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_DELAY_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			reconnectDelay = time.Duration(ms) * time.Millisecond
		}
	}
	reconnectMaxDelay := 30 * time.Second
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_MAX_DELAY_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			reconnectMaxDelay = time.Duration(ms) * time.Millisecond
		}
	}
	reconnectMultiplier := 2.0
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_BACKOFF_MULTIPLIER")); raw != "" {
		if mult, err := strconv.ParseFloat(raw, 64); err == nil && mult > 0 {
			reconnectMultiplier = mult
		}
	}
	reconnectMax := 5
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_MAX")); raw != "" {
		if max, err := strconv.Atoi(raw); err == nil {
			reconnectMax = max
		}
	}
	heartbeatInterval := 10 * time.Second
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_HEARTBEAT_INTERVAL_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			heartbeatInterval = time.Duration(ms) * time.Millisecond
		}
	}
	heartbeatTimeout := 30 * time.Second
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_HEARTBEAT_TIMEOUT_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			heartbeatTimeout = time.Duration(ms) * time.Millisecond
		}
	} else if heartbeatInterval > 0 {
		heartbeatTimeout = heartbeatInterval * 3
	}

	c := &clientImpl{
		baseURL:             baseURL,
		marketURL:           marketURL,
		userURL:             userURL,
		signer:              signer,
		apiKey:              apiKey,
		debug:               os.Getenv("CLOB_WS_DEBUG") != "",
		disablePing:         os.Getenv("CLOB_WS_DISABLE_PING") != "",
		reconnect:           reconnect,
		reconnectDelay:      reconnectDelay,
		reconnectMaxDelay:   reconnectMaxDelay,
		reconnectMultiplier: reconnectMultiplier,
		reconnectMax:        reconnectMax,
		heartbeatInterval:   heartbeatInterval,
		heartbeatTimeout:    heartbeatTimeout,
		done:                make(chan struct{}),
		marketRefs:          make(map[string]int),
		userRefs:            make(map[string]int),
		marketState:         ConnectionDisconnected,
		userState:           ConnectionDisconnected,
		orderbookSubs:       make(map[string]*subscriptionEntry[OrderbookEvent]),
		priceSubs:           make(map[string]*subscriptionEntry[PriceChangeEvent]),
		midpointSubs:        make(map[string]*subscriptionEntry[MidpointEvent]),
		lastTradeSubs:       make(map[string]*subscriptionEntry[LastTradePriceEvent]),
		tickSizeSubs:        make(map[string]*subscriptionEntry[TickSizeChangeEvent]),
		bestBidAskSubs:      make(map[string]*subscriptionEntry[BestBidAskEvent]),
		newMarketSubs:       make(map[string]*subscriptionEntry[NewMarketEvent]),
		marketResolvedSubs:  make(map[string]*subscriptionEntry[MarketResolvedEvent]),
		tradeSubs:           make(map[string]*subscriptionEntry[TradeEvent]),
		orderSubs:           make(map[string]*subscriptionEntry[OrderEvent]),
		stateSubs:           make(map[string]*subscriptionEntry[ConnectionStateEvent]),
		orderbookCh:         make(chan OrderbookEvent, 100),
		priceCh:             make(chan PriceEvent, 100),
		midpointCh:          make(chan MidpointEvent, 100),
		lastTradeCh:         make(chan LastTradePriceEvent, 100),
		tickSizeCh:          make(chan TickSizeChangeEvent, 100),
		bestBidAskCh:        make(chan BestBidAskEvent, 100),
		newMarketCh:         make(chan NewMarketEvent, 100),
		marketResolvedCh:    make(chan MarketResolvedEvent, 100),
		tradeCh:             make(chan TradeEvent, 100),
		orderCh:             make(chan OrderEvent, 100),
	}

	// Initialize atomic readTimeout
	c.readTimeout.Store(int64(DefaultReadTimeout))

	if err := c.ensureMarketConn(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *clientImpl) Authenticate(signer auth.Signer, apiKey *auth.APIKey) Client {
	c.signer = signer
	c.apiKey = apiKey
	c.subMu.Lock()
	c.lastAuth = nil
	c.subMu.Unlock()
	return c
}

func (c *clientImpl) Deauthenticate() Client {
	c.signer = nil
	c.apiKey = nil
	c.subMu.Lock()
	c.lastAuth = nil
	c.subMu.Unlock()
	c.closeConn(ChannelUser)
	return c
}

func normalizeWSURLs(raw string) (string, string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = ProdBaseURL
	}
	trimmed := strings.TrimRight(raw, "/")
	switch {
	case strings.HasSuffix(trimmed, "/ws/market"):
		base := strings.TrimSuffix(trimmed, "/ws/market")
		return trimmed, base + "/ws/user", base
	case strings.HasSuffix(trimmed, "/ws/user"):
		base := strings.TrimSuffix(trimmed, "/ws/user")
		return base + "/ws/market", trimmed, base
	default:
		return trimmed + "/ws/market", trimmed + "/ws/user", trimmed
	}
}

func (c *clientImpl) pingLoop(channel Channel) {
	interval := c.heartbeatInterval
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Get the context for this connection to enable proper cancellation
	ctx := c.getGoroutineContext(channel)
	if ctx == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			// Context cancelled - connection is being replaced or closed
			return
		case <-c.done:
			return
		case <-ticker.C:
			if timeout := c.heartbeatTimeout; timeout > 0 {
				last := c.lastPong(channel)
				if !last.IsZero() && time.Since(last) > timeout {
					if c.debug {
						logger.Warn("heartbeat timeout on %s (last pong %s)", channel, last.Format(time.RFC3339))
					}
					c.closeConn(channel)
					return
				}
			}
			// CLOB WS uses "PING" string for Keep-Alive
			err := c.writeMessage(channel, []byte("PING"))
			if err != nil {
				return
			}
		}
	}
}

func (c *clientImpl) ensureMarketConn() error {
	c.marketInitMu.Lock()
	defer c.marketInitMu.Unlock()
	if c.getConn(ChannelMarket) != nil {
		return nil
	}
	c.setConnState(ChannelMarket, ConnectionConnecting, 0)

	// Cancel any existing goroutines for this connection
	c.cancelGoroutines(ChannelMarket)

	// Create new context for this connection's goroutines
	c.createGoroutineContext(ChannelMarket)

	if err := c.connectMarket(); err != nil {
		c.setConnState(ChannelMarket, ConnectionDisconnected, 0)
		return err
	}
	c.setConnState(ChannelMarket, ConnectionConnected, 0)
	c.setLastPong(ChannelMarket, time.Now())
	go c.readLoop(ChannelMarket)
	if !c.disablePing {
		go c.pingLoop(ChannelMarket)
	}
	return nil
}

func (c *clientImpl) ensureUserConn() error {
	c.userInitMu.Lock()
	defer c.userInitMu.Unlock()
	if c.getConn(ChannelUser) != nil {
		return nil
	}
	c.setConnState(ChannelUser, ConnectionConnecting, 0)

	// Cancel any existing goroutines for this connection
	c.cancelGoroutines(ChannelUser)

	// Create new context for this connection's goroutines
	c.createGoroutineContext(ChannelUser)

	if err := c.connectUser(); err != nil {
		c.setConnState(ChannelUser, ConnectionDisconnected, 0)
		return err
	}
	c.setConnState(ChannelUser, ConnectionConnected, 0)
	c.setLastPong(ChannelUser, time.Now())
	go c.readLoop(ChannelUser)
	if !c.disablePing {
		go c.pingLoop(ChannelUser)
	}
	return nil
}

func (c *clientImpl) ensureConn(channel Channel) error {
	switch channel {
	case ChannelMarket:
		return c.ensureMarketConn()
	case ChannelUser:
		return c.ensureUserConn()
	default:
		return errors.New("unknown subscription channel")
	}
}

func (c *clientImpl) connect(url string, setConn func(*websocket.Conn)) error {
	headers := http.Header{}
	headers.Set("User-Agent", "Go-Polymarket-SDK/1.0")

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		return err
	}
	setConn(conn)

	// If authenticated, send auth message or headers?
	// Polymarket WS usually requires auth for private channels (orders, trades).
	// Public channels don't need auth.
	// For now, we assume public channels work without auth.
	// If we need auth, we might need to send a specialized message after connect.

	return nil
}

func (c *clientImpl) connectMarket() error {
	return c.connect(c.marketURL, c.setMarketConn)
}

func (c *clientImpl) connectUser() error {
	return c.connect(c.userURL, c.setUserConn)
}

func (c *clientImpl) readLoop(channel Channel) {
	// Get the context for this connection to enable proper cancellation
	ctx := c.getGoroutineContext(channel)
	if ctx == nil {
		return
	}

	// Set initial read deadline
	if conn := c.getConn(channel); conn != nil {
		timeout := time.Duration(c.readTimeout.Load())
		_ = conn.SetReadDeadline(time.Now().Add(timeout))
	}

	for {
		// Check if context is cancelled before reading
		select {
		case <-ctx.Done():
			// Context cancelled - connection is being replaced or closed
			return
		default:
		}

		conn := c.getConn(channel)
		if conn == nil {
			if c.closing.Load() {
				break
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			if c.closing.Load() {
				break
			}
			if c.reconnect {
				if c.debug {
					logger.Debug("read error: %v (reconnecting)", err)
				}
				if err := c.reconnectLoop(channel); err == nil {
					// Reconnection successful - a new readLoop has been started
					// Exit this readLoop to avoid multiple goroutines reading from the same connection
					return
				}
			}
			logger.Error("read error: %v", err)
			c.setConnState(channel, ConnectionDisconnected, 0)
			break
		}

		c.setLastPong(channel, time.Now())

		// Refresh read deadline
		timeout := time.Duration(c.readTimeout.Load())
		_ = conn.SetReadDeadline(time.Now().Add(timeout))

		// Check for PONG
		if string(message) == "PONG" {
			if c.debug {
				logger.Debug("Received PONG")
			}
			continue
		}

		// Debug: Print raw message to troubleshoot "no events"
		if c.debug {
			logger.Debug("Raw WS Message: %s", string(message))
		}

		// Parse generic message to determine type
		var rawObj map[string]interface{}
		var rawArr []map[string]interface{}

		// Try unmarshal as array first
		if err := json.Unmarshal(message, &rawArr); err == nil {
			for _, item := range rawArr {
				c.processEvent(item)
			}
			continue
		}

		// Try unmarshal as single object
		if err := json.Unmarshal(message, &rawObj); err == nil {
			c.processEvent(rawObj)
			continue
		}
	}
	if c.closing.Load() {
		c.shutdown()
	}
}

func (c *clientImpl) setLastPong(channel Channel, t time.Time) {
	switch channel {
	case ChannelMarket:
		c.lastPongMarket.Store(t.UnixNano())
	case ChannelUser:
		c.lastPongUser.Store(t.UnixNano())
	}
}

func (c *clientImpl) lastPong(channel Channel) time.Time {
	switch channel {
	case ChannelMarket:
		if nanos := c.lastPongMarket.Load(); nanos > 0 {
			return time.Unix(0, nanos)
		}
	case ChannelUser:
		if nanos := c.lastPongUser.Load(); nanos > 0 {
			return time.Unix(0, nanos)
		}
	}
	return time.Time{}
}

func (c *clientImpl) processEvent(raw map[string]interface{}) {
	eventType, _ := raw["event_type"].(string)
	if eventType == "" {
		eventType, _ = raw["type"].(string)
	}

	// Re-marshal to bytes to use existing logic or decode from map directly
	// For simplicity, let's just use the map or re-marshal for struct decoding
	// Re-marshalling is inefficient but safe for now to reuse struct definitions
	msgBytes, _ := json.Marshal(raw)

	switch eventType {
	case "book", "orderbook": // Orderbook snapshot/update
		var wire struct {
			AssetID   string           `json:"asset_id"`
			Market    string           `json:"market"`
			Bids      []OrderbookLevel `json:"bids"`
			Asks      []OrderbookLevel `json:"asks"`
			Buys      []OrderbookLevel `json:"buys"`
			Sells     []OrderbookLevel `json:"sells"`
			Hash      string           `json:"hash"`
			Timestamp string           `json:"timestamp"`
		}
		if err := json.Unmarshal(msgBytes, &wire); err == nil {
			event := OrderbookEvent{
				AssetID:   wire.AssetID,
				Market:    wire.Market,
				Bids:      wire.Bids,
				Asks:      wire.Asks,
				Hash:      wire.Hash,
				Timestamp: wire.Timestamp,
			}
			if len(event.Bids) == 0 && len(wire.Buys) > 0 {
				event.Bids = wire.Buys
			}
			if len(event.Asks) == 0 && len(wire.Sells) > 0 {
				event.Asks = wire.Sells
			}
			c.dispatchOrderbook(event)

			if len(event.Bids) > 0 && len(event.Asks) > 0 {
				bid, bidErr := decimal.NewFromString(event.Bids[0].Price)
				ask, askErr := decimal.NewFromString(event.Asks[0].Price)
				if bidErr == nil && askErr == nil {
					mid := bid.Add(ask).Div(decimal.NewFromInt(2))
					c.dispatchMidpoint(MidpointEvent{AssetID: event.AssetID, Midpoint: mid.String()})
				}
			}
		}
	case "price", "price_change":
		var event PriceEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			c.dispatchPrice(event)
		}
	case "midpoint":
		var event MidpointEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			c.dispatchMidpoint(event)
		}
	case "last_trade_price":
		var event LastTradePriceEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			c.dispatchLastTrade(event)
		}
	case "tick_size_change":
		var event TickSizeChangeEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			c.dispatchTickSize(event)
		}
	case "best_bid_ask":
		var event BestBidAskEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			c.dispatchBestBidAsk(event)
		}
	case "new_market":
		var wire struct {
			ID           string        `json:"id"`
			Question     string        `json:"question"`
			Market       string        `json:"market"`
			Slug         string        `json:"slug"`
			Description  string        `json:"description"`
			AssetIDs     []string      `json:"assets_ids"`
			AssetIDsAlt  []string      `json:"asset_ids"`
			Outcomes     []string      `json:"outcomes"`
			EventMessage *EventMessage `json:"event_message"`
			Timestamp    string        `json:"timestamp"`
		}
		if err := json.Unmarshal(msgBytes, &wire); err == nil {
			assets := wire.AssetIDs
			if len(assets) == 0 {
				assets = wire.AssetIDsAlt
			}
			event := NewMarketEvent{
				ID:           wire.ID,
				Question:     wire.Question,
				Market:       wire.Market,
				Slug:         wire.Slug,
				Description:  wire.Description,
				AssetIDs:     assets,
				Outcomes:     wire.Outcomes,
				EventMessage: wire.EventMessage,
				Timestamp:    wire.Timestamp,
			}
			c.dispatchNewMarket(event)
		}
	case "market_resolved":
		var wire struct {
			ID             string        `json:"id"`
			Question       string        `json:"question"`
			Market         string        `json:"market"`
			Slug           string        `json:"slug"`
			Description    string        `json:"description"`
			AssetIDs       []string      `json:"assets_ids"`
			AssetIDsAlt    []string      `json:"asset_ids"`
			Outcomes       []string      `json:"outcomes"`
			WinningAssetID string        `json:"winning_asset_id"`
			WinningOutcome string        `json:"winning_outcome"`
			EventMessage   *EventMessage `json:"event_message"`
			Timestamp      string        `json:"timestamp"`
		}
		if err := json.Unmarshal(msgBytes, &wire); err == nil {
			assets := wire.AssetIDs
			if len(assets) == 0 {
				assets = wire.AssetIDsAlt
			}
			event := MarketResolvedEvent{
				ID:             wire.ID,
				Question:       wire.Question,
				Market:         wire.Market,
				Slug:           wire.Slug,
				Description:    wire.Description,
				AssetIDs:       assets,
				Outcomes:       wire.Outcomes,
				WinningAssetID: wire.WinningAssetID,
				WinningOutcome: wire.WinningOutcome,
				EventMessage:   wire.EventMessage,
				Timestamp:      wire.Timestamp,
			}
			c.dispatchMarketResolved(event)
		}
	case "trade":
		var event TradeEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			c.dispatchTrade(event)
		}
	case "order":
		var event OrderEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			c.dispatchOrder(event)
		}
	}
}

func trySendGlobal[T any](ch chan T, msg T) {
	if ch == nil {
		return
	}
	select {
	case ch <- msg:
	default:
	}
}

func (c *clientImpl) dispatchOrderbook(event OrderbookEvent) {
	trySendGlobal(c.orderbookCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.orderbookSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if sub.matchesAsset(event.AssetID) {
			sub.trySend(event)
		}
	}
}

func (c *clientImpl) dispatchPrice(event PriceEvent) {
	trySendGlobal(c.priceCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.priceSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		for _, priceChange := range event.PriceChanges {
			if sub.matchesAsset(priceChange.AssetId) {
				sub.trySend(priceChange)
			}
		}
	}
}

func (c *clientImpl) dispatchMidpoint(event MidpointEvent) {
	trySendGlobal(c.midpointCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.midpointSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if sub.matchesAsset(event.AssetID) {
			sub.trySend(event)
		}
	}
}

func (c *clientImpl) dispatchLastTrade(event LastTradePriceEvent) {
	trySendGlobal(c.lastTradeCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.lastTradeSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if sub.matchesAsset(event.AssetID) {
			sub.trySend(event)
		}
	}
}

func (c *clientImpl) dispatchTickSize(event TickSizeChangeEvent) {
	trySendGlobal(c.tickSizeCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.tickSizeSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if sub.matchesAsset(event.AssetID) {
			sub.trySend(event)
		}
	}
}

func (c *clientImpl) dispatchBestBidAsk(event BestBidAskEvent) {
	trySendGlobal(c.bestBidAskCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.bestBidAskSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if sub.matchesAsset(event.AssetID) {
			sub.trySend(event)
		}
	}
}

func (c *clientImpl) dispatchNewMarket(event NewMarketEvent) {
	trySendGlobal(c.newMarketCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.newMarketSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if sub.matchesAnyAsset(event.AssetIDs) {
			sub.trySend(event)
		}
	}
}

func (c *clientImpl) dispatchMarketResolved(event MarketResolvedEvent) {
	trySendGlobal(c.marketResolvedCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.marketResolvedSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if sub.matchesAnyAsset(event.AssetIDs) {
			sub.trySend(event)
		}
	}
}

func (c *clientImpl) dispatchTrade(event TradeEvent) {
	trySendGlobal(c.tradeCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.tradeSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		if event.Market != "" && !sub.matchesMarket(event.Market) {
			continue
		}
		sub.trySend(event)
	}
}

func (c *clientImpl) dispatchOrder(event OrderEvent) {
	trySendGlobal(c.orderCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.orderSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		sub.trySend(event)
	}
}

func (c *clientImpl) SubscribeOrderbookStream(ctx context.Context, assetIDs []string) (*Stream[OrderbookEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, Orderbook, false, c.orderbookSubs)
}

func (c *clientImpl) SubscribePricesStream(ctx context.Context, assetIDs []string) (*Stream[PriceChangeEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, PriceChange, false, c.priceSubs)
}

func (c *clientImpl) SubscribeMidpointsStream(ctx context.Context, assetIDs []string) (*Stream[MidpointEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, Midpoint, false, c.midpointSubs)
}

func (c *clientImpl) SubscribeLastTradePricesStream(ctx context.Context, assetIDs []string) (*Stream[LastTradePriceEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, LastTradePrice, false, c.lastTradeSubs)
}

func (c *clientImpl) SubscribeTickSizeChangesStream(ctx context.Context, assetIDs []string) (*Stream[TickSizeChangeEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, TickSizeChange, false, c.tickSizeSubs)
}

func (c *clientImpl) SubscribeBestBidAskStream(ctx context.Context, assetIDs []string) (*Stream[BestBidAskEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, BestBidAsk, true, c.bestBidAskSubs)
}

func (c *clientImpl) SubscribeNewMarketsStream(ctx context.Context, assetIDs []string) (*Stream[NewMarketEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, NewMarket, true, c.newMarketSubs)
}

func (c *clientImpl) SubscribeMarketResolutionsStream(ctx context.Context, assetIDs []string) (*Stream[MarketResolvedEvent], error) {
	return subscribeMarketStream(c, ctx, assetIDs, MarketResolved, true, c.marketResolvedSubs)
}

func (c *clientImpl) SubscribeOrdersStream(ctx context.Context) (*Stream[OrderEvent], error) {
	return nil, errors.New("markets required: use SubscribeUserOrdersStream")
}

func (c *clientImpl) SubscribeTradesStream(ctx context.Context) (*Stream[TradeEvent], error) {
	return nil, errors.New("markets required: use SubscribeUserTradesStream")
}

func (c *clientImpl) SubscribeUserOrdersStream(ctx context.Context, markets []string) (*Stream[OrderEvent], error) {
	return subscribeUserStream(c, ctx, markets, UserOrders, c.orderSubs)
}

func (c *clientImpl) SubscribeUserTradesStream(ctx context.Context, markets []string) (*Stream[TradeEvent], error) {
	return subscribeUserStream(c, ctx, markets, UserTrades, c.tradeSubs)
}

func (c *clientImpl) SubscribeOrderbook(ctx context.Context, assetIDs []string) (<-chan OrderbookEvent, error) {
	stream, err := c.SubscribeOrderbookStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribePrices(ctx context.Context, assetIDs []string) (<-chan PriceChangeEvent, error) {
	stream, err := c.SubscribePricesStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeMidpoints(ctx context.Context, assetIDs []string) (<-chan MidpointEvent, error) {
	stream, err := c.SubscribeMidpointsStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeLastTradePrices(ctx context.Context, assetIDs []string) (<-chan LastTradePriceEvent, error) {
	stream, err := c.SubscribeLastTradePricesStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeTickSizeChanges(ctx context.Context, assetIDs []string) (<-chan TickSizeChangeEvent, error) {
	stream, err := c.SubscribeTickSizeChangesStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeBestBidAsk(ctx context.Context, assetIDs []string) (<-chan BestBidAskEvent, error) {
	stream, err := c.SubscribeBestBidAskStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeNewMarkets(ctx context.Context, assetIDs []string) (<-chan NewMarketEvent, error) {
	stream, err := c.SubscribeNewMarketsStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeMarketResolutions(ctx context.Context, assetIDs []string) (<-chan MarketResolvedEvent, error) {
	stream, err := c.SubscribeMarketResolutionsStream(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeOrders(ctx context.Context) (<-chan OrderEvent, error) {
	stream, err := c.SubscribeOrdersStream(ctx)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeTrades(ctx context.Context) (<-chan TradeEvent, error) {
	stream, err := c.SubscribeTradesStream(ctx)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeUserOrders(ctx context.Context, markets []string) (<-chan OrderEvent, error) {
	stream, err := c.SubscribeUserOrdersStream(ctx, markets)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeUserTrades(ctx context.Context, markets []string) (<-chan TradeEvent, error) {
	stream, err := c.SubscribeUserTradesStream(ctx, markets)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) Subscribe(ctx context.Context, req *SubscriptionRequest) error {
	return c.applySubscription(req, OperationSubscribe)
}

func (c *clientImpl) Unsubscribe(ctx context.Context, req *SubscriptionRequest) error {
	if req == nil {
		return errors.New("subscription request is required")
	}
	req.Operation = OperationUnsubscribe
	return c.applySubscription(req, OperationUnsubscribe)
}

func (c *clientImpl) UnsubscribeMarketAssets(ctx context.Context, assetIDs []string) error {
	if len(assetIDs) == 0 {
		return errors.New("assetIDs required")
	}
	return c.Unsubscribe(ctx, NewMarketUnsubscribe(assetIDs))
}

func (c *clientImpl) UnsubscribeUserMarkets(ctx context.Context, markets []string) error {
	if len(markets) == 0 {
		return errors.New("markets required")
	}
	return c.Unsubscribe(ctx, NewUserUnsubscribe(markets))
}

func (c *clientImpl) applySubscription(req *SubscriptionRequest, defaultOp Operation) error {
	if req == nil {
		return errors.New("subscription request is required")
	}
	if req.Type == "" {
		if len(req.AssetIDs) > 0 {
			req.Type = ChannelMarket
		} else if len(req.Markets) > 0 {
			req.Type = ChannelUser
		} else {
			return errors.New("subscription type is required")
		}
	}

	switch req.Type {
	case ChannelMarket:
		if len(req.AssetIDs) == 0 {
			return errors.New("assetIDs required")
		}
	case ChannelUser:
		if len(req.Markets) == 0 {
			return errors.New("markets required")
		}
	default:
		return errors.New("unknown subscription channel")
	}

	if req.Operation == "" {
		req.Operation = defaultOp
	}
	switch req.Type {
	case ChannelMarket:
		custom := req.CustomFeatureEnabled != nil && *req.CustomFeatureEnabled
		switch req.Operation {
		case OperationSubscribe:
			newAssets := c.addMarketRefs(req.AssetIDs, custom)
			if err := c.ensureConn(ChannelMarket); err != nil {
				return err
			}
			if len(newAssets) == 0 {
				return nil
			}
			subReq := NewMarketSubscription(newAssets)
			if custom {
				subReq.WithCustomFeatures(true)
			}
			return c.writeJSON(ChannelMarket, subReq)
		case OperationUnsubscribe:
			toUnsub := c.removeMarketRefs(req.AssetIDs)
			if len(toUnsub) == 0 {
				return nil
			}
			if err := c.ensureConn(ChannelMarket); err != nil {
				return err
			}
			return c.writeJSON(ChannelMarket, NewMarketUnsubscribe(toUnsub))
		default:
			return errors.New("unknown subscription operation")
		}
	case ChannelUser:
		auth := c.resolveAuth(req.Auth)
		if auth == nil {
			return errors.New("user subscription requires API key credentials")
		}
		switch req.Operation {
		case OperationSubscribe:
			newMarkets := c.addUserRefs(req.Markets, auth)
			if err := c.ensureConn(ChannelUser); err != nil {
				return err
			}
			if len(newMarkets) == 0 {
				return nil
			}
			subReq := NewUserSubscription(newMarkets)
			subReq.Auth = auth
			return c.writeJSON(ChannelUser, subReq)
		case OperationUnsubscribe:
			toUnsub := c.removeUserRefs(req.Markets)
			if len(toUnsub) == 0 {
				return nil
			}
			if err := c.ensureConn(ChannelUser); err != nil {
				return err
			}
			unsubReq := NewUserUnsubscribe(toUnsub)
			unsubReq.Auth = auth
			return c.writeJSON(ChannelUser, unsubReq)
		default:
			return errors.New("unknown subscription operation")
		}
	default:
		return errors.New("unknown subscription channel")
	}
}

func (c *clientImpl) Close() error {
	c.closing.Store(true)
	c.cleanupSubscriptions()
	c.closeConn(ChannelMarket)
	c.closeConn(ChannelUser)
	c.setConnState(ChannelMarket, ConnectionDisconnected, 0)
	c.setConnState(ChannelUser, ConnectionDisconnected, 0)
	c.closeAllStreams()
	c.shutdown()
	return nil
}

// setReadTimeout sets the read timeout for WebSocket connections.
// This is primarily used for testing purposes.
func (c *clientImpl) setReadTimeout(timeout time.Duration) {
	c.readTimeout.Store(int64(timeout))
}

func (c *clientImpl) writeJSON(channel Channel, v interface{}) error {
	switch channel {
	case ChannelUser:
		c.userMu.Lock()
		defer c.userMu.Unlock()
		if c.userConn == nil {
			return errors.New("connection is not established")
		}
		return c.userConn.WriteJSON(v)
	default:
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.conn == nil {
			return errors.New("connection is not established")
		}
		return c.conn.WriteJSON(v)
	}
}

func (c *clientImpl) writeMessage(channel Channel, payload []byte) error {
	switch channel {
	case ChannelUser:
		c.userMu.Lock()
		defer c.userMu.Unlock()
		if c.userConn == nil {
			return errors.New("connection is not established")
		}
		return c.userConn.WriteMessage(websocket.TextMessage, payload)
	default:
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.conn == nil {
			return errors.New("connection is not established")
		}
		return c.conn.WriteMessage(websocket.TextMessage, payload)
	}
}

func subscribeMarketStream[T any](c *clientImpl, ctx context.Context, assetIDs []string, eventType EventType, custom bool, subs map[string]*subscriptionEntry[T]) (*Stream[T], error) {
	if len(assetIDs) == 0 {
		return nil, errors.New("assetIDs required")
	}
	newAssets := c.addMarketRefs(assetIDs, custom)
	if err := c.ensureConn(ChannelMarket); err != nil {
		return nil, err
	}
	if len(newAssets) > 0 {
		req := NewMarketSubscription(newAssets)
		if custom {
			req.WithCustomFeatures(true)
		}
		if err := c.writeJSON(ChannelMarket, req); err != nil {
			return nil, err
		}
	}

	entry := newSubscriptionEntry[T](c, ChannelMarket, eventType, assetIDs, nil)
	c.subMu.Lock()
	subs[entry.id] = entry
	c.subMu.Unlock()

	stream := &Stream[T]{
		C:   entry.ch,
		Err: entry.errCh,
		closeF: func() error {
			closeMarketStream(c, entry, assetIDs, subs)
			return nil
		},
	}
	bindContext(ctx, stream)
	return stream, nil
}

func subscribeUserStream[T any](c *clientImpl, ctx context.Context, markets []string, eventType EventType, subs map[string]*subscriptionEntry[T]) (*Stream[T], error) {
	if len(markets) == 0 {
		return nil, errors.New("markets required")
	}
	auth := c.resolveAuth(nil)
	if auth == nil {
		return nil, errors.New("user subscription requires API key credentials")
	}
	newMarkets := c.addUserRefs(markets, auth)
	if err := c.ensureConn(ChannelUser); err != nil {
		return nil, err
	}
	if len(newMarkets) > 0 {
		req := NewUserSubscription(newMarkets)
		req.Auth = auth
		if err := c.writeJSON(ChannelUser, req); err != nil {
			return nil, err
		}
	}

	entry := newSubscriptionEntry[T](c, ChannelUser, eventType, nil, markets)
	c.subMu.Lock()
	subs[entry.id] = entry
	c.subMu.Unlock()

	stream := &Stream[T]{
		C:   entry.ch,
		Err: entry.errCh,
		closeF: func() error {
			closeUserStream(c, entry, markets, subs)
			return nil
		},
	}
	bindContext(ctx, stream)
	return stream, nil
}

func bindContext[T any](ctx context.Context, stream *Stream[T]) {
	if ctx == nil || stream == nil {
		return
	}
	done := ctx.Done()
	if done == nil {
		return
	}
	go func() {
		<-done
		_ = stream.Close()
	}()
}

func newSubscriptionEntry[T any](c *clientImpl, channel Channel, eventType EventType, assets []string, markets []string) *subscriptionEntry[T] {
	id := atomic.AddUint64(&c.nextSubID, 1)
	return &subscriptionEntry[T]{
		id:      strconv.FormatUint(id, 10),
		channel: channel,
		event:   eventType,
		assets:  makeIDSet(assets),
		markets: makeIDSet(markets),
		ch:      make(chan T, defaultStreamBuffer),
		errCh:   make(chan error, defaultErrBuffer),
	}
}

func closeMarketStream[T any](c *clientImpl, entry *subscriptionEntry[T], assetIDs []string, subs map[string]*subscriptionEntry[T]) {
	if entry == nil {
		return
	}
	if !entry.close() {
		return
	}
	c.subMu.Lock()
	delete(subs, entry.id)
	c.subMu.Unlock()

	toUnsub := c.removeMarketRefs(assetIDs)
	if len(toUnsub) == 0 {
		return
	}
	if c.getConn(ChannelMarket) == nil {
		return
	}
	_ = c.writeJSON(ChannelMarket, NewMarketUnsubscribe(toUnsub))
}

func closeUserStream[T any](c *clientImpl, entry *subscriptionEntry[T], markets []string, subs map[string]*subscriptionEntry[T]) {
	if entry == nil {
		return
	}
	if !entry.close() {
		return
	}
	c.subMu.Lock()
	delete(subs, entry.id)
	c.subMu.Unlock()

	toUnsub := c.removeUserRefs(markets)
	if len(toUnsub) == 0 {
		return
	}
	if c.getConn(ChannelUser) == nil {
		return
	}
	auth := c.resolveAuth(nil)
	if auth == nil {
		return
	}
	req := NewUserUnsubscribe(toUnsub)
	req.Auth = auth
	_ = c.writeJSON(ChannelUser, req)
}

func (c *clientImpl) authPayload() *AuthPayload {
	if c.apiKey == nil {
		return nil
	}
	if c.apiKey.Key == "" || c.apiKey.Secret == "" || c.apiKey.Passphrase == "" {
		return nil
	}
	return &AuthPayload{
		APIKey:     c.apiKey.Key,
		Secret:     c.apiKey.Secret,
		Passphrase: c.apiKey.Passphrase,
	}
}

func (c *clientImpl) resolveAuth(explicit *AuthPayload) *AuthPayload {
	if explicit != nil {
		copy := *explicit
		return &copy
	}
	if auth := c.authPayload(); auth != nil {
		return auth
	}
	return c.getLastAuth()
}

func (c *clientImpl) getLastAuth() *AuthPayload {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	if c.lastAuth == nil {
		return nil
	}
	copy := *c.lastAuth
	return &copy
}

func (c *clientImpl) addMarketRefs(assetIDs []string, custom bool) []string {
	if len(assetIDs) == 0 {
		return nil
	}
	c.subMu.Lock()
	defer c.subMu.Unlock()
	if custom {
		c.customFeatures = true
	}
	newAssets := make([]string, 0, len(assetIDs))
	for _, id := range assetIDs {
		if id == "" {
			continue
		}
		if c.marketRefs[id] == 0 {
			newAssets = append(newAssets, id)
		}
		c.marketRefs[id]++
	}
	return newAssets
}

func (c *clientImpl) removeMarketRefs(assetIDs []string) []string {
	if len(assetIDs) == 0 {
		return nil
	}
	c.subMu.Lock()
	defer c.subMu.Unlock()
	toUnsub := make([]string, 0, len(assetIDs))
	for _, id := range assetIDs {
		count := c.marketRefs[id]
		if count <= 1 {
			if count > 0 {
				delete(c.marketRefs, id)
				toUnsub = append(toUnsub, id)
			}
			continue
		}
		c.marketRefs[id] = count - 1
	}
	return toUnsub
}

func (c *clientImpl) addUserRefs(markets []string, auth *AuthPayload) []string {
	if len(markets) == 0 {
		return nil
	}
	c.subMu.Lock()
	defer c.subMu.Unlock()
	if auth != nil {
		copy := *auth
		c.lastAuth = &copy
	}
	newMarkets := make([]string, 0, len(markets))
	for _, id := range markets {
		if id == "" {
			continue
		}
		if c.userRefs[id] == 0 {
			newMarkets = append(newMarkets, id)
		}
		c.userRefs[id]++
	}
	return newMarkets
}

func (c *clientImpl) removeUserRefs(markets []string) []string {
	if len(markets) == 0 {
		return nil
	}
	c.subMu.Lock()
	defer c.subMu.Unlock()
	toUnsub := make([]string, 0, len(markets))
	for _, id := range markets {
		count := c.userRefs[id]
		if count <= 1 {
			if count > 0 {
				delete(c.userRefs, id)
				toUnsub = append(toUnsub, id)
			}
			continue
		}
		c.userRefs[id] = count - 1
	}
	return toUnsub
}

func (c *clientImpl) snapshotSubscriptionRefs() ([]string, []string, bool, *AuthPayload) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	assets := make([]string, 0, len(c.marketRefs))
	for id := range c.marketRefs {
		assets = append(assets, id)
	}
	markets := make([]string, 0, len(c.userRefs))
	for id := range c.userRefs {
		markets = append(markets, id)
	}
	var authCopy *AuthPayload
	if c.lastAuth != nil {
		copy := *c.lastAuth
		authCopy = &copy
	}
	return assets, markets, c.customFeatures, authCopy
}

func (c *clientImpl) reconnectLoop(channel Channel) error {
	var lastErr error
	delay := c.reconnectDelay
	if delay <= 0 {
		delay = 1 * time.Second
	}
	maxDelay := c.reconnectMaxDelay
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}
	multiplier := c.reconnectMultiplier
	if multiplier <= 0 {
		multiplier = 2
	}

	for attempt := 0; c.reconnectMax <= 0 || attempt < c.reconnectMax; attempt++ {
		if c.closing.Load() {
			return lastErr
		}
		if c.debug {
			logger.Debug("ws reconnect attempt %d in %s (%s)", attempt+1, delay, channel)
		}
		c.setConnState(channel, ConnectionReconnecting, attempt+1)
		time.Sleep(delay)

		// Use init mutex to serialize with ensure* methods
		var initMu *sync.Mutex
		switch channel {
		case ChannelMarket:
			initMu = &c.marketInitMu
		case ChannelUser:
			initMu = &c.userInitMu
		}

		if initMu != nil {
			initMu.Lock()
		}

		// Cancel old goroutines and close old connection
		c.cancelGoroutines(channel)
		c.closeConn(channel)

		// Create new context for new connection's goroutines
		c.createGoroutineContext(channel)

		var err error
		switch channel {
		case ChannelMarket:
			err = c.connectMarket()
		case ChannelUser:
			err = c.connectUser()
		default:
			err = errors.New("unknown subscription channel")
		}
		if err == nil {
			if c.debug {
				logger.Debug("ws reconnect success")
			}
			c.setConnState(channel, ConnectionConnected, 0)
			c.setLastPong(channel, time.Now())

			// Restart read and ping loops after successful reconnection
			go c.readLoop(channel)
			if !c.disablePing {
				go c.pingLoop(channel)
			}

			if initMu != nil {
				initMu.Unlock()
			}

			c.resubscribe(channel)
			return nil
		}

		if initMu != nil {
			initMu.Unlock()
		}
		lastErr = err
		if c.debug {
			logger.Debug("ws reconnect failed: %v", err)
		}
		nextDelay := time.Duration(float64(delay) * multiplier)
		if nextDelay <= 0 {
			nextDelay = delay
		}
		if nextDelay > maxDelay {
			nextDelay = maxDelay
		}
		delay = nextDelay
	}
	c.setConnState(channel, ConnectionDisconnected, 0)
	return lastErr
}

func (c *clientImpl) resubscribe(channel Channel) {
	assets, markets, custom, auth := c.snapshotSubscriptionRefs()
	switch channel {
	case ChannelMarket:
		if len(assets) == 0 {
			return
		}
		req := NewMarketSubscription(assets)
		if custom {
			req.WithCustomFeatures(true)
		}
		_ = c.writeJSON(ChannelMarket, req)
	case ChannelUser:
		if len(markets) == 0 || auth == nil {
			return
		}
		req := NewUserSubscription(markets)
		req.Auth = auth
		_ = c.writeJSON(ChannelUser, req)
	}
}

func (c *clientImpl) shutdown() {
	c.closeOnce.Do(func() {
		c.closeAllStreams()
		close(c.done)
		close(c.orderbookCh)
		close(c.priceCh)
		close(c.midpointCh)
		close(c.lastTradeCh)
		close(c.tickSizeCh)
		close(c.bestBidAskCh)
		close(c.newMarketCh)
		close(c.marketResolvedCh)
		close(c.tradeCh)
		close(c.orderCh)
	})
}

func (c *clientImpl) cleanupSubscriptions() {
	assets, markets, _, auth := c.snapshotSubscriptionRefs()
	if len(assets) > 0 && c.getConn(ChannelMarket) != nil {
		req := NewMarketUnsubscribe(assets)
		_ = c.writeJSON(ChannelMarket, req)
	}
	if len(markets) > 0 && c.getConn(ChannelUser) != nil {
		if auth == nil {
			auth = c.authPayload()
		}
		if auth != nil {
			req := NewUserUnsubscribe(markets)
			req.Auth = auth
			_ = c.writeJSON(ChannelUser, req)
		}
	}
}

func (c *clientImpl) closeAllStreams() {
	c.subMu.Lock()
	closeSubMap(c.orderbookSubs)
	closeSubMap(c.priceSubs)
	closeSubMap(c.midpointSubs)
	closeSubMap(c.lastTradeSubs)
	closeSubMap(c.tickSizeSubs)
	closeSubMap(c.bestBidAskSubs)
	closeSubMap(c.newMarketSubs)
	closeSubMap(c.marketResolvedSubs)
	closeSubMap(c.tradeSubs)
	closeSubMap(c.orderSubs)
	c.subMu.Unlock()

	c.stateMu.Lock()
	closeSubMap(c.stateSubs)
	c.stateMu.Unlock()
}

func (c *clientImpl) getConn(channel Channel) *websocket.Conn {
	switch channel {
	case ChannelUser:
		c.userMu.Lock()
		conn := c.userConn
		c.userMu.Unlock()
		return conn
	default:
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		return conn
	}
}

func (c *clientImpl) setMarketConn(conn *websocket.Conn) {
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
}

func (c *clientImpl) setUserConn(conn *websocket.Conn) {
	c.userMu.Lock()
	c.userConn = conn
	c.userMu.Unlock()
}

func (c *clientImpl) closeConn(channel Channel) {
	// Cancel goroutines before closing connection
	c.cancelGoroutines(channel)

	// Atomically get and clear connection to prevent race with concurrent reconnect
	switch channel {
	case ChannelUser:
		c.userMu.Lock()
		conn := c.userConn
		c.userConn = nil
		c.userMu.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
	default:
		c.mu.Lock()
		conn := c.conn
		c.conn = nil
		c.mu.Unlock()
		if conn != nil {
			_ = conn.Close()
		}
	}
}

func (c *clientImpl) ConnectionState(channel Channel) ConnectionState {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	switch channel {
	case ChannelMarket:
		if c.marketState == "" {
			return ConnectionDisconnected
		}
		return c.marketState
	case ChannelUser:
		if c.userState == "" {
			return ConnectionDisconnected
		}
		return c.userState
	default:
		return ConnectionDisconnected
	}
}

func (c *clientImpl) ConnectionStateStream(ctx context.Context) (*Stream[ConnectionStateEvent], error) {
	entry := newSubscriptionEntry[ConnectionStateEvent](c, ChannelMarket, ConnectionStateEventType, nil, nil)
	c.stateMu.Lock()
	c.stateSubs[entry.id] = entry
	market := c.marketState
	user := c.userState
	c.stateMu.Unlock()

	stream := &Stream[ConnectionStateEvent]{
		C:   entry.ch,
		Err: entry.errCh,
		closeF: func() error {
			if entry.close() {
				c.stateMu.Lock()
				delete(c.stateSubs, entry.id)
				c.stateMu.Unlock()
			}
			return nil
		},
	}
	bindContext(ctx, stream)
	entry.trySend(ConnectionStateEvent{
		Channel:  ChannelMarket,
		State:    market,
		Recorded: time.Now().UnixMilli(),
	})
	entry.trySend(ConnectionStateEvent{
		Channel:  ChannelUser,
		State:    user,
		Recorded: time.Now().UnixMilli(),
	})
	return stream, nil
}

func (c *clientImpl) setConnState(channel Channel, state ConnectionState, attempt int) {
	event := ConnectionStateEvent{
		Channel:  channel,
		State:    state,
		Attempt:  attempt,
		Recorded: time.Now().UnixMilli(),
	}

	c.stateMu.Lock()
	switch channel {
	case ChannelMarket:
		c.marketState = state
	case ChannelUser:
		c.userState = state
	default:
		c.stateMu.Unlock()
		return
	}
	subs := snapshotSubs(c.stateSubs)
	c.stateMu.Unlock()

	for _, sub := range subs {
		sub.trySend(event)
	}
}

// createGoroutineContext creates a new context for managing goroutines for a specific channel.
// Must be called before starting readLoop and pingLoop goroutines.
func (c *clientImpl) createGoroutineContext(channel Channel) {
	c.goroutineCtxMu.Lock()
	defer c.goroutineCtxMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	switch channel {
	case ChannelMarket:
		c.marketCtx = ctx
		c.marketCancel = cancel
	case ChannelUser:
		c.userCtx = ctx
		c.userCancel = cancel
	default:
		// Cancel immediately if channel is unknown to prevent context leak
		cancel()
	}
}

// cancelGoroutines cancels the context for a specific channel, signaling all associated
// goroutines (readLoop, pingLoop) to exit gracefully.
func (c *clientImpl) cancelGoroutines(channel Channel) {
	c.goroutineCtxMu.Lock()
	defer c.goroutineCtxMu.Unlock()

	switch channel {
	case ChannelMarket:
		if c.marketCancel != nil {
			c.marketCancel()
			c.marketCancel = nil
		}
	case ChannelUser:
		if c.userCancel != nil {
			c.userCancel()
			c.userCancel = nil
		}
	}
}

// getGoroutineContext returns the context for a specific channel's goroutines.
func (c *clientImpl) getGoroutineContext(channel Channel) context.Context {
	c.goroutineCtxMu.Lock()
	defer c.goroutineCtxMu.Unlock()

	switch channel {
	case ChannelMarket:
		return c.marketCtx
	case ChannelUser:
		return c.userCtx
	default:
		return nil
	}
}
