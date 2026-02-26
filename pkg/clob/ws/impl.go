package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/logger"

	"github.com/gorilla/websocket"
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
	return NewClientWithConfig(url, signer, apiKey, ClientConfigFromEnv())
}

// NewClientWithConfig creates a new WebSocket client using explicit configuration.
func NewClientWithConfig(url string, signer auth.Signer, apiKey *auth.APIKey, cfg ClientConfig) (Client, error) {
	c := newClientImpl(url, signer, apiKey, cfg)
	if err := c.ensureMarketConn(); err != nil {
		return nil, err
	}
	return c, nil
}

func newClientImpl(url string, signer auth.Signer, apiKey *auth.APIKey, cfg ClientConfig) *clientImpl {
	marketURL, userURL, baseURL := normalizeWSURLs(url)
	cfg = cfg.normalize()

	c := &clientImpl{
		baseURL:             baseURL,
		marketURL:           marketURL,
		userURL:             userURL,
		signer:              signer,
		apiKey:              apiKey,
		debug:               cfg.Debug,
		disablePing:         cfg.DisablePing,
		reconnect:           cfg.Reconnect,
		reconnectDelay:      cfg.ReconnectDelay,
		reconnectMaxDelay:   cfg.ReconnectMaxDelay,
		reconnectMultiplier: cfg.ReconnectMultiplier,
		reconnectMax:        cfg.ReconnectMax,
		heartbeatInterval:   cfg.HeartbeatInterval,
		heartbeatTimeout:    cfg.HeartbeatTimeout,
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
	c.readTimeout.Store(int64(cfg.ReadTimeout))
	return c
}

// Clone returns a detached copy for auth scoping without mutating the original client instance.
func (c *clientImpl) Clone() Client {
	if c == nil {
		return nil
	}
	cfg := ClientConfig{
		Debug:               c.debug,
		DisablePing:         c.disablePing,
		Reconnect:           c.reconnect,
		ReconnectDelay:      c.reconnectDelay,
		ReconnectMaxDelay:   c.reconnectMaxDelay,
		ReconnectMultiplier: c.reconnectMultiplier,
		ReconnectMax:        c.reconnectMax,
		HeartbeatInterval:   c.heartbeatInterval,
		HeartbeatTimeout:    c.heartbeatTimeout,
		ReadTimeout:         time.Duration(c.readTimeout.Load()),
	}
	clone := newClientImpl(c.baseURL, c.signer, c.apiKey, cfg)
	if auth := c.getLastAuth(); auth != nil {
		clone.lastAuth = auth
	}
	return clone
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
