package ws

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

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
	defer func() {
		_ = recover() // safe guard against send on closed channel during shutdown
	}()
	select {
	case ch <- msg:
	default:
	}
}

func (c *clientImpl) dispatchOrderbook(event OrderbookEvent) {
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
	trySendGlobal(c.priceCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.priceSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		for _, priceChange := range event.PriceChanges {
			if sub.matchesAsset(priceChange.AssetID) {
				sub.trySend(priceChange)
			}
		}
	}
}

func (c *clientImpl) dispatchMidpoint(event MidpointEvent) {
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
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
	if c.closing.Load() {
		return
	}
	trySendGlobal(c.orderCh, event)
	c.subMu.Lock()
	subs := snapshotSubs(c.orderSubs)
	c.subMu.Unlock()
	for _, sub := range subs {
		sub.trySend(event)
	}
}
