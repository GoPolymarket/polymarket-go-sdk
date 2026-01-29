package gamma

import (
	"context"
)

// Client defines the Gamma API interface.
type Client interface {
	Status(ctx context.Context) (StatusResponse, error)
	Teams(ctx context.Context, req *TeamsRequest) ([]Team, error)
	Sports(ctx context.Context) ([]SportsMetadata, error)
	SportsMarketTypes(ctx context.Context) (SportsMarketTypesResponse, error)
	Tags(ctx context.Context, req *TagsRequest) ([]Tag, error)
	TagByID(ctx context.Context, req *TagByIDRequest) (*Tag, error)
	TagBySlug(ctx context.Context, req *TagBySlugRequest) (*Tag, error)
	RelatedTagsByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]RelatedTag, error)
	RelatedTagsBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]RelatedTag, error)
	TagsRelatedToTagByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]Tag, error)
	TagsRelatedToTagBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]Tag, error)
	
	Events(ctx context.Context, req *EventsRequest) ([]Event, error)
	EventsAll(ctx context.Context, req *EventsRequest) ([]Event, error)
	EventByID(ctx context.Context, req *EventByIDRequest) (*Event, error)
	EventBySlug(ctx context.Context, req *EventBySlugRequest) (*Event, error)
	EventTags(ctx context.Context, req *EventTagsRequest) ([]Tag, error)
	
	Markets(ctx context.Context, req *MarketsRequest) ([]Market, error)
	MarketsAll(ctx context.Context, req *MarketsRequest) ([]Market, error)
	MarketByID(ctx context.Context, req *MarketByIDRequest) (*Market, error)
	MarketBySlug(ctx context.Context, req *MarketBySlugRequest) (*Market, error)
	MarketTags(ctx context.Context, req *MarketTagsRequest) ([]Tag, error)
	
	Series(ctx context.Context, req *SeriesRequest) ([]Series, error)
	SeriesByID(ctx context.Context, req *SeriesByIDRequest) (*Series, error)
	Comments(ctx context.Context, req *CommentsRequest) ([]Comment, error)
	CommentByID(ctx context.Context, req *CommentByIDRequest) ([]Comment, error)
	CommentsByUserAddress(ctx context.Context, req *CommentsByUserAddressRequest) ([]Comment, error)
	PublicProfile(ctx context.Context, req *PublicProfileRequest) (*PublicProfile, error)
	PublicSearch(ctx context.Context, req *PublicSearchRequest) (SearchResults, error)

	// Backwards compatible aliases.
	GetMarkets(ctx context.Context, req *MarketsRequest) ([]Market, error)
	GetMarket(ctx context.Context, id string) (*Market, error)
	GetEvents(ctx context.Context, req *MarketsRequest) ([]Event, error) // Reuse MarketsRequest for simplicity or define EventsRequest
	GetEvent(ctx context.Context, id string) (*Event, error)
}