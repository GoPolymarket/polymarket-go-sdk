package data

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/errors"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/types"

	"github.com/ethereum/go-ethereum/common"
)

const (
	BaseURL = "https://data-api.polymarket.com"
)

// Use unified error definitions from pkg/errors
var (
	ErrMissingRequest      = sdkerrors.ErrMissingRequest
	ErrMissingUser         = sdkerrors.ErrMissingUser
	ErrInvalidMarketFilter = sdkerrors.ErrInvalidMarketFilter
	ErrInvalidTradeFilter  = sdkerrors.ErrInvalidTradeFilter
)

type clientImpl struct {
	httpClient *transport.Client
}

// NewClient creates a Data API client.
func NewClient(httpClient *transport.Client) Client {
	if httpClient == nil {
		httpClient = transport.NewClient(nil, BaseURL)
	}
	return &clientImpl{httpClient: httpClient}
}

func (c *clientImpl) Health(ctx context.Context) (string, error) {
	var resp HealthResponse
	if err := c.httpClient.Get(ctx, "", nil, &resp); err != nil {
		var raw string
		if err2 := c.httpClient.Get(ctx, "", nil, &raw); err2 == nil {
			return raw, nil
		}
		return "", err
	}
	return resp.Data, nil
}

func (c *clientImpl) Positions(ctx context.Context, req *PositionsRequest) (PositionsResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if req.User == (common.Address{}) {
		return nil, ErrMissingUser
	}
	if err := validateIntRange(req.Limit, 0, 500, "limit"); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.Offset, 0, 10000, "offset"); err != nil {
		return nil, err
	}
	if err := validateTitle(req.Title, 100); err != nil {
		return nil, err
	}
	q.Set("user", req.User.Hex())
	if err := applyMarketFilter(q, req.Filter); err != nil {
		return nil, err
	}
	addDecimal(q, "sizeThreshold", req.SizeThreshold)
	addBool(q, "redeemable", req.Redeemable)
	addBool(q, "mergeable", req.Mergeable)
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)
	addString(q, "sortBy", req.SortBy)
	addString(q, "sortDirection", req.SortDirection)
	addString(q, "title", req.Title)

	var resp PositionsResponse
	err := c.httpClient.Get(ctx, "/positions", q, &resp)
	return resp, err
}

func (c *clientImpl) Trades(ctx context.Context, req *TradesRequest) (TradesResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if err := validateIntRange(req.Limit, 0, 10000, "limit"); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.Offset, 0, 10000, "offset"); err != nil {
		return nil, err
	}
	addAddress(q, "user", req.User)
	if err := applyMarketFilter(q, req.Filter); err != nil {
		return nil, err
	}
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)
	addBool(q, "takerOnly", req.TakerOnly)
	if err := applyTradeFilter(q, req.TradeFilter); err != nil {
		return nil, err
	}
	addString(q, "side", req.Side)

	var resp TradesResponse
	err := c.httpClient.Get(ctx, "/trades", q, &resp)
	return resp, err
}

func (c *clientImpl) Activity(ctx context.Context, req *ActivityRequest) (ActivityResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if req.User == (common.Address{}) {
		return nil, ErrMissingUser
	}
	if err := validateIntRange(req.Limit, 0, 500, "limit"); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.Offset, 0, 10000, "offset"); err != nil {
		return nil, err
	}
	if err := validateInt64Min(req.Start, 0, "start"); err != nil {
		return nil, err
	}
	if err := validateInt64Min(req.End, 0, "end"); err != nil {
		return nil, err
	}
	q.Set("user", req.User.Hex())
	if err := applyMarketFilter(q, req.Filter); err != nil {
		return nil, err
	}
	addStringSlice(q, "type", activityTypeStrings(req.ActivityTypes))
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)
	addInt64(q, "start", req.Start)
	addInt64(q, "end", req.End)
	addString(q, "sortBy", req.SortBy)
	addString(q, "sortDirection", req.SortDirection)
	addString(q, "side", req.Side)

	var resp ActivityResponse
	err := c.httpClient.Get(ctx, "/activity", q, &resp)
	return resp, err
}

func (c *clientImpl) Holders(ctx context.Context, req *HoldersRequest) (HoldersResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if err := validateIntRange(req.Limit, 0, 20, "limit"); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.MinBalance, 0, 999999, "min_balance"); err != nil {
		return nil, err
	}
	addHashSlice(q, "market", req.Markets)
	addInt(q, "limit", req.Limit)
	addInt(q, "minBalance", req.MinBalance)

	var resp HoldersResponse
	err := c.httpClient.Get(ctx, "/holders", q, &resp)
	return resp, err
}

