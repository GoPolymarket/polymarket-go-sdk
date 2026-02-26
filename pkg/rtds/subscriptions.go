package rtds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
)

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

	// Wait for the WebSocket connection to be established before subscribing.
	select {
	case <-c.connReady:
	case <-c.done:
		return nil, errors.New("client closed before connection was established")
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
