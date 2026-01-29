// Package gamma provides the implementation for the Polymarket Gamma metadata service.
package gamma

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

const (
	// BaseURL is the default production endpoint for Gamma.
	BaseURL = "https://gamma-api.polymarket.com"
)

type clientImpl struct {
	httpClient *transport.Client
}

// NewClient creates a new Gamma API client.
func NewClient(httpClient *transport.Client) Client {
	if httpClient == nil {
		httpClient = transport.NewClient(nil, BaseURL)
	}
	return &clientImpl{
		httpClient: httpClient,
	}
}

func addInt(q url.Values, key string, val *int) {
	if val != nil {
		q.Set(key, strconv.Itoa(*val))
	}
}

func addBool(q url.Values, key string, val *bool) {
	if val != nil {
		q.Set(key, strconv.FormatBool(*val))
	}
}

func addString(q url.Values, key, val string) {
	if val != "" {
		q.Set(key, val)
	}
}

func addStringSlice(q url.Values, key string, vals []string) {
	for _, v := range vals {
		if v == "" {
			continue
		}
		q.Add(key, v)
	}
}

func valueOrEmpty(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

func optionalStringSlice(val string) []string {
	if val == "" {
		return nil
	}
	return []string{val}
}

func (c *clientImpl) Status(ctx context.Context) (StatusResponse, error) {
	var resp string
	err := c.httpClient.Get(ctx, "/status", nil, &resp)
	return StatusResponse(resp), err
}

func (c *clientImpl) Teams(ctx context.Context, req *TeamsRequest) ([]Team, error) {
	q := url.Values{}
	if req != nil {
		addInt(q, "limit", req.Limit)
		addInt(q, "offset", req.Offset)
		addString(q, "order", req.Order)
		addBool(q, "ascending", req.Ascending)
		addStringSlice(q, "league", req.League)
		addStringSlice(q, "name", req.Name)
		addStringSlice(q, "abbreviation", req.Abbreviation)
	}
	var resp []Team
	err := c.httpClient.Get(ctx, "/teams", q, &resp)
	return resp, err
}

func (c *clientImpl) Sports(ctx context.Context) ([]SportsMetadata, error) {
	var resp []SportsMetadata
	err := c.httpClient.Get(ctx, "/sports", nil, &resp)
	return resp, err
}

func (c *clientImpl) SportsMarketTypes(ctx context.Context) (SportsMarketTypesResponse, error) {
	var resp SportsMarketTypesResponse
	err := c.httpClient.Get(ctx, "/sports/market-types", nil, &resp)
	return resp, err
}

func (c *clientImpl) Tags(ctx context.Context, req *TagsRequest) ([]Tag, error) {
	q := url.Values{}
	if req != nil {
		addInt(q, "limit", req.Limit)
		addInt(q, "offset", req.Offset)
		addString(q, "order", req.Order)
		addBool(q, "ascending", req.Ascending)
		addBool(q, "include_template", req.IncludeTemplate)
		addBool(q, "is_carousel", req.IsCarousel)
	}
	var resp []Tag
	err := c.httpClient.Get(ctx, "/tags", q, &resp)
	return resp, err
}

func (c *clientImpl) TagByID(ctx context.Context, req *TagByIDRequest) (*Tag, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	q := url.Values{}
	addBool(q, "include_template", req.IncludeTemplate)
	var resp Tag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/tags/%s", req.ID), q, &resp)
	return &resp, err
}

func (c *clientImpl) TagBySlug(ctx context.Context, req *TagBySlugRequest) (*Tag, error) {
	if req == nil || req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	q := url.Values{}
	addBool(q, "include_template", req.IncludeTemplate)
	var resp Tag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/tags/slug/%s", req.Slug), q, &resp)
	return &resp, err
}

func (c *clientImpl) RelatedTagsByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]RelatedTag, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	q := url.Values{}
	addBool(q, "omit_empty", req.OmitEmpty)
	addString(q, "status", req.Status)
	var resp []RelatedTag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/tags/%s/related-tags", req.ID), q, &resp)
	return resp, err
}

func (c *clientImpl) RelatedTagsBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]RelatedTag, error) {
	if req == nil || req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	q := url.Values{}
	addBool(q, "omit_empty", req.OmitEmpty)
	addString(q, "status", req.Status)
	var resp []RelatedTag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/tags/slug/%s/related-tags", req.Slug), q, &resp)
	return resp, err
}

