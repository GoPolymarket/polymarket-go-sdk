package rtds

import (
	"context"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
)

// Client defines the RTDS WebSocket interface.
type Client interface {
	// Authenticate sets default CLOB credentials for authenticated comment streams.
	Authenticate(apiKey *auth.APIKey) Client
	// Deauthenticate clears any stored credentials for authenticated streams.
	Deauthenticate() Client

	SubscribeCryptoPricesStream(ctx context.Context, symbols []string) (*Stream[CryptoPriceEvent], error)
	SubscribeChainlinkPricesStream(ctx context.Context, feeds []string) (*Stream[ChainlinkPriceEvent], error)
	SubscribeCommentsStream(ctx context.Context, req *CommentFilter) (*Stream[CommentEvent], error)
	SubscribeOrdersMatchedStream(ctx context.Context) (*Stream[OrdersMatchedEvent], error)
	SubscribeRawStream(ctx context.Context, sub *Subscription) (*Stream[RtdsMessage], error)
	SubscribeCryptoPrices(ctx context.Context, symbols []string) (<-chan CryptoPriceEvent, error)
	SubscribeChainlinkPrices(ctx context.Context, feeds []string) (<-chan ChainlinkPriceEvent, error)
	SubscribeComments(ctx context.Context, req *CommentFilter) (<-chan CommentEvent, error)
	SubscribeOrdersMatched(ctx context.Context) (<-chan OrdersMatchedEvent, error)
	SubscribeRaw(ctx context.Context, sub *Subscription) (<-chan RtdsMessage, error)
	UnsubscribeCryptoPrices(ctx context.Context) error
	UnsubscribeChainlinkPrices(ctx context.Context) error
	UnsubscribeComments(ctx context.Context, commentType *CommentType) error
	UnsubscribeOrdersMatched(ctx context.Context) error
	UnsubscribeRaw(ctx context.Context, sub *Subscription) error
	ConnectionState() ConnectionState
	ConnectionStateStream(ctx context.Context) (*Stream[ConnectionStateEvent], error)
	SubscriptionCount() int
	Close() error
}
