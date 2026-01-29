// Package gamma provides the client for interacting with the Polymarket Gamma API.
// Gamma is the metadata and discovery service for Polymarket, used to find markets,
// events, teams, and search for active prediction topics.
package gamma

import (
	"context"
)

// Client defines the interface for the Polymarket Gamma metadata service.
// It is primarily a read-only API used for discovery and metadata retrieval.
type Client interface {
	// -- System Status --

	// Status returns the current operational status of the Gamma service.
	Status(ctx context.Context) (StatusResponse, error)

	// -- Metadata Discovery --

	// Teams retrieves a list of teams associated with sports markets.
	Teams(ctx context.Context, req *TeamsRequest) ([]Team, error)
	// Sports retrieves metadata about supported sport categories.
	Sports(ctx context.Context) ([]SportsMetadata, error)
	// SportsMarketTypes lists the types of prediction markets available for sports.
	SportsMarketTypes(ctx context.Context) (SportsMarketTypesResponse, error)
	
	// -- Tags --

	// Tags retrieves a list of market tags (categories).
	Tags(ctx context.Context, req *TagsRequest) ([]Tag, error)
	// TagByID retrieves a specific tag by its unique ID.
	TagByID(ctx context.Context, req *TagByIDRequest) (*Tag, error)
	// TagBySlug retrieves a specific tag by its URL slug.
	TagBySlug(ctx context.Context, req *TagBySlugRequest) (*Tag, error)
	// RelatedTagsByID finds tags related to a given tag ID.
	RelatedTagsByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]RelatedTag, error)
	// RelatedTagsBySlug finds tags related to a given tag slug.
	RelatedTagsBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]RelatedTag, error)
	// TagsRelatedToTagByID retrieves full tag objects related to a specific tag ID.
	TagsRelatedToTagByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]Tag, error)
	// TagsRelatedToTagBySlug retrieves full tag objects related to a specific tag slug.
	TagsRelatedToTagBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]Tag, error)
	
	// -- Events --

	// Events retrieves a list of prediction events (groups of markets).
	Events(ctx context.Context, req *EventsRequest) ([]Event, error)
	// EventsAll automatically iterates through all pages to retrieve all available events.
	EventsAll(ctx context.Context, req *EventsRequest) ([]Event, error)
	// EventByID retrieves a specific event by its ID.
	EventByID(ctx context.Context, req *EventByIDRequest) (*Event, error)
	// EventBySlug retrieves a specific event by its URL slug.
	EventBySlug(ctx context.Context, req *EventBySlugRequest) (*Event, error)
	// EventTags lists tags associated with a specific event.
	EventTags(ctx context.Context, req *EventTagsRequest) ([]Tag, error)
	
	// -- Markets --

	// Markets retrieves a list of specific prediction markets.
	Markets(ctx context.Context, req *MarketsRequest) ([]Market, error)
	// MarketsAll automatically iterates through all pages to retrieve all available markets.
	MarketsAll(ctx context.Context, req *MarketsRequest) ([]Market, error)
	// MarketByID retrieves a specific market by its ID.
	MarketByID(ctx context.Context, req *MarketByIDRequest) (*Market, error)
	// MarketBySlug retrieves a specific market by its URL slug.
	MarketBySlug(ctx context.Context, req *MarketBySlugRequest) (*Market, error)
	// MarketTags lists tags associated with a specific market.
	MarketTags(ctx context.Context, req *MarketTagsRequest) ([]Tag, error)
	
	// -- Series & Collections --

	// Series retrieves a list of market series (related groups of events).
	Series(ctx context.Context, req *SeriesRequest) ([]Series, error)
	// SeriesByID retrieves a specific series by its ID.
	SeriesByID(ctx context.Context, req *SeriesByIDRequest) (*Series, error)

	// -- Social & Search --

	// Comments retrieves comments for a specific entity (market or event).
	Comments(ctx context.Context, req *CommentsRequest) ([]Comment, error)
	// CommentByID retrieves a specific comment by its ID.
	CommentByID(ctx context.Context, req *CommentByIDRequest) ([]Comment, error)
	// CommentsByUserAddress retrieves all comments made by a specific wallet address.
	CommentsByUserAddress(ctx context.Context, req *CommentsByUserAddressRequest) ([]Comment, error)
	// PublicProfile retrieves the public user profile associated with a wallet address.
	PublicProfile(ctx context.Context, req *PublicProfileRequest) (*PublicProfile, error)
	// PublicSearch performs a global search across markets, events, and tags.
	PublicSearch(ctx context.Context, req *PublicSearchRequest) (SearchResults, error)

	// -- Legacy / Compatibility Aliases --

	// GetMarkets is a legacy alias for Markets.
	GetMarkets(ctx context.Context, req *MarketsRequest) ([]Market, error)
	// GetMarket is a legacy alias for MarketByID.
	GetMarket(ctx context.Context, id string) (*Market, error)
	// GetEvents is a legacy alias for Events.
	GetEvents(ctx context.Context, req *MarketsRequest) ([]Event, error)
	// GetEvent is a legacy alias for EventByID.
	GetEvent(ctx context.Context, id string) (*Event, error)
}