func (c *clientImpl) TagsRelatedToTagByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]Tag, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	q := url.Values{}
	addBool(q, "omit_empty", req.OmitEmpty)
	addString(q, "status", req.Status)
	var resp []Tag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/tags/%s/related-tags/tags", req.ID), q, &resp)
	return resp, err
}

func (c *clientImpl) TagsRelatedToTagBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]Tag, error) {
	if req == nil || req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	q := url.Values{}
	addBool(q, "omit_empty", req.OmitEmpty)
	addString(q, "status", req.Status)
	var resp []Tag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/tags/slug/%s/related-tags/tags", req.Slug), q, &resp)
	return resp, err
}

func buildEventsQuery(req *EventsRequest) url.Values {
	q := url.Values{}
	if req == nil {
		return q
	}
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)
	addBool(q, "ascending", req.Ascending)
	addStringSlice(q, "order", req.Order)
	addStringSlice(q, "id", req.IDs)
	addString(q, "tag_id", req.TagID)
	addStringSlice(q, "exclude_tag_id", req.ExcludeTagID)
	addStringSlice(q, "slug", req.Slugs)
	addString(q, "tag_slug", req.TagSlug)
	addBool(q, "related_tags", req.RelatedTags)
	addBool(q, "active", req.Active)
	addBool(q, "archived", req.Archived)
	addBool(q, "featured", req.Featured)
	addBool(q, "cyom", req.Cyom)
	addBool(q, "include_chat", req.IncludeChat)
	addBool(q, "include_template", req.IncludeTemplate)
	addString(q, "recurrence", req.Recurrence)
	addBool(q, "closed", req.Closed)
	addString(q, "liquidity_min", valueOrEmpty(req.LiquidityMin))
	addString(q, "liquidity_max", valueOrEmpty(req.LiquidityMax))
	addString(q, "volume_min", valueOrEmpty(req.VolumeMin))
	addString(q, "volume_max", valueOrEmpty(req.VolumeMax))
	addString(q, "start_date_min", req.StartDateMin)
	addString(q, "start_date_max", req.StartDateMax)
	addString(q, "end_date_min", req.EndDateMin)
	addString(q, "end_date_max", req.EndDateMax)
	return q
}

func (c *clientImpl) Events(ctx context.Context, req *EventsRequest) ([]Event, error) {
	q := buildEventsQuery(req)
	var resp []Event
	err := c.httpClient.Get(ctx, "/events", q, &resp)
	return resp, err
}

func (c *clientImpl) EventsAll(ctx context.Context, req *EventsRequest) ([]Event, error) {
	limit := 100
	if req != nil && req.Limit != nil {
		limit = *req.Limit
	}
	offset := 0
	if req != nil && req.Offset != nil {
		offset = *req.Offset
	}

	var results []Event
	for {
		nextReq := EventsRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.Limit = &limit
		nextReq.Offset = &offset

		resp, err := c.Events(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp...)

		if len(resp) < limit {
			break
		}
		offset += limit
	}
	return results, nil
}

func (c *clientImpl) EventByID(ctx context.Context, req *EventByIDRequest) (*Event, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	q := url.Values{}
	addBool(q, "include_chat", req.IncludeChat)
	addBool(q, "include_template", req.IncludeTemplate)
	var resp Event
	err := c.httpClient.Get(ctx, fmt.Sprintf("/events/%s", req.ID), q, &resp)
	return &resp, err
}

func (c *clientImpl) EventBySlug(ctx context.Context, req *EventBySlugRequest) (*Event, error) {
	if req == nil || req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	q := url.Values{}
	addBool(q, "include_chat", req.IncludeChat)
	addBool(q, "include_template", req.IncludeTemplate)
	var resp Event
	err := c.httpClient.Get(ctx, fmt.Sprintf("/events/slug/%s", req.Slug), q, &resp)
	return &resp, err
}

func (c *clientImpl) EventTags(ctx context.Context, req *EventTagsRequest) ([]Tag, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	var resp []Tag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/events/%s/tags", req.ID), nil, &resp)
	return resp, err
}

