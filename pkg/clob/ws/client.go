// Package ws provides a high-level WebSocket client for Polymarket.
// It manages connections to both market data and user-specific event streams,
// handling automatic reconnection, heartbeats, and event dispatching via channels.
package ws

import (
	"context"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
)

// Client defines the interface for interacting with Polymarket's WebSocket services.
// It provides a stream-based API for real-time market data and private account updates.
type Client interface {
	// -- Connection Management --

	// Authenticate sets API credentials for private user streams.
	Authenticate(signer auth.Signer, apiKey *auth.APIKey) Client
	// Deauthenticate clears API credentials for private user streams.
	Deauthenticate() Client

	// ConnectionState returns the current status of a specific WebSocket channel.
	ConnectionState(channel Channel) ConnectionState
	// ConnectionStateStream returns a stream of connection state transition events.
	ConnectionStateStream(ctx context.Context) (*Stream[ConnectionStateEvent], error)
	// Close gracefully shuts down all active WebSocket connections and closes all event channels.
	Close() error

	// -- Market Data Streams (Public) --

	// SubscribeOrderbook subscribes to L2 order book snapshots and updates for specific assets.
	SubscribeOrderbook(ctx context.Context, assetIDs []string) (<-chan OrderbookEvent, error)
	// SubscribePrices subscribes to real-time price change events for specific assets.
	SubscribePrices(ctx context.Context, assetIDs []string) (<-chan PriceChangeEvent, error)
	// SubscribeMidpoints subscribes to mid-price update events for specific assets.
	SubscribeMidpoints(ctx context.Context, assetIDs []string) (<-chan MidpointEvent, error)
	// SubscribeLastTradePrices subscribes to the price of the latest executed trades for specific assets.
	SubscribeLastTradePrices(ctx context.Context, assetIDs []string) (<-chan LastTradePriceEvent, error)
	// SubscribeTickSizeChanges subscribes to minimum price increment changes for specific assets.
	SubscribeTickSizeChanges(ctx context.Context, assetIDs []string) (<-chan TickSizeChangeEvent, error)
	// SubscribeBestBidAsk subscribes to top-of-book (BBO) events for specific assets.
	SubscribeBestBidAsk(ctx context.Context, assetIDs []string) (<-chan BestBidAskEvent, error)
	// SubscribeNewMarkets subscribes to events triggered when new markets are created.
	SubscribeNewMarkets(ctx context.Context, assetIDs []string) (<-chan NewMarketEvent, error)
	// SubscribeMarketResolutions subscribes to events triggered when markets are resolved.
	SubscribeMarketResolutions(ctx context.Context, assetIDs []string) (<-chan MarketResolvedEvent, error)

	// -- User Activity Streams (Private) --

	// SubscribeUserOrders subscribes to status updates for orders belonging to the authenticated account.
	// Requires an API key to be configured on the client.
	SubscribeUserOrders(ctx context.Context, markets []string) (<-chan OrderEvent, error)
	// SubscribeUserTrades subscribes to trade execution events for the authenticated account.
	// Requires an API key to be configured on the client.
	SubscribeUserTrades(ctx context.Context, markets []string) (<-chan TradeEvent, error)

	// -- Advanced Stream Control --

	// SubscribeOrderbookStream is like SubscribeOrderbook but returns a managed Stream object.
	SubscribeOrderbookStream(ctx context.Context, assetIDs []string) (*Stream[OrderbookEvent], error)
	// SubscribePricesStream is like SubscribePrices but returns a managed Stream object.
	SubscribePricesStream(ctx context.Context, assetIDs []string) (*Stream[PriceChangeEvent], error)
	// SubscribeMidpointsStream is like SubscribeMidpoints but returns a managed Stream object.
	SubscribeMidpointsStream(ctx context.Context, assetIDs []string) (*Stream[MidpointEvent], error)
	// SubscribeLastTradePricesStream is like SubscribeLastTradePrices but returns a managed Stream object.
	SubscribeLastTradePricesStream(ctx context.Context, assetIDs []string) (*Stream[LastTradePriceEvent], error)
	// SubscribeTickSizeChangesStream is like SubscribeTickSizeChanges but returns a managed Stream object.
	SubscribeTickSizeChangesStream(ctx context.Context, assetIDs []string) (*Stream[TickSizeChangeEvent], error)
	// SubscribeBestBidAskStream is like SubscribeBestBidAsk but returns a managed Stream object.
	SubscribeBestBidAskStream(ctx context.Context, assetIDs []string) (*Stream[BestBidAskEvent], error)
	// SubscribeNewMarketsStream is like SubscribeNewMarkets but returns a managed Stream object.
	SubscribeNewMarketsStream(ctx context.Context, assetIDs []string) (*Stream[NewMarketEvent], error)
	// SubscribeMarketResolutionsStream is like SubscribeMarketResolutions but returns a managed Stream object.
	SubscribeMarketResolutionsStream(ctx context.Context, assetIDs []string) (*Stream[MarketResolvedEvent], error)
	// SubscribeUserOrdersStream is like SubscribeUserOrders but returns a managed Stream object.
	SubscribeUserOrdersStream(ctx context.Context, markets []string) (*Stream[OrderEvent], error)
	// SubscribeUserTradesStream is like SubscribeUserTrades but returns a managed Stream object.
	SubscribeUserTradesStream(ctx context.Context, markets []string) (*Stream[TradeEvent], error)

	// -- Low-level Subscription Control --

	// Subscribe sends a raw subscription request to the WebSocket server.
	Subscribe(ctx context.Context, req *SubscriptionRequest) error
	// Unsubscribe sends a raw unsubscription request to the WebSocket server.
	Unsubscribe(ctx context.Context, req *SubscriptionRequest) error
	// UnsubscribeMarketAssets unsubscribes from all events related to specific assets on the market channel.
	UnsubscribeMarketAssets(ctx context.Context, assetIDs []string) error
	// UnsubscribeUserMarkets unsubscribes from all account events related to specific markets.
	UnsubscribeUserMarkets(ctx context.Context, markets []string) error
}