func (c *clientImpl) Value(ctx context.Context, req *ValueRequest) (ValueResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	q.Set("user", req.User.Hex())
	addHashSlice(q, "market", req.Markets)

	var resp ValueResponse
	err := c.httpClient.Get(ctx, "/value", q, &resp)
	return resp, err
}

func (c *clientImpl) ClosedPositions(ctx context.Context, req *ClosedPositionsRequest) (ClosedPositionsResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if req.User == (common.Address{}) {
		return nil, ErrMissingUser
	}
	if err := validateTitle(req.Title, 100); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.Limit, 0, 50, "limit"); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.Offset, 0, 100000, "offset"); err != nil {
		return nil, err
	}
	q.Set("user", req.User.Hex())
	if err := applyMarketFilter(q, req.Filter); err != nil {
		return nil, err
	}
	addString(q, "title", req.Title)
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)
	addString(q, "sortBy", req.SortBy)
	addString(q, "sortDirection", req.SortDirection)

	var resp ClosedPositionsResponse
	err := c.httpClient.Get(ctx, "/closed-positions", q, &resp)
	return resp, err
}

func (c *clientImpl) Traded(ctx context.Context, req *TradedRequest) (TradedResponse, error) {
	if req == nil {
		return TradedResponse{}, ErrMissingRequest
	}
	q := url.Values{}
	q.Set("user", req.User.Hex())

	var resp TradedResponse
	err := c.httpClient.Get(ctx, "/traded", q, &resp)
	return resp, err
}

func (c *clientImpl) OpenInterest(ctx context.Context, req *OpenInterestRequest) (OpenInterestResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	addHashSlice(q, "market", req.Markets)

	var resp OpenInterestResponse
	err := c.httpClient.Get(ctx, "/oi", q, &resp)
	return resp, err
}

func (c *clientImpl) LiveVolume(ctx context.Context, req *LiveVolumeRequest) (LiveVolumeResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if req.ID == 0 {
		return nil, fmt.Errorf("id is required")
	}
	q.Set("id", strconv.FormatInt(req.ID, 10))

	var resp LiveVolumeResponse
	err := c.httpClient.Get(ctx, "/live-volume", q, &resp)
	return resp, err
}

func (c *clientImpl) Leaderboard(ctx context.Context, req *LeaderboardRequest) (LeaderboardResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if err := validateIntRange(req.Limit, 1, 50, "limit"); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.Offset, 0, 1000, "offset"); err != nil {
		return nil, err
	}
	addString(q, "category", req.Category)
	addString(q, "timePeriod", req.TimePeriod)
	addString(q, "orderBy", req.OrderBy)
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)
	addAddress(q, "user", req.User)
	addString(q, "userName", req.UserName)

	var resp LeaderboardResponse
	err := c.httpClient.Get(ctx, "/v1/leaderboard", q, &resp)
	return resp, err
}

func (c *clientImpl) BuildersLeaderboard(ctx context.Context, req *BuildersLeaderboardRequest) (BuildersLeaderboardResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	if err := validateIntRange(req.Limit, 0, 50, "limit"); err != nil {
		return nil, err
	}
	if err := validateIntRange(req.Offset, 0, 1000, "offset"); err != nil {
		return nil, err
	}
	addString(q, "timePeriod", req.TimePeriod)
	addInt(q, "limit", req.Limit)
	addInt(q, "offset", req.Offset)

	var resp BuildersLeaderboardResponse
	err := c.httpClient.Get(ctx, "/v1/builders/leaderboard", q, &resp)
	return resp, err
}

func (c *clientImpl) BuildersVolume(ctx context.Context, req *BuildersVolumeRequest) (BuildersVolumeResponse, error) {
	if req == nil {
		return nil, ErrMissingRequest
	}
	q := url.Values{}
	addString(q, "timePeriod", req.TimePeriod)

	var resp BuildersVolumeResponse
	err := c.httpClient.Get(ctx, "/v1/builders/volume", q, &resp)
	return resp, err
}