func buildMarketsQuery(req *MarketsRequest) url.Values {
	q := url.Values{}
	if req != nil {
		addInt(q, "limit", req.Limit)
		addInt(q, "offset", req.Offset)
		addString(q, "order", req.Order)
		addBool(q, "ascending", req.Ascending)
		addString(q, "slug", req.Slug)
		addStringSlice(q, "slug", req.Slugs)
		addStringSlice(q, "id", req.IDs)
		addStringSlice(q, "clob_token_ids", req.ClobTokenIDs)
		addStringSlice(q, "condition_ids", req.ConditionIDs)
		addStringSlice(q, "market_maker_address", req.MarketMakerAddress)
		addBool(q, "active", req.Active)
		addBool(q, "closed", req.Closed)
		addString(q, "tag_id", req.TagID)
		addString(q, "tag_slug", req.TagSlug)
		addBool(q, "related_tags", req.RelatedTags)
		addBool(q, "cyom", req.Cyom)
		addString(q, "uma_resolution_status", req.UmaResolutionStatus)
		addString(q, "game_id", req.GameID)
		addStringSlice(q, "sports_market_types", req.SportsMarketTypes)
		addString(q, "volume_min", valueOrEmpty(req.VolumeMin))
		addString(q, "volume_max", valueOrEmpty(req.VolumeMax))
		addString(q, "liquidity_min", valueOrEmpty(req.LiquidityMin))
		addString(q, "liquidity_max", valueOrEmpty(req.LiquidityMax))
		addString(q, "liquidity_num_min", valueOrEmpty(req.LiquidityNumMin))
		addString(q, "liquidity_num_max", valueOrEmpty(req.LiquidityNumMax))
		addString(q, "volume_num_min", valueOrEmpty(req.VolumeNumMin))
		addString(q, "volume_num_max", valueOrEmpty(req.VolumeNumMax))
		addString(q, "start_date_min", req.StartDateMin)
		addString(q, "start_date_max", req.StartDateMax)
		addString(q, "end_date_min", req.EndDateMin)
		addString(q, "end_date_max", req.EndDateMax)
		addString(q, "rewards_min_size", valueOrEmpty(req.RewardsMinSize))
		addString(q, "rewards_max_size", valueOrEmpty(req.RewardsMaxSize))
	}
	return q
}

func (c *clientImpl) Markets(ctx context.Context, req *MarketsRequest) ([]Market, error) {
	q := buildMarketsQuery(req)
	var resp []Market
	err := c.httpClient.Get(ctx, "/markets", q, &resp)
	return resp, err
}

func (c *clientImpl) MarketsAll(ctx context.Context, req *MarketsRequest) ([]Market, error) {
	limit := 100
	if req != nil && req.Limit != nil {
		limit = *req.Limit
	}
	offset := 0
	if req != nil && req.Offset != nil {
		offset = *req.Offset
	}

	var results []Market
	for {
		nextReq := MarketsRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.Limit = &limit
		nextReq.Offset = &offset

		resp, err := c.Markets(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp...)

		if len(resp) < limit {
			break
		}
		offset += limit
	}
	return results, nil
}

func (c *clientImpl) MarketByID(ctx context.Context, req *MarketByIDRequest) (*Market, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	q := url.Values{}
	addBool(q, "include_tag", req.IncludeTag)
	var resp Market
	err := c.httpClient.Get(ctx, fmt.Sprintf("/markets/%s", req.ID), q, &resp)
	return &resp, err
}

func (c *clientImpl) MarketBySlug(ctx context.Context, req *MarketBySlugRequest) (*Market, error) {
	if req == nil || req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	q := url.Values{}
	addBool(q, "include_tag", req.IncludeTag)
	var resp Market
	err := c.httpClient.Get(ctx, fmt.Sprintf("/markets/slug/%s", req.Slug), q, &resp)
	return &resp, err
}

func (c *clientImpl) MarketTags(ctx context.Context, req *MarketTagsRequest) ([]Tag, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	var resp []Tag
	err := c.httpClient.Get(ctx, fmt.Sprintf("/markets/%s/tags", req.ID), nil, &resp)
	return resp, err
}

func (c *clientImpl) Series(ctx context.Context, req *SeriesRequest) ([]Series, error) {
	q := url.Values{}
	if req != nil {
		addInt(q, "limit", req.Limit)
		addInt(q, "offset", req.Offset)
		addString(q, "order", req.Order)
		addBool(q, "ascending", req.Ascending)
		addStringSlice(q, "slug", req.Slugs)
		addStringSlice(q, "categories_ids", req.CategoriesIDs)
		addStringSlice(q, "categories_labels", req.CategoriesLabels)
		addBool(q, "closed", req.Closed)
		addBool(q, "include_chat", req.IncludeChat)
		addString(q, "recurrence", req.Recurrence)
	}
	var resp []Series
	err := c.httpClient.Get(ctx, "/series", q, &resp)
	return resp, err
}

func (c *clientImpl) SeriesByID(ctx context.Context, req *SeriesByIDRequest) (*Series, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	q := url.Values{}
	addBool(q, "include_chat", req.IncludeChat)
	var resp Series
	err := c.httpClient.Get(ctx, fmt.Sprintf("/series/%s", req.ID), q, &resp)
	return &resp, err
}

