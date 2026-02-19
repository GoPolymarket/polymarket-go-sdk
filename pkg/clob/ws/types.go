package ws

// Event types.

type EventType string

const (
	Orderbook                EventType = "orderbook"
	Price                    EventType = "price"
	PriceChange              EventType = "price_change"
	Midpoint                 EventType = "midpoint"
	LastTrade                EventType = "trade" // user trade message
	LastTradePrice           EventType = "last_trade_price"
	TickSizeChange           EventType = "tick_size_change"
	BestBidAsk               EventType = "best_bid_ask"
	NewMarket                EventType = "new_market"
	MarketResolved           EventType = "market_resolved"
	UserOrders               EventType = "orders"
	UserTrades               EventType = "trades"
	ConnectionStateEventType EventType = "connection_state"
)

type Operation string

const (
	OperationSubscribe   Operation = "subscribe"
	OperationUnsubscribe Operation = "unsubscribe"
)

type Channel string

const (
	ChannelMarket    Channel = "market"
	ChannelUser      Channel = "user"
	ChannelSubscribe Channel = "subscribe"
)

// ConnectionState represents CLOB WS connection status.
type ConnectionState string

const (
	ConnectionDisconnected ConnectionState = "disconnected"
	ConnectionConnecting   ConnectionState = "connecting"
	ConnectionConnected    ConnectionState = "connected"
	ConnectionReconnecting ConnectionState = "reconnecting"
)

// ConnectionStateEvent captures connection transitions.
type ConnectionStateEvent struct {
	Channel  Channel         `json:"channel"`
	State    ConnectionState `json:"state"`
	Attempt  int             `json:"attempt,omitempty"`
	Recorded int64           `json:"recorded"`
}

type AuthPayload struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// SubscriptionRequest matches the CLOB WS subscription format.
type SubscriptionRequest struct {
	Type                 Channel      `json:"type"`
	Operation            Operation    `json:"operation,omitempty"`
	Markets              []string     `json:"markets,omitempty"`
	AssetIDs             []string     `json:"assets_ids,omitempty"`
	InitialDump          *bool        `json:"initial_dump,omitempty"`
	CustomFeatureEnabled *bool        `json:"custom_feature_enabled,omitempty"`
	Auth                 *AuthPayload `json:"auth,omitempty"`
}

func NewMarketSubscription(assetIDs []string) *SubscriptionRequest {
	initial := true
	return &SubscriptionRequest{
		Type:        ChannelMarket,
		Operation:   OperationSubscribe,
		AssetIDs:    assetIDs,
		InitialDump: &initial,
	}
}

func NewMarketUnsubscribe(assetIDs []string) *SubscriptionRequest {
	return &SubscriptionRequest{
		Type:      ChannelMarket,
		Operation: OperationUnsubscribe,
		AssetIDs:  assetIDs,
	}
}

func NewUserSubscription(markets []string) *SubscriptionRequest {
	initial := true
	return &SubscriptionRequest{
		Type:        ChannelSubscribe,
		Operation:   OperationSubscribe,
		Markets:     markets,
		InitialDump: &initial,
	}
}

func NewUserUnsubscribe(markets []string) *SubscriptionRequest {
	return &SubscriptionRequest{
		Type:      ChannelUser,
		Operation: OperationUnsubscribe,
		Markets:   markets,
	}
}

func (r *SubscriptionRequest) WithCustomFeatures(enabled bool) *SubscriptionRequest {
	if r == nil {
		return nil
	}
	r.CustomFeatureEnabled = &enabled
	return r
}

type BaseEvent struct {
	Type      EventType `json:"type"`
	Timestamp int64     `json:"timestamp,omitempty"`
}

type OrderbookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type OrderbookEvent struct {
	AssetID   string           `json:"asset_id"`
	Market    string           `json:"market,omitempty"`
	Bids      []OrderbookLevel `json:"bids"`
	Asks      []OrderbookLevel `json:"asks"`
	Hash      string           `json:"hash"`
	Timestamp string           `json:"timestamp"` // Sometimes string in JSON
}

