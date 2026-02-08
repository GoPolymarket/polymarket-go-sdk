package rtds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
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

type subscriptionEntry struct {
	id        string
	key       string
	topic     string
	msgType   string
	filter    func(RtdsMessage) bool
	ch        chan RtdsMessage
	errCh     chan error
	closed    atomic.Bool
	closeOnce sync.Once
}

func (s *subscriptionEntry) matches(msg RtdsMessage) bool {
	if msg.Topic != s.topic {
		return false
	}
	if s.msgType != "*" && msg.MsgType != s.msgType {
		return false
	}
	if s.filter != nil {
		return s.filter(msg)
	}
	return true
}

func (s *subscriptionEntry) trySend(msg RtdsMessage) {
	if s.closed.Load() {
		return
	}
	defer func() {
		_ = recover()
	}()
	select {
	case s.ch <- msg:
		return
	default:
		s.notifyLag(1)
	}
}

func (s *subscriptionEntry) notifyLag(count int) {
	if count <= 0 {
		return
	}
	err := LaggedError{Count: count, Topic: s.topic, MsgType: s.msgType}
	select {
	case s.errCh <- err:
	default:
	}
}

func (s *subscriptionEntry) close() {
	if s.closed.Swap(true) {
		return
	}
	s.closeOnce.Do(func() {
		close(s.ch)
		close(s.errCh)
	})
}

type clientImpl struct {
	url       string
	conn      *websocket.Conn
	mu        sync.Mutex
	done      chan struct{}
	state     int32
	closeOnce sync.Once
	closing   atomic.Bool

	reconnect      bool
	reconnectDelay time.Duration
	reconnectMax   int

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
	if url == "" {
		url = ProdURL
	}

	reconnect := true
	if raw := strings.TrimSpace(os.Getenv("RTDS_WS_RECONNECT")); raw != "" {
		reconnect = raw != "0" && strings.ToLower(raw) != "false"
	}
	reconnectDelay := 2 * time.Second
	if raw := strings.TrimSpace(os.Getenv("RTDS_WS_RECONNECT_DELAY_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			reconnectDelay = time.Duration(ms) * time.Millisecond
		}
	}
	reconnectMax := 5
	if raw := strings.TrimSpace(os.Getenv("RTDS_WS_RECONNECT_MAX")); raw != "" {
		if max, err := strconv.Atoi(raw); err == nil {
			reconnectMax = max
		}
	}

	c := &clientImpl{
		url:            url,
		done:           make(chan struct{}),
		stateSubs:      make(map[string]*stateSubscription),
		subRefs:        make(map[string]int),
		subDetails:     make(map[string]Subscription),
		subs:           make(map[string]*subscriptionEntry),
		subsByKey:      make(map[string]map[string]*subscriptionEntry),
		reconnect:      reconnect,
		reconnectDelay: reconnectDelay,
		reconnectMax:   reconnectMax,
	}

	go c.run()
	go c.pingLoop()

	return c, nil
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
	c.conn = conn
	c.setState(ConnectionConnected)
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
	ticker := time.NewTicker(5 * time.Second)
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
	for {
		_, message, err := c.conn.ReadMessage()
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

func (c *clientImpl) SubscribeCryptoPricesStream(ctx context.Context, symbols []string) (*Stream[CryptoPriceEvent], error) {
	sub := Subscription{Topic: string(CryptoPrice), MsgType: "update"}
	if len(symbols) > 0 {
		sub.Filters = symbols
	}
	rawStream, err := c.subscribeRawStream(sub, nil)
	if err != nil {
		return nil, err
	}
	set := symbolSet(symbols)
	return mapStream(rawStream, sub.Topic, sub.MsgType, func(msg RtdsMessage) (CryptoPriceEvent, bool) {
		var payload CryptoPriceEvent
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return CryptoPriceEvent{}, false
		}
		if len(set) > 0 {
			if _, ok := set[strings.ToLower(payload.Symbol)]; !ok {
				return CryptoPriceEvent{}, false
			}
		}
		payload.BaseEvent = BaseEvent{
			Topic:            CryptoPrice,
			MessageType:      msg.MsgType,
			MessageTimestamp: msg.Timestamp,
		}
		return payload, true
	}), nil
}

func (c *clientImpl) SubscribeChainlinkPricesStream(ctx context.Context, feeds []string) (*Stream[ChainlinkPriceEvent], error) {
	msgType := "*"
	sub := Subscription{Topic: string(ChainlinkPrice), MsgType: msgType}
	if len(feeds) == 1 {
		filterMap := map[string]string{"symbol": feeds[0]}
		if filterBytes, err := json.Marshal(filterMap); err == nil {
			sub.Filters = string(filterBytes)
		}
	}
	rawStream, err := c.subscribeRawStream(sub, nil)
	if err != nil {
		return nil, err
	}
	set := symbolSet(feeds)
	return mapStream(rawStream, sub.Topic, sub.MsgType, func(msg RtdsMessage) (ChainlinkPriceEvent, bool) {
		var payload ChainlinkPriceEvent
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return ChainlinkPriceEvent{}, false
		}
		if len(set) > 0 {
			if _, ok := set[strings.ToLower(payload.Symbol)]; !ok {
				return ChainlinkPriceEvent{}, false
			}
		}
		payload.BaseEvent = BaseEvent{
			Topic:            ChainlinkPrice,
			MessageType:      msg.MsgType,
			MessageTimestamp: msg.Timestamp,
		}
		return payload, true
	}), nil
}

