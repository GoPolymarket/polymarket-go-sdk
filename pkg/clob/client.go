// Package clob provides the client for interacting with the Polymarket Central Limit Order Book.
// It handles order placement, market data retrieval, account management, and real-time streaming.
package clob

import (
	"context"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/heartbeat"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/rfq"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

// Client defines the primary interface for interacting with the Polymarket CLOB.
// It supports both unauthenticated public data access and authenticated private operations.
type Client interface {
	// -- Authentication & Configuration --

	// WithAuth returns a new client instance configured with the provided signer and API credentials.
	WithAuth(signer auth.Signer, apiKey *auth.APIKey) Client
	// WithBuilderConfig returns a new client instance configured for builder attribution.
	WithBuilderConfig(config *auth.BuilderConfig) Client
	// PromoteToBuilder switches the client into builder attribution mode.
	PromoteToBuilder(config *auth.BuilderConfig) Client
	// WithSignatureType sets the default signature type used for order signing and balance/rewards requests.
	WithSignatureType(sigType auth.SignatureType) Client
	// WithAuthNonce sets the default nonce used when creating/deriving API keys.
	WithAuthNonce(nonce int64) Client
	// WithFunder sets the default funder (maker) address used for orders.
	WithFunder(funder types.Address) Client
	// WithSaltGenerator sets the default salt generator used for new orders.
	WithSaltGenerator(gen SaltGenerator) Client
	// WithUseServerTime configures the client to synchronize with server time for request signing.
	WithUseServerTime(use bool) Client
	// WithGeoblockHost overrides the host used for checking geoblocking status.
	WithGeoblockHost(host string) Client
	// WithWS associates a WebSocket client with this REST client.
	WithWS(ws ws.Client) Client
	// WithHeartbeatInterval enables automatic heartbeat scheduling.
	WithHeartbeatInterval(interval time.Duration) Client
	// StopHeartbeats stops any active heartbeat loop.
	StopHeartbeats()

	// -- High-level Helpers --

	// CreateOrder builds, signs, and submits a new order to the exchange in one call.
	CreateOrder(ctx context.Context, order *clobtypes.Order) (clobtypes.OrderResponse, error)
	// CreateOrderWithOptions is like CreateOrder but allows specifying advanced order options.
	CreateOrderWithOptions(ctx context.Context, order *clobtypes.Order, opts *clobtypes.OrderOptions) (clobtypes.OrderResponse, error)
	// CreateOrderFromSignable submits an order that has already been prepared as a SignableOrder.
	CreateOrderFromSignable(ctx context.Context, order *clobtypes.SignableOrder) (clobtypes.OrderResponse, error)

	// -- System Status --

	// Health returns the current health status of the CLOB API.
	Health(ctx context.Context) (string, error)
	// Time retrieves the current server time from the exchange.
	Time(ctx context.Context) (clobtypes.TimeResponse, error)
	// Geoblock checks if the current IP address is restricted from accessing the exchange.
	Geoblock(ctx context.Context) (clobtypes.GeoblockResponse, error)

	// -- Market Data --

	// Markets retrieves a paginated list of available markets.
	Markets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)
	// MarketsAll automatically iterates through all pages to retrieve all available markets.
	MarketsAll(ctx context.Context, req *clobtypes.MarketsRequest) ([]clobtypes.Market, error)
	// Market retrieves detailed information for a single market by its ID.
	Market(ctx context.Context, id string) (clobtypes.MarketResponse, error)
	// SimplifiedMarkets retrieves a simplified view of available markets.
	SimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)
	// SamplingMarkets retrieves a sampled list of markets.
	SamplingMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)
	// SamplingSimplifiedMarkets retrieves a sampled and simplified list of markets.
	SamplingSimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)

	// -- Order Book & Pricing --

	// OrderBook retrieves the current L2 order book for a specific token.
	OrderBook(ctx context.Context, req *clobtypes.BookRequest) (clobtypes.OrderBookResponse, error)
	// OrderBooks retrieves multiple order books in a single batch request.
	OrderBooks(ctx context.Context, req *clobtypes.BooksRequest) (clobtypes.OrderBooksResponse, error)
	// Midpoint retrieves the current mid-price for a token.
	Midpoint(ctx context.Context, req *clobtypes.MidpointRequest) (clobtypes.MidpointResponse, error)
	// Midpoints retrieves multiple mid-prices in a batch request.
	Midpoints(ctx context.Context, req *clobtypes.MidpointsRequest) (clobtypes.MidpointsResponse, error)
	// Price retrieves the current price for a token on a specific side.
	Price(ctx context.Context, req *clobtypes.PriceRequest) (clobtypes.PriceResponse, error)
	// Prices retrieves multiple prices in a batch request.
	Prices(ctx context.Context, req *clobtypes.PricesRequest) (clobtypes.PricesResponse, error)
	// AllPrices retrieves current prices for all active tokens.
	AllPrices(ctx context.Context) (clobtypes.PricesResponse, error)
	// Spread retrieves the current bid-ask spread for a token.
	Spread(ctx context.Context, req *clobtypes.SpreadRequest) (clobtypes.SpreadResponse, error)
	// Spreads retrieves multiple spreads in a batch request.
	Spreads(ctx context.Context, req *clobtypes.SpreadsRequest) (clobtypes.SpreadsResponse, error)
	// LastTradePrice retrieves the price of the last executed trade for a token.
	LastTradePrice(ctx context.Context, req *clobtypes.LastTradePriceRequest) (clobtypes.LastTradePriceResponse, error)
	// LastTradesPrices retrieves last trade prices for multiple tokens in a batch.
	LastTradesPrices(ctx context.Context, req *clobtypes.LastTradesPricesRequest) (clobtypes.LastTradesPricesResponse, error)
	// TickSize retrieves the minimum price increment for a token.
	TickSize(ctx context.Context, req *clobtypes.TickSizeRequest) (clobtypes.TickSizeResponse, error)
	// NegRisk checks if a token belongs to a negative risk market.
	NegRisk(ctx context.Context, req *clobtypes.NegRiskRequest) (clobtypes.NegRiskResponse, error)
	// FeeRate retrieves the current fee rate applicable to a token.
	FeeRate(ctx context.Context, req *clobtypes.FeeRateRequest) (clobtypes.FeeRateResponse, error)
	// PricesHistory retrieves historical price points for a market (condition ID) or token.
	PricesHistory(ctx context.Context, req *clobtypes.PricesHistoryRequest) (clobtypes.PricesHistoryResponse, error)

	// -- Cache Management --

	// InvalidateCaches clears all internally cached market metadata (tick sizes, fee rates).
	InvalidateCaches()
	// SetTickSize manually populates the tick size cache for a token.
	SetTickSize(tokenID string, tickSize float64)
	// SetNegRisk manually populates the negative risk cache for a token.
	SetNegRisk(tokenID string, negRisk bool)
	// SetFeeRateBps manually populates the fee rate cache for a token.
	SetFeeRateBps(tokenID string, feeRateBps int64)

	// -- Order & Trade Management --

	// PostOrder submits a pre-signed order to the exchange.
	PostOrder(ctx context.Context, req *clobtypes.SignedOrder) (clobtypes.OrderResponse, error)
	// PostOrders submits multiple pre-signed orders in a single batch.
	PostOrders(ctx context.Context, req *clobtypes.SignedOrders) (clobtypes.PostOrdersResponse, error)
	// CancelOrder requests the cancellation of a single open order by its ID.
	CancelOrder(ctx context.Context, req *clobtypes.CancelOrderRequest) (clobtypes.CancelResponse, error)
	// CancelOrders requests the cancellation of multiple orders by their IDs.
	CancelOrders(ctx context.Context, req *clobtypes.CancelOrdersRequest) (clobtypes.CancelResponse, error)
	// CancelAll requests the cancellation of all open orders for the authenticated account.
	CancelAll(ctx context.Context) (clobtypes.CancelAllResponse, error)
	// CancelMarketOrders requests the cancellation of all orders in a specific market.
	CancelMarketOrders(ctx context.Context, req *clobtypes.CancelMarketOrdersRequest) (clobtypes.CancelMarketOrdersResponse, error)
	// Order retrieves the current status and details of a specific order.
	Order(ctx context.Context, id string) (clobtypes.OrderResponse, error)
	// Orders retrieves a paginated list of open orders for the authenticated account.
	Orders(ctx context.Context, req *clobtypes.OrdersRequest) (clobtypes.OrdersResponse, error)
	// Trades retrieves a paginated list of executed trades.
	Trades(ctx context.Context, req *clobtypes.TradesRequest) (clobtypes.TradesResponse, error)

	// OrdersAll automatically iterates through all pages to retrieve all open orders.
	OrdersAll(ctx context.Context, req *clobtypes.OrdersRequest) ([]clobtypes.OrderResponse, error)
	// TradesAll automatically iterates through all pages to retrieve all recent trades.
	TradesAll(ctx context.Context, req *clobtypes.TradesRequest) ([]clobtypes.Trade, error)
	// BuilderTradesAll automatically iterates through all pages to retrieve all trades attributed to a builder.
	BuilderTradesAll(ctx context.Context, req *clobtypes.BuilderTradesRequest) ([]clobtypes.Trade, error)

	// -- Scoring & Performance --

	// OrderScoring retrieves the liquidity scoring details for a specific order.
	OrderScoring(ctx context.Context, req *clobtypes.OrderScoringRequest) (clobtypes.OrderScoringResponse, error)
	// OrdersScoring retrieves scoring details for multiple orders in a batch.
	OrdersScoring(ctx context.Context, req *clobtypes.OrdersScoringRequest) (clobtypes.OrdersScoringResponse, error)

	// -- Account & Notifications --

	// BalanceAllowance retrieves the current balance and exchange allowance for a specific asset.
	BalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceRequest) (clobtypes.BalanceAllowanceResponse, error)
	// UpdateBalanceAllowance (Internal use) prepares a request to update the asset allowance.
	UpdateBalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceUpdateRequest) (clobtypes.BalanceAllowanceResponse, error)
	// Notifications retrieves recent account notifications.
	Notifications(ctx context.Context, req *clobtypes.NotificationsRequest) (clobtypes.NotificationsResponse, error)
	// DropNotifications acknowledges and clears a specific notification.
	DropNotifications(ctx context.Context, req *clobtypes.DropNotificationsRequest) (clobtypes.DropNotificationsResponse, error)

	// -- Rewards & Earnings --

	// UserEarnings retrieves the current pending rewards for the user.
	UserEarnings(ctx context.Context, req *clobtypes.UserEarningsRequest) (clobtypes.UserEarningsResponse, error)
	// UserTotalEarnings retrieves the lifetime cumulative earnings for the user.
	UserTotalEarnings(ctx context.Context, req *clobtypes.UserTotalEarningsRequest) (clobtypes.UserTotalEarningsResponse, error)
	// UserRewardPercentages retrieves the current reward rate multipliers for the user.
	UserRewardPercentages(ctx context.Context, req *clobtypes.UserRewardPercentagesRequest) (clobtypes.UserRewardPercentagesResponse, error)
	// RewardsMarketsCurrent retrieves the list of markets currently eligible for liquidity rewards.
	RewardsMarketsCurrent(ctx context.Context, req *clobtypes.RewardsMarketsRequest) (clobtypes.RewardsMarketsResponse, error)
	// RewardsMarkets retrieves historical reward details for a specific market.
	RewardsMarkets(ctx context.Context, req *clobtypes.RewardsMarketRequest) (clobtypes.RewardsMarketResponse, error)
	// UserRewardsByMarket retrieves user earnings alongside market rewards configuration.
	UserRewardsByMarket(ctx context.Context, req *clobtypes.UserRewardsByMarketRequest) (clobtypes.UserRewardsByMarketResponse, error)

	// -- API Key Management --

	// CreateAPIKey creates a new set of L2 API credentials using an L1 signature.
	CreateAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	// CreateAPIKeyWithNonce creates a new set of L2 API credentials with an explicit nonce.
	CreateAPIKeyWithNonce(ctx context.Context, nonce int64) (clobtypes.APIKeyResponse, error)
	// ListAPIKeys lists all active L2 API keys for the authenticated account.
	ListAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error)
	// DeleteAPIKey revokes a specific L2 API key.
	DeleteAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error)
	// DeriveAPIKey computes the deterministic L2 API key associated with the L1 wallet.
	DeriveAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	// DeriveAPIKeyWithNonce computes the deterministic L2 API key with an explicit nonce.
	DeriveAPIKeyWithNonce(ctx context.Context, nonce int64) (clobtypes.APIKeyResponse, error)
	// CreateOrDeriveAPIKey attempts to create a new API key, falling back to derive on failure.
	CreateOrDeriveAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	// CreateOrDeriveAPIKeyWithNonce attempts to create a new API key with an explicit nonce, falling back to derive on failure.
	CreateOrDeriveAPIKeyWithNonce(ctx context.Context, nonce int64) (clobtypes.APIKeyResponse, error)
	// ClosedOnlyStatus checks if the account is restricted to "close-only" trading.
	ClosedOnlyStatus(ctx context.Context) (clobtypes.ClosedOnlyResponse, error)

	// -- Read-only API Keys --

	// CreateReadonlyAPIKey creates a new API key with read-only permissions.
	CreateReadonlyAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	// ListReadonlyAPIKeys lists all active read-only keys.
	ListReadonlyAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error)
	// DeleteReadonlyAPIKey revokes a read-only API key.
	DeleteReadonlyAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error)
	// ValidateReadonlyAPIKey verifies if a read-only key is valid for a given address.
	ValidateReadonlyAPIKey(ctx context.Context, req *clobtypes.ValidateReadonlyAPIKeyRequest) (clobtypes.ValidateReadonlyAPIKeyResponse, error)

	// -- Builder API Keys --

	// CreateBuilderAPIKey creates a new API key for builder attribution.
	CreateBuilderAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	// ListBuilderAPIKeys lists all active builder keys.
	ListBuilderAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error)
	// RevokeBuilderAPIKey revokes a builder API key.
	RevokeBuilderAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error)
	// BuilderTrades retrieves trades attributed to the authenticated builder.
	BuilderTrades(ctx context.Context, req *clobtypes.BuilderTradesRequest) (clobtypes.BuilderTradesResponse, error)

	// MarketTradesEvents retrieves a stream of recent trade events for a market.
	MarketTradesEvents(ctx context.Context, id string) (clobtypes.MarketTradesEventsResponse, error)

	// -- Sub-Client Accessors --

	// RFQ returns the Request For Quote sub-client.
	RFQ() rfq.Client
	// WS returns the WebSocket streaming sub-client.
	WS() ws.Client
	// Heartbeat returns the L2 heartbeat sub-client.
	Heartbeat() heartbeat.Client
}
