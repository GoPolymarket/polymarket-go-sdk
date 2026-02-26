package ws

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
)

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