func (c *clientImpl) SubscribeCommentsStream(ctx context.Context, req *CommentFilter) (*Stream[CommentEvent], error) {
	msgType := "*"
	sub := Subscription{Topic: string(Comments), MsgType: msgType}
	if req != nil {
		if req.Type != nil {
			msgType = string(*req.Type)
			sub.MsgType = msgType
		}
		if req.Auth != nil {
			sub.ClobAuth = &ClobAuth{
				Key:        req.Auth.Key,
				Secret:     req.Auth.Secret,
				Passphrase: req.Auth.Passphrase,
			}
		}
		if req.Filters != nil {
			sub.Filters = req.Filters
		}
	}
	if sub.ClobAuth == nil {
		c.authMu.RLock()
		authCopy := c.auth
		c.authMu.RUnlock()
		if authCopy != nil {
			sub.ClobAuth = &ClobAuth{
				Key:        authCopy.Key,
				Secret:     authCopy.Secret,
				Passphrase: authCopy.Passphrase,
			}
		}
	}
	rawStream, err := c.subscribeRawStream(sub, nil)
	if err != nil {
		return nil, err
	}
	return mapStream(rawStream, sub.Topic, sub.MsgType, func(msg RtdsMessage) (CommentEvent, bool) {
		var payload CommentEvent
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return CommentEvent{}, false
		}
		payload.BaseEvent = BaseEvent{
			Topic:            Comments,
			MessageType:      msg.MsgType,
			MessageTimestamp: msg.Timestamp,
		}
		return payload, true
	}), nil
}

func (c *clientImpl) SubscribeOrdersMatchedStream(ctx context.Context) (*Stream[OrdersMatchedEvent], error) {
	sub := Subscription{Topic: string(Activity), MsgType: "orders_matched"}
	rawStream, err := c.subscribeRawStream(sub, nil)
	if err != nil {
		return nil, err
	}
	return mapStream(rawStream, sub.Topic, sub.MsgType, func(msg RtdsMessage) (OrdersMatchedEvent, bool) {
		var payload OrdersMatchedEvent
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return OrdersMatchedEvent{}, false
		}
		payload.BaseEvent = BaseEvent{
			Topic:            Activity,
			MessageType:      msg.MsgType,
			MessageTimestamp: msg.Timestamp,
		}
		return payload, true
	}), nil
}