func (c *clientImpl) Comments(ctx context.Context, req *CommentsRequest) ([]Comment, error) {
	q := url.Values{}
	if req != nil {
		addString(q, "parent_entity_type", req.ParentEntityType)
		addString(q, "parent_entity_id", req.ParentEntityID)
		addInt(q, "limit", req.Limit)
		addInt(q, "offset", req.Offset)
		addString(q, "order", req.Order)
		addBool(q, "ascending", req.Ascending)
		addBool(q, "get_positions", req.GetPositions)
		addBool(q, "holders_only", req.HoldersOnly)
	}
	var resp []Comment
	err := c.httpClient.Get(ctx, "/comments", q, &resp)
	return resp, err
}

func (c *clientImpl) CommentByID(ctx context.Context, req *CommentByIDRequest) ([]Comment, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	q := url.Values{}
	addBool(q, "get_positions", req.GetPositions)
	var resp []Comment
	err := c.httpClient.Get(ctx, fmt.Sprintf("/comments/%s", req.ID), q, &resp)
	return resp, err
}

func (c *clientImpl) CommentsByUserAddress(ctx context.Context, req *CommentsByUserAddressRequest) ([]Comment, error) {
	if req == nil || req.UserAddress == "" {
		return nil, fmt.Errorf("user_address is required")
	}
	q := url.Values{}
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)
	addString(q, "order", req.Order)
	addBool(q, "ascending", req.Ascending)
	var resp []Comment
	err := c.httpClient.Get(ctx, fmt.Sprintf("/comments/user_address/%s", req.UserAddress), q, &resp)
	return resp, err
}

func (c *clientImpl) PublicProfile(ctx context.Context, req *PublicProfileRequest) (*PublicProfile, error) {
	if req == nil || req.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	q := url.Values{}
	addString(q, "address", req.Address)
	var resp PublicProfile
	err := c.httpClient.Get(ctx, "/public-profile", q, &resp)
	return &resp, err
}

func (c *clientImpl) PublicSearch(ctx context.Context, req *PublicSearchRequest) (SearchResults, error) {
	q := url.Values{}
	if req != nil {
		addString(q, "q", req.Query)
		addBool(q, "cache", req.Cache)
		addString(q, "events_status", req.EventsStatus)
		addInt(q, "limit_per_type", req.LimitPerType)
		addInt(q, "page", req.Page)
		addStringSlice(q, "events_tag", req.EventsTag)
		if req.KeepClosedMarkets != nil {
			q.Set("keep_closed_markets", strconv.Itoa(*req.KeepClosedMarkets))
		}
		addString(q, "sort", req.Sort)
		addBool(q, "ascending", req.Ascending)
		addBool(q, "search_tags", req.SearchTags)
		addBool(q, "search_profiles", req.SearchProfiles)
		addString(q, "recurrence", req.Recurrence)
		addStringSlice(q, "exclude_tag_id", req.ExcludeTagID)
		addBool(q, "optimized", req.Optimized)
	}
	var resp SearchResults
	err := c.httpClient.Get(ctx, "/public-search", q, &resp)
	return resp, err
}

// Backwards compatible aliases.
func (c *clientImpl) GetMarkets(ctx context.Context, req *MarketsRequest) ([]Market, error) {
	return c.Markets(ctx, req)
}

func (c *clientImpl) GetMarket(ctx context.Context, id string) (*Market, error) {
	return c.MarketByID(ctx, &MarketByIDRequest{ID: id})
}

func (c *clientImpl) GetEvents(ctx context.Context, req *MarketsRequest) ([]Event, error) {
	// Preserve legacy signature by mapping a subset of fields.
	var mapped *EventsRequest
	if req != nil {
		mapped = &EventsRequest{
			Limit:        req.Limit,
			Offset:       req.Offset,
			Order:        optionalStringSlice(req.Order),
			Ascending:    req.Ascending,
			Slugs:        optionalStringSlice(req.Slug),
			Active:       req.Active,
			Closed:       req.Closed,
			TagID:        req.TagID,
			VolumeMin:    req.VolumeMin,
			VolumeMax:    req.VolumeMax,
			LiquidityMin: req.LiquidityMin,
			StartDateMin: req.StartDateMin,
			EndDateMax:   req.EndDateMax,
		}
	}
	return c.Events(ctx, mapped)
}

func (c *clientImpl) GetEvent(ctx context.Context, id string) (*Event, error) {
	return c.EventByID(ctx, &EventByIDRequest{ID: id})
}