package rtds

import (
	"errors"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/pkg/errors"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/logger"
	"github.com/gorilla/websocket"
)

const (
	ProdURL = "wss://ws-live-data.polymarket.com"
)

const (
	connDisconnected int32 = iota
	connConnected
)

// Use unified error definitions from pkg/errors
var (
	ErrInvalidSubscription = sdkerrors.ErrInvalidSubscription
)

const (
	defaultStreamBuffer = 100
	defaultErrBuffer    = 10
)

type clientImpl struct {
	url       string
	conn      *websocket.Conn
	mu        sync.Mutex
	done      chan struct{}
	connReady chan struct{}
	connOnce  sync.Once
	state     int32
	closeOnce sync.Once
	closing   atomic.Bool

	reconnect      bool
	reconnectDelay time.Duration
	reconnectMax   int
	pingInterval   time.Duration

	stateMu     sync.Mutex
	stateSubs   map[string]*stateSubscription
	nextStateID uint64

	subMu      sync.Mutex
	subRefs    map[string]int
	subDetails map[string]Subscription
	subs       map[string]*subscriptionEntry
	subsByKey  map[string]map[string]*subscriptionEntry
	nextSubID  uint64

	authMu sync.RWMutex
	auth   *auth.APIKey
}

func NewClient(url string) (Client, error) {
	return NewClientWithConfig(url, ClientConfigFromEnv())
}

// NewClientWithConfig creates an RTDS client with explicit reconnect/heartbeat settings.
func NewClientWithConfig(url string, cfg ClientConfig) (Client, error) {
	if url == "" {
		url = ProdURL
	}
	if err := validateRTDSURL(url); err != nil {
		return nil, err
	}
	cfg = cfg.normalize()

	c := &clientImpl{
		url:            url,
		done:           make(chan struct{}),
		connReady:      make(chan struct{}),
		stateSubs:      make(map[string]*stateSubscription),
		subRefs:        make(map[string]int),
		subDetails:     make(map[string]Subscription),
		subs:           make(map[string]*subscriptionEntry),
		subsByKey:      make(map[string]map[string]*subscriptionEntry),
		reconnect:      cfg.Reconnect,
		reconnectDelay: cfg.ReconnectDelay,
		reconnectMax:   cfg.ReconnectMax,
		pingInterval:   cfg.PingInterval,
	}

	go c.run()
	go c.pingLoop()

	return c, nil
}

func validateRTDSURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	switch parsed.Scheme {
	case "ws", "wss":
	default:
		return errors.New("rtds URL must use ws:// or wss://")
	}
	if parsed.Host == "" {
		return errors.New("rtds URL host is required")
	}
	return nil
}

func (c *clientImpl) Authenticate(apiKey *auth.APIKey) Client {
	c.authMu.Lock()
	c.auth = apiKey
	c.authMu.Unlock()
	return c
}

func (c *clientImpl) Deauthenticate() Client {
	c.authMu.Lock()
	c.auth = nil
	c.authMu.Unlock()
	return c
}

func (c *clientImpl) connect() error {
	c.closeConn()
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		c.setState(ConnectionDisconnected)
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.setState(ConnectionConnected)
	c.connOnce.Do(func() { close(c.connReady) })
	return nil
}

func (c *clientImpl) run() {
	attempts := 0
	for {
		if c.closing.Load() {
			c.signalDone()
			return
		}
		if err := c.connect(); err != nil {
			if !c.shouldReconnect(attempts) {
				c.signalDone()
				return
			}
			attempts++
			time.Sleep(c.reconnectDelay)
			continue
		}

		attempts = 0
		c.resubscribeAll()

		if err := c.readLoop(); err != nil {
			if c.closing.Load() {
				c.signalDone()
				return
			}
			if !c.shouldReconnect(attempts) {
				c.signalDone()
				return
			}
			attempts++
			time.Sleep(c.reconnectDelay)
			continue
		}
	}
}

func (c *clientImpl) shouldReconnect(attempts int) bool {
	if !c.reconnect {
		return false
	}
	if c.reconnectMax == 0 {
		return true
	}
	return attempts < c.reconnectMax
}

func (c *clientImpl) pingLoop() {
	interval := c.pingInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.conn == nil {
				c.mu.Unlock()
				continue
			}
			// RTDS expects a "PING" text message
			err := c.conn.WriteMessage(websocket.TextMessage, []byte("PING"))
			c.mu.Unlock()
			if err != nil {
				c.setState(ConnectionDisconnected)
				continue
			}
		}
	}
}

func (c *clientImpl) readLoop() error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return errors.New("connection not established")
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Error("rtds read error: %v", err)
			c.setState(ConnectionDisconnected)
			return err
		}

		msgs, err := parseMessages(message)
		if err != nil {
			continue
		}

		for _, msg := range msgs {
			c.dispatch(msg)
		}
	}
}

func (c *clientImpl) dispatch(msg RtdsMessage) {
	c.subMu.Lock()
	subs := make([]*subscriptionEntry, 0, len(c.subs))
	for _, sub := range c.subs {
		subs = append(subs, sub)
	}
	c.subMu.Unlock()

	for _, sub := range subs {
		if sub.matches(msg) {
			sub.trySend(msg)
		}
	}
}

func (c *clientImpl) ConnectionState() ConnectionState {
	if atomic.LoadInt32(&c.state) == connConnected {
		return ConnectionConnected
	}
	return ConnectionDisconnected
}

func (c *clientImpl) SubscriptionCount() int {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	return len(c.subs)
}

func (c *clientImpl) Close() error {
	c.closing.Store(true)
	c.setState(ConnectionDisconnected)
	c.closeConn()
	c.closeAllSubscriptions()
	c.closeStateSubscriptions()
	c.signalDone()
	return nil
}

func (c *clientImpl) writeJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return errors.New("connection not established")
	}
	return c.conn.WriteJSON(v)
}

func (c *clientImpl) closeConn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

func (c *clientImpl) signalDone() {
	c.closeOnce.Do(func() {
		close(c.done)
	})
}

func (c *clientImpl) sendSubscriptions(action SubscriptionAction, subs []Subscription) error {
	if len(subs) == 0 {
		return nil
	}
	return c.writeJSON(SubscriptionRequest{
		Action:        action,
		Subscriptions: subs,
	})
}

func (c *clientImpl) resubscribeAll() {
	c.subMu.Lock()
	subs := make([]Subscription, 0, len(c.subDetails))
	for _, sub := range c.subDetails {
		subs = append(subs, sub)
	}
	c.subMu.Unlock()
	if len(subs) == 0 {
		return
	}
	if err := c.sendSubscriptions(SubscribeAction, subs); err != nil {
		logger.Error("rtds resubscribe failed: %v", err)
	}
}