func (c *clientImpl) SubscribeRawStream(ctx context.Context, sub *Subscription) (*Stream[RtdsMessage], error) {
	if sub == nil {
		return nil, ErrInvalidSubscription
	}
	return c.subscribeRawStream(*sub, nil)
}

func (c *clientImpl) SubscribeCryptoPrices(ctx context.Context, symbols []string) (<-chan CryptoPriceEvent, error) {
	stream, err := c.SubscribeCryptoPricesStream(ctx, symbols)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeChainlinkPrices(ctx context.Context, feeds []string) (<-chan ChainlinkPriceEvent, error) {
	stream, err := c.SubscribeChainlinkPricesStream(ctx, feeds)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeComments(ctx context.Context, req *CommentFilter) (<-chan CommentEvent, error) {
	stream, err := c.SubscribeCommentsStream(ctx, req)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeOrdersMatched(ctx context.Context) (<-chan OrdersMatchedEvent, error) {
	stream, err := c.SubscribeOrdersMatchedStream(ctx)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) SubscribeRaw(ctx context.Context, sub *Subscription) (<-chan RtdsMessage, error) {
	stream, err := c.SubscribeRawStream(ctx, sub)
	if err != nil {
		return nil, err
	}
	return stream.C, nil
}

func (c *clientImpl) UnsubscribeCryptoPrices(ctx context.Context) error {
	topic := string(CryptoPrice)
	msgType := "update"
	return c.unsubscribeTopic(topic, msgType)
}

func (c *clientImpl) UnsubscribeChainlinkPrices(ctx context.Context) error {
	topic := string(ChainlinkPrice)
	msgType := "*"
	return c.unsubscribeTopic(topic, msgType)
}

func (c *clientImpl) UnsubscribeComments(ctx context.Context, commentType *CommentType) error {
	msgType := "*"
	if commentType != nil {
		msgType = string(*commentType)
	}
	return c.unsubscribeTopic(string(Comments), msgType)
}

func (c *clientImpl) UnsubscribeOrdersMatched(ctx context.Context) error {
	return c.unsubscribeTopic(string(Activity), "orders_matched")
}

func (c *clientImpl) UnsubscribeRaw(ctx context.Context, sub *Subscription) error {
	if sub == nil {
		return ErrInvalidSubscription
	}
	return c.unsubscribeTopic(sub.Topic, sub.MsgType)
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

func subscriptionKey(topic, msgType string) string {
	return topic + "|" + msgType
}

func (c *clientImpl) subscribeRawStream(sub Subscription, filter func(RtdsMessage) bool) (*Stream[RtdsMessage], error) {
	entry, err := c.subscribeRaw(sub, filter)
	if err != nil {
		return nil, err
	}
	stream := &Stream[RtdsMessage]{
		C:   entry.ch,
		Err: entry.errCh,
		closeF: func() error {
			return c.unsubscribeByID(entry.id)
		},
	}
	return stream, nil
}

func (c *clientImpl) subscribeRaw(sub Subscription, filter func(RtdsMessage) bool) (*subscriptionEntry, error) {
	if strings.TrimSpace(sub.Topic) == "" || strings.TrimSpace(sub.MsgType) == "" {
		return nil, ErrInvalidSubscription
	}
	key := subscriptionKey(sub.Topic, sub.MsgType)

	c.subMu.Lock()
	defer c.subMu.Unlock()

	if c.subRefs[key] == 0 {
		if err := c.sendSubscriptions(SubscribeAction, []Subscription{sub}); err != nil {
			return nil, err
		}
	}

	c.subRefs[key]++
	c.subDetails[key] = sub

	id := fmt.Sprintf("%s#%d", key, atomic.AddUint64(&c.nextSubID, 1))
	entry := &subscriptionEntry{
		id:      id,
		key:     key,
		topic:   sub.Topic,
		msgType: sub.MsgType,
		filter:  filter,
		ch:      make(chan RtdsMessage, defaultStreamBuffer),
		errCh:   make(chan error, defaultErrBuffer),
	}
	c.subs[id] = entry
	if c.subsByKey[key] == nil {
		c.subsByKey[key] = make(map[string]*subscriptionEntry)
	}
	c.subsByKey[key][id] = entry

	return entry, nil
}

func (c *clientImpl) unsubscribeByID(id string) error {
	c.subMu.Lock()
	entry := c.subs[id]
	if entry == nil {
		c.subMu.Unlock()
		return nil
	}
	delete(c.subs, id)
	if byKey, ok := c.subsByKey[entry.key]; ok {
		delete(byKey, id)
		if len(byKey) == 0 {
			delete(c.subsByKey, entry.key)
		}
	}

	shouldUnsub := false
	if count := c.subRefs[entry.key]; count <= 1 {
		delete(c.subRefs, entry.key)
		delete(c.subDetails, entry.key)
		shouldUnsub = true
	} else {
		c.subRefs[entry.key] = count - 1
	}

	var sendErr error
	if shouldUnsub {
		sub := Subscription{Topic: entry.topic, MsgType: entry.msgType}
		sendErr = c.sendSubscriptions(UnsubscribeAction, []Subscription{sub})
	}
	c.subMu.Unlock()

	entry.close()
	return sendErr
}

func (c *clientImpl) unsubscribeTopic(topic, msgType string) error {
	key := subscriptionKey(topic, msgType)
	c.subMu.Lock()
	byKey := c.subsByKey[key]
	var entry *subscriptionEntry
	for _, sub := range byKey {
		entry = sub
		break
	}
	c.subMu.Unlock()
	if entry == nil {
		return nil
	}
	return c.unsubscribeByID(entry.id)
}

func (c *clientImpl) closeAllSubscriptions() {
	c.subMu.Lock()
	subs := make([]*subscriptionEntry, 0, len(c.subs))
	for _, sub := range c.subs {
		subs = append(subs, sub)
	}
	c.subs = make(map[string]*subscriptionEntry)
	c.subsByKey = make(map[string]map[string]*subscriptionEntry)
	c.subRefs = make(map[string]int)
	c.subDetails = make(map[string]Subscription)
	c.subMu.Unlock()

	for _, sub := range subs {
		sub.close()
	}
}

func symbolSet(symbols []string) map[string]struct{} {
	if len(symbols) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(symbols))
	for _, s := range symbols {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			continue
		}
		set[s] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

func mapStream[T any](src *Stream[RtdsMessage], topic, msgType string, mapFn func(RtdsMessage) (T, bool)) *Stream[T] {
	outC := make(chan T, defaultStreamBuffer)
	errC := make(chan error, defaultErrBuffer)

	go func() {
		defer close(outC)
		defer close(errC)
		for {
			select {
			case msg, ok := <-src.C:
				if !ok {
					return
				}
				mapped, ok := mapFn(msg)
				if !ok {
					continue
				}
				select {
				case outC <- mapped:
				default:
					select {
					case errC <- LaggedError{Count: 1, Topic: topic, MsgType: msgType}:
					default:
					}
				}
			case err, ok := <-src.Err:
				if !ok {
					return
				}
				select {
				case errC <- err:
				default:
				}
			}
		}
	}()

	return &Stream[T]{
		C:      outC,
		Err:    errC,
		closeF: src.Close,
	}
}

func parseMessages(message []byte) ([]RtdsMessage, error) {
	trimmed := bytes.TrimSpace(message)
	if len(trimmed) == 0 {
		return nil, nil
	}
	if trimmed[0] == '[' {
		var msgs []RtdsMessage
		if err := json.Unmarshal(trimmed, &msgs); err != nil {
			return nil, err
		}
		return msgs, nil
	}
	var msg RtdsMessage
	if err := json.Unmarshal(trimmed, &msg); err != nil {
		return nil, err
	}
	return []RtdsMessage{msg}, nil
}