type PriceEvent struct {
	Market       string             `json:"market"`
	PriceChanges []PriceChangeEvent `json:"price_changes"`
	Timestamp    string             `json:"timestamp"`
}

type PriceChangeEvent struct {
	AssetId string `json:"asset_id"`
	BestAsk string `json:"best_ask"`
	BestBid string `json:"best_bid"`
	Hash    string `json:"hash"`
	Price   string `json:"price"`
	Side    string `json:"side"`
	Size    string `json:"size"`
}

type MidpointEvent struct {
	AssetID  string `json:"asset_id"`
	Midpoint string `json:"midpoint"`
}

type TickSizeChangeEvent struct {
	AssetID         string `json:"asset_id"`
	Market          string `json:"market,omitempty"`
	TickSize        string `json:"tick_size,omitempty"`
	MinimumTickSize string `json:"minimum_tick_size,omitempty"`
	Timestamp       string `json:"timestamp,omitempty"`
}

type LastTradePriceEvent struct {
	AssetID    string `json:"asset_id"`
	Market     string `json:"market,omitempty"`
	Price      string `json:"price"`
	Side       string `json:"side,omitempty"`
	Size       string `json:"size,omitempty"`
	FeeRateBps string `json:"fee_rate_bps,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
}

type BestBidAskEvent struct {
	Market    string `json:"market,omitempty"`
	AssetID   string `json:"asset_id"`
	BestBid   string `json:"best_bid,omitempty"`
	BestAsk   string `json:"best_ask,omitempty"`
	Spread    string `json:"spread,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

type EventMessage struct {
	ID          string `json:"id"`
	Ticker      string `json:"ticker"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type NewMarketEvent struct {
	ID           string        `json:"id"`
	Question     string        `json:"question"`
	Market       string        `json:"market,omitempty"`
	Slug         string        `json:"slug,omitempty"`
	Description  string        `json:"description,omitempty"`
	AssetIDs     []string      `json:"assets_ids,omitempty"`
	Outcomes     []string      `json:"outcomes,omitempty"`
	EventMessage *EventMessage `json:"event_message,omitempty"`
	Timestamp    string        `json:"timestamp,omitempty"`
}

type MarketResolvedEvent struct {
	ID             string        `json:"id"`
	Question       string        `json:"question"`
	Market         string        `json:"market,omitempty"`
	Slug           string        `json:"slug,omitempty"`
	Description    string        `json:"description,omitempty"`
	AssetIDs       []string      `json:"assets_ids,omitempty"`
	Outcomes       []string      `json:"outcomes,omitempty"`
	WinningAssetID string        `json:"winning_asset_id,omitempty"`
	WinningOutcome string        `json:"winning_outcome,omitempty"`
	EventMessage   *EventMessage `json:"event_message,omitempty"`
	Timestamp      string        `json:"timestamp,omitempty"`
}

type TradeEvent struct {
	AssetID   string `json:"asset_id"`
	Price     string `json:"price"`
	Size      string `json:"size"`
	Side      string `json:"side"`
	Timestamp string `json:"timestamp"`
	ID        string `json:"id,omitempty"`
	Market    string `json:"market,omitempty"`
	Status    string `json:"status,omitempty"`
}

type OrderEvent struct {
	ID              string   `json:"id"`
	AssetID         string   `json:"asset_id"`
	Market          string   `json:"market"`
	Side            string   `json:"side"`
	Price           string   `json:"price"`
	OriginalSize    string   `json:"original_size"`
	SizeMatched     string   `json:"size_matched"`
	Status          string   `json:"status"` // LIVE, CANCELED, MATCHED
	Type            string   `json:"type"`   // PLACEMENT, UPDATE, CANCELLATION
	Outcome         string   `json:"outcome"`
	OrderOwner      string   `json:"order_owner"`
	Owner           string   `json:"owner"`
	Timestamp       string   `json:"timestamp"` // string
	CreatedAt       string   `json:"created_at"`
	Expiration      string   `json:"expiration"`
	OrderType       string   `json:"order_type"` // GTC, FOK, etc
	MakerAddress    string   `json:"maker_address"`
	AssociateTrades []string `json:"associate_trades"`
	EventType       string   `json:"event_type"`
}
