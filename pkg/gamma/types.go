package gamma

import "encoding/json"

// Request parameters
type MarketsRequest struct {
	Limit               *int     `json:"limit,omitempty"`
	Offset              *int     `json:"offset,omitempty"`
	Order               string   `json:"order,omitempty"` // "volume", "created", "liquidity"
	Ascending           *bool    `json:"ascending,omitempty"`
	Slug                string   `json:"slug,omitempty"`
	Slugs               []string `json:"slugs,omitempty"`
	IDs                 []string `json:"ids,omitempty"`
	ClobTokenIDs        []string `json:"clob_token_ids,omitempty"`
	ConditionIDs        []string `json:"condition_ids,omitempty"`
	MarketMakerAddress  []string `json:"market_maker_address,omitempty"`
	Active              *bool    `json:"active,omitempty"`
	Closed              *bool    `json:"closed,omitempty"`
	TagID               string   `json:"tag_id,omitempty"`
	TagSlug             string   `json:"tag_slug,omitempty"`
	RelatedTags         *bool    `json:"related_tags,omitempty"`
	Cyom                *bool    `json:"cyom,omitempty"`
	UmaResolutionStatus string   `json:"uma_resolution_status,omitempty"`
	GameID              string   `json:"game_id,omitempty"`
	SportsMarketTypes   []string `json:"sports_market_types,omitempty"`
	VolumeMin           *string  `json:"volume_min,omitempty"` // string to support large numbers
	VolumeMax           *string  `json:"volume_max,omitempty"`
	LiquidityMin        *string  `json:"liquidity_min,omitempty"`
	LiquidityMax        *string  `json:"liquidity_max,omitempty"`
	LiquidityNumMin     *string  `json:"liquidity_num_min,omitempty"`
	LiquidityNumMax     *string  `json:"liquidity_num_max,omitempty"`
	VolumeNumMin        *string  `json:"volume_num_min,omitempty"`
	VolumeNumMax        *string  `json:"volume_num_max,omitempty"`
	StartDateMin        string   `json:"start_date_min,omitempty"` // ISO string
	StartDateMax        string   `json:"start_date_max,omitempty"`
	EndDateMin          string   `json:"end_date_min,omitempty"`
	EndDateMax          string   `json:"end_date_max,omitempty"`
	RewardsMinSize      *string  `json:"rewards_min_size,omitempty"`
	RewardsMaxSize      *string  `json:"rewards_max_size,omitempty"`
}

type TeamsRequest struct {
	Limit        *int     `json:"limit,omitempty"`
	Offset       *int     `json:"offset,omitempty"`
	Order        string   `json:"order,omitempty"`
	Ascending    *bool    `json:"ascending,omitempty"`
	League       []string `json:"league,omitempty"`
	Name         []string `json:"name,omitempty"`
	Abbreviation []string `json:"abbreviation,omitempty"`
}

type TagsRequest struct {
	Limit           *int   `json:"limit,omitempty"`
	Offset          *int   `json:"offset,omitempty"`
	Order           string `json:"order,omitempty"`
	Ascending       *bool  `json:"ascending,omitempty"`
	IncludeTemplate *bool  `json:"include_template,omitempty"`
	IsCarousel      *bool  `json:"is_carousel,omitempty"`
}

type TagByIDRequest struct {
	ID              string `json:"-"`
	IncludeTemplate *bool  `json:"include_template,omitempty"`
}

type TagBySlugRequest struct {
	Slug            string `json:"-"`
	IncludeTemplate *bool  `json:"include_template,omitempty"`
}

type RelatedTagsByIDRequest struct {
	ID        string `json:"-"`
	OmitEmpty *bool  `json:"omit_empty,omitempty"`
	Status    string `json:"status,omitempty"`
}

type RelatedTagsBySlugRequest struct {
	Slug      string `json:"-"`
	OmitEmpty *bool  `json:"omit_empty,omitempty"`
	Status    string `json:"status,omitempty"`
}

type EventsRequest struct {
	Limit           *int     `json:"limit,omitempty"`
	Offset          *int     `json:"offset,omitempty"`
	Order           []string `json:"order,omitempty"`
	Ascending       *bool    `json:"ascending,omitempty"`
	IDs             []string `json:"id,omitempty"`
	TagID           string   `json:"tag_id,omitempty"`
	ExcludeTagID    []string `json:"exclude_tag_id,omitempty"`
	Slugs           []string `json:"slug,omitempty"`
	TagSlug         string   `json:"tag_slug,omitempty"`
	RelatedTags     *bool    `json:"related_tags,omitempty"`
	Active          *bool    `json:"active,omitempty"`
	Archived        *bool    `json:"archived,omitempty"`
	Featured        *bool    `json:"featured,omitempty"`
	Cyom            *bool    `json:"cyom,omitempty"`
	IncludeChat     *bool    `json:"include_chat,omitempty"`
	IncludeTemplate *bool    `json:"include_template,omitempty"`
	Recurrence      string   `json:"recurrence,omitempty"`
	Closed          *bool    `json:"closed,omitempty"`
	LiquidityMin    *string  `json:"liquidity_min,omitempty"`
	LiquidityMax    *string  `json:"liquidity_max,omitempty"`
	VolumeMin       *string  `json:"volume_min,omitempty"`
	VolumeMax       *string  `json:"volume_max,omitempty"`
	StartDateMin    string   `json:"start_date_min,omitempty"`
	StartDateMax    string   `json:"start_date_max,omitempty"`
	EndDateMin      string   `json:"end_date_min,omitempty"`
	EndDateMax      string   `json:"end_date_max,omitempty"`
}

