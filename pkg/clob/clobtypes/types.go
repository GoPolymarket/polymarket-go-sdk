package clobtypes

import (
	"bytes"
	"encoding/json"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

// OrderType represents time-in-force / order type values.
type OrderType string

const (
	OrderTypeGTC OrderType = "GTC"
	OrderTypeGTD OrderType = "GTD"
	OrderTypeFAK OrderType = "FAK"
	OrderTypeFOK OrderType = "FOK"
)

// PriceHistoryInterval represents the supported time intervals for price history.
type PriceHistoryInterval string

const (
	PriceHistoryInterval1m  PriceHistoryInterval = "1m"
	PriceHistoryInterval1h  PriceHistoryInterval = "1h"
	PriceHistoryInterval6h  PriceHistoryInterval = "6h"
	PriceHistoryInterval1d  PriceHistoryInterval = "1d"
	PriceHistoryInterval1w  PriceHistoryInterval = "1w"
	PriceHistoryIntervalMax PriceHistoryInterval = "max"
)

const (
	InitialCursor = "MA=="
	EndCursor     = "LTE="
)

// Request types.
type (
	MarketsRequest struct {
		Limit   int    `json:"limit,omitempty"`
		Cursor  string `json:"cursor,omitempty"`
		Active  *bool  `json:"active,omitempty"`
		AssetID string `json:"asset_id,omitempty"`
	}
	BookRequest struct {
		TokenID string `json:"token_id"`
		Side    string `json:"side,omitempty"`
	}
	BooksRequest struct {
		// Requests is the preferred batch form (one entry per token, optional side).
		Requests []BookRequest `json:"requests,omitempty"`
		// TokenIDs is deprecated; prefer Requests.
		TokenIDs []string `json:"token_ids,omitempty"`
	}
	MidpointRequest struct {
		TokenID string `json:"token_id"`
	}
	MidpointsRequest struct {
		TokenIDs []string `json:"token_ids"`
	}
	PriceRequest struct {
		TokenID string `json:"token_id"`
		Side    string `json:"side,omitempty"`
	}
	PricesRequest struct {
		// Requests is the preferred batch form (one entry per token with side).
		Requests []PriceRequest `json:"requests,omitempty"`
		// TokenIDs is deprecated; prefer Requests.
		TokenIDs []string `json:"token_ids,omitempty"`
		// Side is deprecated; prefer per-request side in Requests.
		Side string `json:"side,omitempty"`
	}
	SpreadRequest struct {
		TokenID string `json:"token_id"`
		Side    string `json:"side,omitempty"`
	}
	SpreadsRequest struct {
		// Requests is the preferred batch form (one entry per token, optional side).
		Requests []SpreadRequest `json:"requests,omitempty"`
		// TokenIDs is deprecated; prefer Requests.
		TokenIDs []string `json:"token_ids,omitempty"`
	}
	LastTradePriceRequest struct {
		TokenID string `json:"token_id"`
	}
	LastTradesPricesRequest struct {
		TokenIDs []string `json:"token_ids"`
	}
	TickSizeRequest struct {
		TokenID string `json:"token_id"`
	}
	NegRiskRequest struct {
		TokenID string `json:"token_id"`
	}
	FeeRateRequest struct {
		TokenID string `json:"token_id"`
	}
	PricesHistoryRequest struct {
		// Market is the condition ID (preferred by the API).
		Market string `json:"market,omitempty"`
		// TokenID is a legacy token identifier supported for backwards compatibility.
		TokenID string `json:"token_id,omitempty"`
		// Interval specifies a predefined time range (e.g. "1m", "1h", "1d", "1w", "max").
		Interval PriceHistoryInterval `json:"interval,omitempty"`
		// StartTs and EndTs specify an explicit time range (Unix seconds).
		StartTs int64 `json:"start_ts,omitempty"`
		EndTs   int64 `json:"end_ts,omitempty"`
		// Resolution is a legacy alias for Interval.
		Resolution string `json:"resolution,omitempty"`
		// Fidelity controls the number of datapoints to return (optional).
		Fidelity int `json:"fidelity,omitempty"`
	}
	SignableOrder struct {
		Order     *Order    `json:"order"`
		OrderType OrderType `json:"order_type"`
		PostOnly  *bool     `json:"post_only,omitempty"`
	}
	OrderOptions struct {
		OrderType OrderType
		PostOnly  *bool
		DeferExec *bool
	}
	SignedOrder struct {
		Order     Order  `json:"order"`
		Signature string `json:"signature"`
		Owner     string `json:"owner"`

		// Options used when submitting the order (not serialized directly).
		OrderType OrderType `json:"-"`
		PostOnly  *bool     `json:"-"`
		DeferExec *bool     `json:"-"`
	}
	SignedOrders struct {
		Orders []SignedOrder `json:"orders"`
	}
	CancelOrderRequest struct {
		// Deprecated: prefer OrderID.
		ID string `json:"id,omitempty"`
		// OrderID is the canonical field used by the API ("orderId").
		OrderID string `json:"orderId,omitempty"`
	}
	CancelOrdersRequest struct {
		// Deprecated: prefer OrderIDs.
		IDs []string `json:"ids,omitempty"`
		// OrderIDs is the canonical batch payload.
		OrderIDs []string `json:"orderIds,omitempty"`
	}
	CancelMarketOrdersRequest struct {
		// Market is the condition ID (preferred by the API).
		Market string `json:"market,omitempty"`
		// AssetID is an optional asset filter.
		AssetID string `json:"asset_id,omitempty"`
		// Deprecated: legacy field name.
		MarketID string `json:"market_id,omitempty"`
	}
	OrdersRequest struct {
		ID         string `json:"id,omitempty"`
		Market     string `json:"market,omitempty"`
		AssetID    string `json:"asset_id,omitempty"`
		Limit      int    `json:"limit,omitempty"`
		Cursor     string `json:"cursor,omitempty"`
		NextCursor string `json:"next_cursor,omitempty"`
	}
	TradesRequest struct {
		ID         string `json:"id,omitempty"`
		Taker      string `json:"taker,omitempty"`
		Maker      string `json:"maker,omitempty"`
		Market     string `json:"market,omitempty"`
		AssetID    string `json:"asset_id,omitempty"`
		Before     int64  `json:"before,omitempty"`
		After      int64  `json:"after,omitempty"`
		Limit      int    `json:"limit,omitempty"`
		Cursor     string `json:"cursor,omitempty"`
		NextCursor string `json:"next_cursor,omitempty"`
	}
	OrderScoringRequest struct {
		ID string `json:"id"`
	}
	OrdersScoringRequest struct {
		IDs []string `json:"ids"`
	}
	AssetType               string
	BalanceAllowanceRequest struct {
		// Asset is deprecated; prefer AssetType + TokenID.
		Asset string `json:"asset,omitempty"`
		// AssetType is "COLLATERAL" or "CONDITIONAL".
		AssetType AssetType `json:"asset_type,omitempty"`
		// TokenID is required when AssetType=CONDITIONAL.
		TokenID string `json:"token_id,omitempty"`
		// SignatureType is the user signature type (0=EOA, 1=Proxy, 2=Safe).
		SignatureType *int `json:"signature_type,omitempty"`
	}
	BalanceAllowanceUpdateRequest struct {
		// Asset is deprecated; prefer AssetType + TokenID.
		Asset string `json:"asset,omitempty"`
		// AssetType is "COLLATERAL" or "CONDITIONAL".
		AssetType AssetType `json:"asset_type,omitempty"`
		// TokenID is required when AssetType=CONDITIONAL.
		TokenID string `json:"token_id,omitempty"`
		// SignatureType is the user signature type (0=EOA, 1=Proxy, 2=Safe).
		SignatureType *int `json:"signature_type,omitempty"`
		// Amount is deprecated by the API but kept for compatibility.
		Amount string `json:"amount,omitempty"`
	}
	NotificationsRequest struct {
		Limit int `json:"limit,omitempty"`
	}
	DropNotificationsRequest struct {
		// Deprecated: prefer IDs.
		ID string `json:"id,omitempty"`
		// IDs is a list of notification IDs to drop.
		IDs []string `json:"ids,omitempty"`
	}
	UserEarningsRequest struct {
		// Date is required by the API (YYYY-MM-DD).
		Date string `json:"date,omitempty"`
		// SignatureType is the user signature type (0=EOA, 1=Proxy, 2=Safe).
		SignatureType *int `json:"signature_type,omitempty"`
		// NextCursor paginates results.
		NextCursor string `json:"next_cursor,omitempty"`
		// Asset is deprecated and kept for compatibility.
		Asset string `json:"asset,omitempty"`
	}
	UserTotalEarningsRequest struct {
		// Date is required by the API (YYYY-MM-DD).
		Date string `json:"date,omitempty"`
		// SignatureType is the user signature type (0=EOA, 1=Proxy, 2=Safe).
		SignatureType *int `json:"signature_type,omitempty"`
		// Asset is deprecated and kept for compatibility.
		Asset string `json:"asset,omitempty"`
	}
	UserRewardPercentagesRequest struct{}
	UserRewardsByMarketRequest   struct {
		// Date is required by the API (YYYY-MM-DD).
		Date string `json:"date,omitempty"`
		// OrderBy is the sorting key.
		OrderBy string `json:"order_by,omitempty"`
		// Position is the pagination position (if applicable).
		Position string `json:"position,omitempty"`
		// NoCompetition toggles competition filtering.
		NoCompetition bool `json:"no_competition,omitempty"`
		// SignatureType is the user signature type (0=EOA, 1=Proxy, 2=Safe).
		SignatureType *int `json:"signature_type,omitempty"`
		// NextCursor paginates results.
		NextCursor string `json:"next_cursor,omitempty"`
	}
	RewardsMarketsRequest struct {
		NextCursor string `json:"next_cursor,omitempty"`
	}
	RewardsMarketRequest struct {
		MarketID   string `json:"market_id,omitempty"`
		NextCursor string `json:"next_cursor,omitempty"`
	}
	ValidateReadonlyAPIKeyRequest struct {
		Address string `json:"address"`
		APIKey  string `json:"key"`
	}
	BuilderTradesRequest struct {
		ID         string `json:"id,omitempty"`
		Maker      string `json:"maker,omitempty"`
		Market     string `json:"market,omitempty"`
		AssetID    string `json:"asset_id,omitempty"`
		Before     int64  `json:"before,omitempty"`
		After      int64  `json:"after,omitempty"`
		Limit      int    `json:"limit,omitempty"`
		Cursor     string `json:"cursor,omitempty"`
		NextCursor string `json:"next_cursor,omitempty"`
	}
)

const (
	AssetTypeCollateral  AssetType = "COLLATERAL"
	AssetTypeConditional AssetType = "CONDITIONAL"
)

// Response types.
type (
	TimeResponse struct {
		ServerTime string `json:"server_time,omitempty"`
		Timestamp  int64  `json:"timestamp"`
	}
	MarketsResponse struct {
		Data       []Market `json:"data"`
		NextCursor string   `json:"next_cursor"`
		Limit      int      `json:"limit"`
		Count      int      `json:"count"`
	}
	MarketResponse     Market
	OrderBookResponse  OrderBook
	OrderBooksResponse []OrderBook
	MidpointResponse   struct {
		Midpoint string `json:"midpoint"`
	}
	MidpointsResponse []MidpointResponse
	PriceResponse     struct {
		Price string `json:"price"`
	}
	PricesResponse []PriceResponse
	SpreadResponse struct {
		Spread string `json:"spread"`
	}
	SpreadsResponse        []SpreadResponse
	LastTradePriceResponse struct {
		Price string `json:"price"`
	}
	LastTradesPricesResponse []LastTradePriceResponse
	TickSizeResponse         struct {
		MinimumTickSize string `json:"minimum_tick_size,omitempty"`
		TickSize        string `json:"tick_size,omitempty"`
	}
	NegRiskResponse struct {
		NegRisk bool `json:"neg_risk"`
	}
	FeeRateResponse struct {
		BaseFee int    `json:"base_fee,omitempty"`
		FeeRate string `json:"fee_rate,omitempty"`
	}
	GeoblockResponse struct {
		Blocked bool   `json:"blocked"`
		IP      string `json:"ip"`
		Country string `json:"country"`
		Region  string `json:"region"`
	}
	PricesHistoryResponse []PriceHistoryPoint
	OrderResponse         struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	PostOrdersResponse []OrderResponse
	OrdersResponse     struct {
		Data       []OrderResponse `json:"data"`
		NextCursor string          `json:"next_cursor"`
		Limit      int             `json:"limit"`
		Count      int             `json:"count"`
	}
	CancelResponse struct {
		Status string `json:"status"`
	}
	CancelAllResponse struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	CancelMarketOrdersResponse struct {
		Status string `json:"status"`
	}
	TradesResponse struct {
		Data       []Trade `json:"data"`
		NextCursor string  `json:"next_cursor"`
		Limit      int     `json:"limit"`
		Count      int     `json:"count"`
	}
	OrderScoringResponse struct {
		Scoring bool   `json:"scoring"`
		Score   string `json:"score,omitempty"`
	}
	OrdersScoringResponse    map[string]bool
	BalanceAllowanceResponse struct {
		Balance    string            `json:"balance"`
		Allowances map[string]string `json:"allowances,omitempty"`
		// Allowance is deprecated; prefer Allowances.
		Allowance string `json:"allowance,omitempty"`
	}
	NotificationsResponse     []Notification
	DropNotificationsResponse struct {
		Status string `json:"status"`
	}
	UserEarningsResponse struct {
		Data       []UserEarning `json:"data"`
		NextCursor string        `json:"next_cursor"`
		Limit      int           `json:"limit"`
		Count      int           `json:"count"`
	}
	UserTotalEarningsResponse     []TotalUserEarning
	UserRewardPercentagesResponse struct {
		Percentages map[string]string `json:"percentages"`
	}
	RewardsMarketsResponse struct {
		Data       []CurrentReward `json:"data"`
		NextCursor string          `json:"next_cursor"`
		Limit      int             `json:"limit"`
		Count      int             `json:"count"`
	}
	RewardsMarketResponse struct {
		Data       []MarketReward `json:"data"`
		NextCursor string         `json:"next_cursor"`
		Limit      int            `json:"limit"`
		Count      int            `json:"count"`
	}
	UserRewardsByMarketResponse []UserRewardsEarning
	MarketTradesEventsResponse  []TradeEvent
	APIKeyResponse              struct {
		APIKey     string `json:"apiKey"`
		Secret     string `json:"secret,omitempty"`
		Passphrase string `json:"passphrase,omitempty"`
	}
	APIKeyListResponse struct {
		APIKeys []APIKeyResponse `json:"apiKeys"`
	}
	ClosedOnlyResponse struct {
		ClosedOnly bool `json:"closed_only"`
	}
	ValidateReadonlyAPIKeyResponse struct {
		Valid bool `json:"valid"`
	}
	BuilderTradesResponse struct {
		Data       []Trade `json:"data"`
		NextCursor string  `json:"next_cursor"`
		Limit      int     `json:"limit"`
		Count      int     `json:"count"`
	}
)

// Auxiliary types.
type (
	Market struct {
		ID          string        `json:"id"`
		Question    string        `json:"question"`
		ConditionID string        `json:"condition_id"`
		Slug        string        `json:"slug"`
		Resolution  string        `json:"resolution"`
		EndDate     string        `json:"end_date"`
		Tokens      []MarketToken `json:"tokens"`
		// Add minimal fields to match "Simplified" or "Active"
		Active bool `json:"active"`
		Closed bool `json:"closed"`
	}

	MarketToken struct {
		TokenID string  `json:"token_id"`
		Outcome string  `json:"outcome"`
		Price   float64 `json:"price"`
	}

	OrderBook struct {
		MarketID string       `json:"market_id"`
		Bids     []PriceLevel `json:"bids"`
		Asks     []PriceLevel `json:"asks"`
		Hash     string       `json:"hash"`
	}

	PriceLevel struct {
		Price string `json:"price"`
		Size  string `json:"size"`
	}

	Order struct {
		// Define order fields
		Salt          types.U256    `json:"salt"`
		Signer        types.Address `json:"signer"`
		Maker         types.Address `json:"maker"`
		Taker         types.Address `json:"taker"`
		TokenID       types.U256    `json:"token_id"`
		MakerAmount   types.Decimal `json:"maker_amount"`
		TakerAmount   types.Decimal `json:"taker_amount"`
		Expiration    types.U256    `json:"expiration"`
		Side          string        `json:"side"` // BUY/SELL
		FeeRateBps    types.Decimal `json:"fee_rate_bps"`
		Nonce         types.U256    `json:"nonce"`
		SignatureType *int          `json:"signature_type,omitempty"` // 0=EOA, 1=Proxy, 2=Safe
	}

	PriceHistoryPoint struct {
		Timestamp int64  `json:"t"`
		Price     string `json:"p"`
	}

	Trade struct {
		ID        string `json:"id"`
		Price     string `json:"price"`
		Size      string `json:"size"`
		Side      string `json:"side"`
		Timestamp int64  `json:"timestamp"`
	}

	Notification struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	RewardToken struct {
		TokenID string `json:"token_id"`
		Outcome string `json:"outcome"`
		Price   string `json:"price"`
		Winner  bool   `json:"winner,omitempty"`
	}

	RewardsConfig struct {
		AssetAddress string `json:"asset_address"`
		StartDate    string `json:"start_date"`
		EndDate      string `json:"end_date"`
		RatePerDay   string `json:"rate_per_day"`
		TotalRewards string `json:"total_rewards"`
	}

	MarketRewardsConfig struct {
		ID           string `json:"id"`
		AssetAddress string `json:"asset_address"`
		StartDate    string `json:"start_date"`
		EndDate      string `json:"end_date"`
		RatePerDay   string `json:"rate_per_day"`
		TotalRewards string `json:"total_rewards"`
		TotalDays    string `json:"total_days"`
	}

	Earning struct {
		AssetAddress string `json:"asset_address"`
		Earnings     string `json:"earnings"`
		AssetRate    string `json:"asset_rate"`
	}

	UserEarning struct {
		Date         string `json:"date"`
		ConditionID  string `json:"condition_id"`
		AssetAddress string `json:"asset_address"`
		MakerAddress string `json:"maker_address"`
		Earnings     string `json:"earnings"`
		AssetRate    string `json:"asset_rate"`
	}

	TotalUserEarning struct {
		Date         string `json:"date"`
		AssetAddress string `json:"asset_address"`
		MakerAddress string `json:"maker_address"`
		Earnings     string `json:"earnings"`
		AssetRate    string `json:"asset_rate"`
	}

	UserRewardsEarning struct {
		ConditionID           string          `json:"condition_id"`
		Question              string          `json:"question"`
		MarketSlug            string          `json:"market_slug"`
		EventSlug             string          `json:"event_slug"`
		Image                 string          `json:"image"`
		RewardsMaxSpread      string          `json:"rewards_max_spread"`
		RewardsMinSize        string          `json:"rewards_min_size"`
		MarketCompetitiveness string          `json:"market_competitiveness"`
		Tokens                []RewardToken   `json:"tokens,omitempty"`
		RewardsConfig         []RewardsConfig `json:"rewards_config,omitempty"`
		MakerAddress          string          `json:"maker_address"`
		EarningPercentage     string          `json:"earning_percentage"`
		Earnings              []Earning       `json:"earnings,omitempty"`
	}

	CurrentReward struct {
		ConditionID      string          `json:"condition_id"`
		RewardsConfig    []RewardsConfig `json:"rewards_config,omitempty"`
		RewardsMaxSpread string          `json:"rewards_max_spread"`
		RewardsMinSize   string          `json:"rewards_min_size"`
	}

	MarketReward struct {
		ConditionID           string                `json:"condition_id"`
		Question              string                `json:"question"`
		MarketSlug            string                `json:"market_slug"`
		EventSlug             string                `json:"event_slug"`
		Image                 string                `json:"image"`
		RewardsMaxSpread      string                `json:"rewards_max_spread"`
		RewardsMinSize        string                `json:"rewards_min_size"`
		MarketCompetitiveness string                `json:"market_competitiveness"`
		Tokens                []RewardToken         `json:"tokens,omitempty"`
		RewardsConfig         []MarketRewardsConfig `json:"rewards_config,omitempty"`
	}

	TradeEvent struct {
		// ...
	}

	APIKeyInfo struct {
		APIKey string `json:"apiKey"`
		Type   string `json:"type"`
	}
)

// PricesHistoryResponse supports both legacy array responses and the current
// object-wrapped form returned by the API (e.g. {"history":[...]}).
func (p *PricesHistoryResponse) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*p = nil
		return nil
	}

	var points []PriceHistoryPoint
	if err := json.Unmarshal(trimmed, &points); err == nil {
		*p = points
		return nil
	}

	var wrapper struct {
		History []PriceHistoryPoint `json:"history"`
		Data    []PriceHistoryPoint `json:"data"`
	}
	if err := json.Unmarshal(trimmed, &wrapper); err != nil {
		return err
	}
	if wrapper.History != nil {
		*p = wrapper.History
		return nil
	}
	if wrapper.Data != nil {
		*p = wrapper.Data
		return nil
	}
	*p = nil
	return nil
}
