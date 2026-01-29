package clob

import (
	"context"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/heartbeat"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/rfq"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

// Client defines the CLOB REST interface.
type Client interface {
	// Auth
	WithAuth(signer auth.Signer, apiKey *auth.APIKey) Client
	WithBuilderConfig(config *auth.BuilderConfig) Client
	WithUseServerTime(use bool) Client
	WithGeoblockHost(host string) Client
	WithWS(ws ws.Client) Client

	// High-level Helpers
	CreateOrder(ctx context.Context, order *clobtypes.Order) (clobtypes.OrderResponse, error)

	Health(ctx context.Context) (string, error)
	Time(ctx context.Context) (clobtypes.TimeResponse, error)
	Geoblock(ctx context.Context) (clobtypes.GeoblockResponse, error)

	Markets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)
	MarketsAll(ctx context.Context, req *clobtypes.MarketsRequest) ([]clobtypes.Market, error)
	Market(ctx context.Context, id string) (clobtypes.MarketResponse, error)
	SimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)
	SamplingMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)
	SamplingSimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error)

	OrderBook(ctx context.Context, req *clobtypes.BookRequest) (clobtypes.OrderBookResponse, error)
	OrderBooks(ctx context.Context, req *clobtypes.BooksRequest) (clobtypes.OrderBooksResponse, error)

	Midpoint(ctx context.Context, req *clobtypes.MidpointRequest) (clobtypes.MidpointResponse, error)
	Midpoints(ctx context.Context, req *clobtypes.MidpointsRequest) (clobtypes.MidpointsResponse, error)
	Price(ctx context.Context, req *clobtypes.PriceRequest) (clobtypes.PriceResponse, error)
	Prices(ctx context.Context, req *clobtypes.PricesRequest) (clobtypes.PricesResponse, error)
	AllPrices(ctx context.Context) (clobtypes.PricesResponse, error)
	Spread(ctx context.Context, req *clobtypes.SpreadRequest) (clobtypes.SpreadResponse, error)
	Spreads(ctx context.Context, req *clobtypes.SpreadsRequest) (clobtypes.SpreadsResponse, error)
	LastTradePrice(ctx context.Context, req *clobtypes.LastTradePriceRequest) (clobtypes.LastTradePriceResponse, error)
	LastTradesPrices(ctx context.Context, req *clobtypes.LastTradesPricesRequest) (clobtypes.LastTradesPricesResponse, error)
	TickSize(ctx context.Context, req *clobtypes.TickSizeRequest) (clobtypes.TickSizeResponse, error)
	NegRisk(ctx context.Context, req *clobtypes.NegRiskRequest) (clobtypes.NegRiskResponse, error)
	FeeRate(ctx context.Context, req *clobtypes.FeeRateRequest) (clobtypes.FeeRateResponse, error)
	PricesHistory(ctx context.Context, req *clobtypes.PricesHistoryRequest) (clobtypes.PricesHistoryResponse, error)

	// Cache helpers
	InvalidateCaches()
	SetTickSize(tokenID, tickSize string)
	SetNegRisk(tokenID string, negRisk bool)
	SetFeeRateBps(tokenID string, feeRateBps int64)

	PostOrder(ctx context.Context, req *clobtypes.SignedOrder) (clobtypes.OrderResponse, error)
	PostOrders(ctx context.Context, req *clobtypes.SignedOrders) (clobtypes.PostOrdersResponse, error)
	CancelOrder(ctx context.Context, req *clobtypes.CancelOrderRequest) (clobtypes.CancelResponse, error)
	CancelOrders(ctx context.Context, req *clobtypes.CancelOrdersRequest) (clobtypes.CancelResponse, error)
	CancelAll(ctx context.Context) (clobtypes.CancelAllResponse, error)
	CancelMarketOrders(ctx context.Context, req *clobtypes.CancelMarketOrdersRequest) (clobtypes.CancelMarketOrdersResponse, error)
	Order(ctx context.Context, id string) (clobtypes.OrderResponse, error)
	Orders(ctx context.Context, req *clobtypes.OrdersRequest) (clobtypes.OrdersResponse, error)
	Trades(ctx context.Context, req *clobtypes.TradesRequest) (clobtypes.TradesResponse, error)

	CreateOrderWithOptions(ctx context.Context, order *clobtypes.Order, opts *clobtypes.OrderOptions) (clobtypes.OrderResponse, error)
	CreateOrderFromSignable(ctx context.Context, order *clobtypes.SignableOrder) (clobtypes.OrderResponse, error)

	OrdersAll(ctx context.Context, req *clobtypes.OrdersRequest) ([]clobtypes.OrderResponse, error)
	TradesAll(ctx context.Context, req *clobtypes.TradesRequest) ([]clobtypes.Trade, error)
	BuilderTradesAll(ctx context.Context, req *clobtypes.BuilderTradesRequest) ([]clobtypes.Trade, error)

	OrderScoring(ctx context.Context, req *clobtypes.OrderScoringRequest) (clobtypes.OrderScoringResponse, error)
	OrdersScoring(ctx context.Context, req *clobtypes.OrdersScoringRequest) (clobtypes.OrdersScoringResponse, error)

	BalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceRequest) (clobtypes.BalanceAllowanceResponse, error)
	UpdateBalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceUpdateRequest) (clobtypes.BalanceAllowanceResponse, error)

	Notifications(ctx context.Context, req *clobtypes.NotificationsRequest) (clobtypes.NotificationsResponse, error)
	DropNotifications(ctx context.Context, req *clobtypes.DropNotificationsRequest) (clobtypes.DropNotificationsResponse, error)

	UserEarnings(ctx context.Context, req *clobtypes.UserEarningsRequest) (clobtypes.UserEarningsResponse, error)
	UserTotalEarnings(ctx context.Context, req *clobtypes.UserTotalEarningsRequest) (clobtypes.UserTotalEarningsResponse, error)
	UserRewardPercentages(ctx context.Context, req *clobtypes.UserRewardPercentagesRequest) (clobtypes.UserRewardPercentagesResponse, error)
	RewardsMarketsCurrent(ctx context.Context) (clobtypes.RewardsMarketsResponse, error)
	RewardsMarkets(ctx context.Context, id string) (clobtypes.RewardsMarketResponse, error)
	UserRewardsByMarket(ctx context.Context, req *clobtypes.UserRewardsByMarketRequest) (clobtypes.UserRewardsByMarketResponse, error)

	MarketTradesEvents(ctx context.Context, id string) (clobtypes.MarketTradesEventsResponse, error)

	CreateAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	ListAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error)
	DeleteAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error)
	DeriveAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	ClosedOnlyStatus(ctx context.Context) (clobtypes.ClosedOnlyResponse, error)

	CreateReadonlyAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	ListReadonlyAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error)
	DeleteReadonlyAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error)
	ValidateReadonlyAPIKey(ctx context.Context, req *clobtypes.ValidateReadonlyAPIKeyRequest) (clobtypes.ValidateReadonlyAPIKeyResponse, error)

	CreateBuilderAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error)
	ListBuilderAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error)
	RevokeBuilderAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error)
	BuilderTrades(ctx context.Context, req *clobtypes.BuilderTradesRequest) (clobtypes.BuilderTradesResponse, error)

	RFQ() rfq.Client
	WS() ws.Client
	Heartbeat() heartbeat.Client
}
