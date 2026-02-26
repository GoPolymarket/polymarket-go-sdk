package ws

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/logger"
	"github.com/gorilla/websocket"
)

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