type EventByIDRequest struct {
	ID              string `json:"-"`
	IncludeChat     *bool  `json:"include_chat,omitempty"`
	IncludeTemplate *bool  `json:"include_template,omitempty"`
}

type EventBySlugRequest struct {
	Slug            string `json:"-"`
	IncludeChat     *bool  `json:"include_chat,omitempty"`
	IncludeTemplate *bool  `json:"include_template,omitempty"`
}

type EventTagsRequest struct {
	ID string `json:"-"`
}

type MarketByIDRequest struct {
	ID         string `json:"-"`
	IncludeTag *bool  `json:"include_tag,omitempty"`
}

type MarketBySlugRequest struct {
	Slug       string `json:"-"`
	IncludeTag *bool  `json:"include_tag,omitempty"`
}

type MarketTagsRequest struct {
	ID string `json:"-"`
}

type SeriesRequest struct {
	Limit            *int     `json:"limit,omitempty"`
	Offset           *int     `json:"offset,omitempty"`
	Order            string   `json:"order,omitempty"`
	Ascending        *bool    `json:"ascending,omitempty"`
	Slugs            []string `json:"slug,omitempty"`
	CategoriesIDs    []string `json:"categories_ids,omitempty"`
	CategoriesLabels []string `json:"categories_labels,omitempty"`
	Closed           *bool    `json:"closed,omitempty"`
	IncludeChat      *bool    `json:"include_chat,omitempty"`
	Recurrence       string   `json:"recurrence,omitempty"`
}

type SeriesByIDRequest struct {
	ID          string `json:"-"`
	IncludeChat *bool  `json:"include_chat,omitempty"`
}

type CommentsRequest struct {
	ParentEntityType string `json:"parent_entity_type,omitempty"`
	ParentEntityID   string `json:"parent_entity_id,omitempty"`
	Limit            *int   `json:"limit,omitempty"`
	Offset           *int   `json:"offset,omitempty"`
	Order            string `json:"order,omitempty"`
	Ascending        *bool  `json:"ascending,omitempty"`
	GetPositions     *bool  `json:"get_positions,omitempty"`
	HoldersOnly      *bool  `json:"holders_only,omitempty"`
}

type CommentByIDRequest struct {
	ID           string `json:"-"`
	GetPositions *bool  `json:"get_positions,omitempty"`
}

type CommentsByUserAddressRequest struct {
	UserAddress string `json:"-"`
	Limit       *int   `json:"limit,omitempty"`
	Offset      *int   `json:"offset,omitempty"`
	Order       string `json:"order,omitempty"`
	Ascending   *bool  `json:"ascending,omitempty"`
}

type PublicProfileRequest struct {
	Address string `json:"address"`
}

type PublicSearchRequest struct {
	Query             string   `json:"q"`
	Cache             *bool    `json:"cache,omitempty"`
	EventsStatus      string   `json:"events_status,omitempty"`
	LimitPerType      *int     `json:"limit_per_type,omitempty"`
	Page              *int     `json:"page,omitempty"`
	EventsTag         []string `json:"events_tag,omitempty"`
	KeepClosedMarkets *int     `json:"keep_closed_markets,omitempty"`
	Sort              string   `json:"sort,omitempty"`
	Ascending         *bool    `json:"ascending,omitempty"`
	SearchTags        *bool    `json:"search_tags,omitempty"`
	SearchProfiles    *bool    `json:"search_profiles,omitempty"`
	Recurrence        string   `json:"recurrence,omitempty"`
	ExcludeTagID      []string `json:"exclude_tag_id,omitempty"`
	Optimized         *bool    `json:"optimized,omitempty"`
}

type Market struct {
	ID                 string  `json:"id"`
	Question           string  `json:"question"`
	ConditionID        string  `json:"conditionId"`
	Slug               string  `json:"slug"`
	ResolutionSource   string  `json:"resolutionSource"`
	EndDate            string  `json:"endDate"`
	Liquidity          string  `json:"liquidity"`
	StartDate          string  `json:"startDate"`
	Volume             string  `json:"volume"`
	Active             bool    `json:"active"`
	Closed             bool    `json:"closed"`
	MarketMakerAddress string  `json:"marketMakerAddress"`
	Tags               []Tag   `json:"tags"`
	Tokens             []Token `json:"tokens"`
	ClobTokenIds       string  `json:"clobTokenIds"`     // JSON string of token IDs
	Outcomes           string  `json:"outcomes"`          // JSON string of outcome labels
	OutcomePrices      string  `json:"outcomePrices"`     // JSON string of outcome prices
	Rewards            Rewards `json:"rewards"`
}