// Helpers.
func addString(q url.Values, key string, val interface{}) {
	if val == nil {
		return
	}
	switch v := val.(type) {
	case *string:
		if v != nil && *v != "" {
			q.Set(key, *v)
		}
	case *PositionSortBy:
		q.Set(key, stringValue(v))
	case *ClosedPositionSortBy:
		q.Set(key, stringValue(v))
	case *ActivitySortBy:
		q.Set(key, stringValue(v))
	case *SortDirection:
		q.Set(key, stringValue(v))
	case *Side:
		q.Set(key, stringValue(v))
	case *LeaderboardCategory:
		q.Set(key, stringValue(v))
	case *LeaderboardOrderBy:
		q.Set(key, stringValue(v))
	case *TimePeriod:
		q.Set(key, stringValue(v))
	default:
		if s, ok := val.(string); ok && s != "" {
			q.Set(key, s)
		}
	}
}

func addStringSlice(q url.Values, key string, vals []string) {
	if len(vals) == 0 {
		return
	}
	filtered := make([]string, 0, len(vals))
	for _, v := range vals {
		if v != "" {
			filtered = append(filtered, v)
		}
	}
	if len(filtered) == 0 {
		return
	}
	q.Set(key, strings.Join(filtered, ","))
}

func addInt64Slice(q url.Values, key string, vals []int64) {
	if len(vals) == 0 {
		return
	}
	filtered := make([]string, 0, len(vals))
	for _, v := range vals {
		filtered = append(filtered, strconv.FormatInt(v, 10))
	}
	if len(filtered) == 0 {
		return
	}
	q.Set(key, strings.Join(filtered, ","))
}

func addInt(q url.Values, key string, val *int) {
	if val != nil {
		q.Set(key, strconv.Itoa(*val))
	}
}

func addInt64(q url.Values, key string, val *int64) {
	if val != nil {
		q.Set(key, strconv.FormatInt(*val, 10))
	}
}

func addBool(q url.Values, key string, val *bool) {
	if val != nil {
		q.Set(key, strconv.FormatBool(*val))
	}
}

func addDecimal(q url.Values, key string, val *types.Decimal) {
	if val != nil {
		q.Set(key, val.String())
	}
}

func addAddress(q url.Values, key string, val *common.Address) {
	if val != nil {
		q.Set(key, val.Hex())
	}
}

func addHashSlice(q url.Values, key string, vals []common.Hash) {
	if len(vals) == 0 {
		return
	}
	out := make([]string, 0, len(vals))
	for _, h := range vals {
		if (h != common.Hash{}) {
			out = append(out, h.Hex())
		}
	}
	if len(out) == 0 {
		return
	}
	q.Set(key, strings.Join(out, ","))
}

func applyMarketFilter(q url.Values, filter *MarketFilter) error {
	if filter == nil {
		return nil
	}
	if len(filter.Markets) > 0 && len(filter.EventIDs) > 0 {
		return ErrInvalidMarketFilter
	}
	if len(filter.Markets) > 0 {
		addHashSlice(q, "market", filter.Markets)
	}
	if len(filter.EventIDs) > 0 {
		addInt64Slice(q, "eventId", filter.EventIDs)
	}
	return nil
}

func applyTradeFilter(q url.Values, filter *TradeFilter) error {
	if filter == nil {
		return nil
	}
	if filter.FilterType == "" {
		return ErrInvalidTradeFilter
	}
	if filter.FilterAmount.Sign() < 0 {
		return fmt.Errorf("filter_amount must be non-negative")
	}
	q.Set("filterType", string(filter.FilterType))
	q.Set("filterAmount", filter.FilterAmount.String())
	return nil
}

func validateIntRange(val *int, min, max int, name string) error {
	if val == nil {
		return nil
	}
	if *val < min || *val > max {
		return BoundedIntError{Value: *val, Min: min, Max: max, ParamName: name}
	}
	return nil
}

func validateInt64Min(val *int64, min int64, name string) error {
	if val == nil {
		return nil
	}
	if *val < min {
		return fmt.Errorf("%s must be >= %d (got %d)", name, min, *val)
	}
	return nil
}

func validateTitle(val *string, maxLen int) error {
	if val == nil {
		return nil
	}
	if len(*val) > maxLen {
		return fmt.Errorf("title must be at most %d characters", maxLen)
	}
	return nil
}

func stringValue[T ~string](v *T) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

func activityTypeStrings(vals []ActivityType) []string {
	if len(vals) == 0 {
		return nil
	}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if v == "" {
			continue
		}
		out = append(out, string(v))
	}
	return out
}