// ParsedTokens builds a Token slice by combining ClobTokenIds and Outcomes.
// Returns nil if the fields cannot be parsed.
func (m *Market) ParsedTokens() []Token {
	if len(m.Tokens) > 0 {
		return m.Tokens
	}
	var ids []string
	if err := json.Unmarshal([]byte(m.ClobTokenIds), &ids); err != nil {
		return nil
	}
	var outcomes []string
	_ = json.Unmarshal([]byte(m.Outcomes), &outcomes)

	tokens := make([]Token, len(ids))
	for i, id := range ids {
		tokens[i].TokenID = id
		if i < len(outcomes) {
			tokens[i].Outcome = outcomes[i]
		}
	}
	return tokens
}

type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

type Token struct {
	TokenID string  `json:"tokenId"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
	Winner  bool    `json:"winner"`
}

type Rewards struct {
	MinIncentive string `json:"minIncentive"`
	MaxIncentive string `json:"maxIncentive"`
}

type Event struct {
	ID           string   `json:"id"`
	Ticker       string   `json:"ticker"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	StartDate    string   `json:"startDate"`
	CreationDate string   `json:"creationDate"`
	EndDate      string   `json:"endDate"`
	Image        string   `json:"image"`
	Icon         string   `json:"icon"`
	Active       bool     `json:"active"`
	Closed       bool     `json:"closed"`
	Archived     bool     `json:"archived"`
	New          bool     `json:"new"`
	Featured     bool     `json:"featured"`
	Restricted   bool     `json:"restricted"`
	Liquidity    string   `json:"liquidity"`
	Volume       string   `json:"volume"`
	Markets      []Market `json:"markets"`
}

type Team struct {
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	League       string `json:"league,omitempty"`
	Record       string `json:"record,omitempty"`
	Logo         string `json:"logo,omitempty"`
	Abbreviation string `json:"abbreviation,omitempty"`
	Alias        string `json:"alias,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
	Color        string `json:"color,omitempty"`
	ProviderID   *int   `json:"providerId,omitempty"`
}

type SportsMetadata struct {
	ID         *int     `json:"id,omitempty"`
	Sport      string   `json:"sport"`
	Image      string   `json:"image"`
	Resolution string   `json:"resolution"`
	Ordering   string   `json:"ordering"`
	Tags       []string `json:"tags,omitempty"`
	Series     string   `json:"series,omitempty"`
	CreatedAt  string   `json:"createdAt,omitempty"`
}

type SportsMarketTypesResponse struct {
	MarketTypes []string `json:"marketTypes"`
}

type RelatedTag struct {
	ID           string `json:"id"`
	TagID        string `json:"tagID,omitempty"`
	RelatedTagID string `json:"relatedTagID,omitempty"`
	Rank         *int   `json:"rank,omitempty"`
}

type Series struct {
	ID          string  `json:"id"`
	Slug        string  `json:"slug,omitempty"`
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Closed      bool    `json:"closed,omitempty"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	UpdatedAt   string  `json:"updatedAt,omitempty"`
	Events      []Event `json:"events,omitempty"`
}

type Comment struct {
	ID               string `json:"id"`
	Body             string `json:"body,omitempty"`
	ParentEntityType string `json:"parentEntityType,omitempty"`
	ParentEntityID   string `json:"parentEntityID,omitempty"`
	ParentCommentID  string `json:"parentCommentID,omitempty"`
	UserAddress      string `json:"userAddress,omitempty"`
	ReplyAddress     string `json:"replyAddress,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
}

type PublicProfileUser struct {
	ID      string `json:"id,omitempty"`
	Creator *bool  `json:"creator,omitempty"`
	IsMod   *bool  `json:"mod,omitempty"`
}

type PublicProfile struct {
	CreatedAt             string              `json:"createdAt,omitempty"`
	ProxyWallet           string              `json:"proxyWallet,omitempty"`
	ProfileImage          string              `json:"profileImage,omitempty"`
	DisplayUsernamePublic *bool               `json:"displayUsernamePublic,omitempty"`
	Bio                   string              `json:"bio,omitempty"`
	Pseudonym             string              `json:"pseudonym,omitempty"`
	Name                  string              `json:"name,omitempty"`
	XUsername             string              `json:"xUsername,omitempty"`
	VerifiedBadge         *bool               `json:"verifiedBadge,omitempty"`
	Users                 []PublicProfileUser `json:"users,omitempty"`
}

type SearchTag struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
	Slug  string `json:"slug,omitempty"`
}

type Profile struct {
	ID           string `json:"id,omitempty"`
	Pseudonym    string `json:"pseudonym,omitempty"`
	Name         string `json:"name,omitempty"`
	ProfileImage string `json:"profileImage,omitempty"`
	ProxyWallet  string `json:"proxyWallet,omitempty"`
}

type Pagination struct {
	HasMore      *bool `json:"hasMore,omitempty"`
	TotalResults *int  `json:"totalResults,omitempty"`
}

type SearchResults struct {
	Events     []Event     `json:"events,omitempty"`
	Tags       []SearchTag `json:"tags,omitempty"`
	Profiles   []Profile   `json:"profiles,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type StatusResponse string
