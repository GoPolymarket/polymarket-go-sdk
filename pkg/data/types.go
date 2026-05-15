package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/types"

	"github.com/ethereum/go-ethereum/common"
)

// Enum types.
type (
	Side                 string
	ActivityType         string
	PositionSortBy       string
	ClosedPositionSortBy string
	ActivitySortBy       string
	SortDirection        string
	FilterType           string
	TimePeriod           string
	LeaderboardCategory  string
	LeaderboardOrderBy   string
)

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

const (
	ActivityTrade       ActivityType = "TRADE"
	ActivitySplit       ActivityType = "SPLIT"
	ActivityMerge       ActivityType = "MERGE"
	ActivityRedeem      ActivityType = "REDEEM"
	ActivityReward      ActivityType = "REWARD"
	ActivityConversion  ActivityType = "CONVERSION"
	ActivityYield       ActivityType = "YIELD"
	ActivityMakerRebate ActivityType = "MAKER_REBATE"
)

const (
	PositionSortCurrent    PositionSortBy = "CURRENT"
	PositionSortInitial    PositionSortBy = "INITIAL"
	PositionSortTokens     PositionSortBy = "TOKENS"
	PositionSortCashPnl    PositionSortBy = "CASHPNL"
	PositionSortPercentPnl PositionSortBy = "PERCENTPNL"
	PositionSortTitle      PositionSortBy = "TITLE"
	PositionSortResolving  PositionSortBy = "RESOLVING"
	PositionSortPrice      PositionSortBy = "PRICE"
	PositionSortAvgPrice   PositionSortBy = "AVGPRICE"
)

const (
	ClosedPositionSortRealizedPnl ClosedPositionSortBy = "REALIZEDPNL"
	ClosedPositionSortTitle       ClosedPositionSortBy = "TITLE"
	ClosedPositionSortPrice       ClosedPositionSortBy = "PRICE"
	ClosedPositionSortAvgPrice    ClosedPositionSortBy = "AVGPRICE"
	ClosedPositionSortTimestamp   ClosedPositionSortBy = "TIMESTAMP"
)

const (
	ActivitySortTimestamp ActivitySortBy = "TIMESTAMP"
	ActivitySortTokens    ActivitySortBy = "TOKENS"
	ActivitySortCash      ActivitySortBy = "CASH"
)

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

const (
	FilterCash   FilterType = "CASH"
	FilterTokens FilterType = "TOKENS"
)

const (
	TimePeriodDay   TimePeriod = "DAY"
	TimePeriodWeek  TimePeriod = "WEEK"
	TimePeriodMonth TimePeriod = "MONTH"
	TimePeriodAll   TimePeriod = "ALL"
)

const (
	LeaderboardOverall   LeaderboardCategory = "OVERALL"
	LeaderboardPolitics  LeaderboardCategory = "POLITICS"
	LeaderboardSports    LeaderboardCategory = "SPORTS"
	LeaderboardCrypto    LeaderboardCategory = "CRYPTO"
	LeaderboardCulture   LeaderboardCategory = "CULTURE"
	LeaderboardMentions  LeaderboardCategory = "MENTIONS"
	LeaderboardWeather   LeaderboardCategory = "WEATHER"
	LeaderboardEconomics LeaderboardCategory = "ECONOMICS"
	LeaderboardTech      LeaderboardCategory = "TECH"
	LeaderboardFinance   LeaderboardCategory = "FINANCE"
)

const (
	OrderByPnl LeaderboardOrderBy = "PNL"
	OrderByVol LeaderboardOrderBy = "VOL"
)

// MarketFilter restricts queries by markets or events.
type MarketFilter struct {
	Markets  []common.Hash
	EventIDs []int64
}

func MarketFilterByMarkets(markets []common.Hash) *MarketFilter {
	return &MarketFilter{Markets: markets}
}

func MarketFilterByEventIDs(eventIDs []int64) *MarketFilter {
	return &MarketFilter{EventIDs: eventIDs}
}

// TradeFilter restricts trades by minimum value.
type TradeFilter struct {
	FilterType   FilterType
	FilterAmount types.Decimal
}

func NewTradeFilter(filterType FilterType, amount types.Decimal) *TradeFilter {
	return &TradeFilter{FilterType: filterType, FilterAmount: amount}
}

func TradeFilterCash(amount types.Decimal) *TradeFilter {
	return NewTradeFilter(FilterCash, amount)
}

func TradeFilterTokens(amount types.Decimal) *TradeFilter {
	return NewTradeFilter(FilterTokens, amount)
}

// Request types.
type (
	PositionsRequest struct {
		User          common.Address
		Filter        *MarketFilter
		SizeThreshold *types.Decimal
		Redeemable    *bool
		Mergeable     *bool
		Limit         *int
		Offset        *int
		SortBy        *PositionSortBy
		SortDirection *SortDirection
		Title         *string
	}
	TradesRequest struct {
		User        *common.Address
		Filter      *MarketFilter
		Limit       *int
		Offset      *int
		TakerOnly   *bool
		TradeFilter *TradeFilter
		Side        *Side
	}
	ActivityRequest struct {
		User          common.Address
		Filter        *MarketFilter
		ActivityTypes []ActivityType
		Limit         *int
		Offset        *int
		Start         *int64
		End           *int64
		SortBy        *ActivitySortBy
		SortDirection *SortDirection
		Side          *Side
	}
	HoldersRequest struct {
		Markets    []common.Hash
		Limit      *int
		MinBalance *int
	}
	ValueRequest struct {
		User    common.Address
		Markets []common.Hash
	}
	TradedRequest struct {
		User common.Address
	}
	OpenInterestRequest struct {
		Markets []common.Hash
	}
	LiveVolumeRequest struct {
		ID int64
	}
	ClosedPositionsRequest struct {
		User          common.Address
		Filter        *MarketFilter
		Title         *string
		Limit         *int
		Offset        *int
		SortBy        *ClosedPositionSortBy
		SortDirection *SortDirection
	}
	LeaderboardRequest struct {
		Category   *LeaderboardCategory
		TimePeriod *TimePeriod
		OrderBy    *LeaderboardOrderBy
		Limit      *int
		Offset     *int
		User       *common.Address
		UserName   *string
	}
	BuildersLeaderboardRequest struct {
		TimePeriod *TimePeriod
		Limit      *int
		Offset     *int
	}
	BuildersVolumeRequest struct {
		TimePeriod *TimePeriod
	}
)

// BoundedIntError indicates a numeric parameter is outside allowed bounds.
type BoundedIntError struct {
	Value     int
	Min       int
	Max       int
	ParamName string
}

func (e BoundedIntError) Error() string {
	return fmt.Sprintf("%s must be between %d and %d (got %d)", e.ParamName, e.Min, e.Max, e.Value)
}

// FlexibleTime supports both date-only and full RFC3339 timestamps.
type FlexibleTime struct {
	time.Time
}

func (t *FlexibleTime) UnmarshalJSON(data []byte) error {
	data = bytesTrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("invalid time value: %q", s)
}

func (t FlexibleTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Time.Format(time.RFC3339Nano))
}

// Market represents a market condition ID or the global aggregate.
type Market struct {
	Global bool
	ID     common.Hash
	Raw    string
}

func (m *Market) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	m.Raw = s
	if strings.EqualFold(s, "global") {
		m.Global = true
		return nil
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		m.ID = common.HexToHash(s)
		return nil
	}
	if len(s) == 64 {
		m.ID = common.HexToHash("0x" + s)
		return nil
	}
	return nil
}

func (m Market) MarshalJSON() ([]byte, error) {
	if m.Raw != "" {
		return json.Marshal(m.Raw)
	}
	if m.Global {
		return json.Marshal("global")
	}
	return json.Marshal(m.ID.Hex())
}

// IntString parses either a JSON string or number into an int.
type IntString int

func (i *IntString) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*i = IntString(int(val))
	case string:
		if val == "" {
			*i = 0
			return nil
		}
		var parsed int
		if _, err := fmt.Sscanf(val, "%d", &parsed); err != nil {
			return err
		}
		*i = IntString(parsed)
	default:
		return fmt.Errorf("unsupported int value: %v", v)
	}
	return nil
}

// Response types.
type (
	HealthResponse struct {
		Data string `json:"data"`
	}
	Position struct {
		ProxyWallet        common.Address `json:"proxyWallet"`
		Asset              types.U256     `json:"asset"`
		ConditionID        common.Hash    `json:"conditionId"`
		Size               types.Decimal  `json:"size"`
		AvgPrice           types.Decimal  `json:"avgPrice"`
		InitialValue       types.Decimal  `json:"initialValue"`
		CurrentValue       types.Decimal  `json:"currentValue"`
		CashPnl            types.Decimal  `json:"cashPnl"`
		PercentPnl         types.Decimal  `json:"percentPnl"`
		TotalBought        types.Decimal  `json:"totalBought"`
		RealizedPnl        types.Decimal  `json:"realizedPnl"`
		PercentRealizedPnl types.Decimal  `json:"percentRealizedPnl"`
		CurPrice           types.Decimal  `json:"curPrice"`
		Redeemable         bool           `json:"redeemable"`
		Mergeable          bool           `json:"mergeable"`
		Title              string         `json:"title"`
		Slug               string         `json:"slug"`
		Icon               string         `json:"icon"`
		EventSlug          string         `json:"eventSlug"`
		EventID            *string        `json:"eventId,omitempty"`
		Outcome            string         `json:"outcome"`
		OutcomeIndex       int            `json:"outcomeIndex"`
		OppositeOutcome    string         `json:"oppositeOutcome"`
		OppositeAsset      types.U256     `json:"oppositeAsset"`
		EndDate            FlexibleTime   `json:"endDate"`
		NegativeRisk       bool           `json:"negativeRisk"`
	}
	ClosedPosition struct {
		ProxyWallet     common.Address `json:"proxyWallet"`
		Asset           types.U256     `json:"asset"`
		ConditionID     common.Hash    `json:"conditionId"`
		AvgPrice        types.Decimal  `json:"avgPrice"`
		TotalBought     types.Decimal  `json:"totalBought"`
		RealizedPnl     types.Decimal  `json:"realizedPnl"`
		CurPrice        types.Decimal  `json:"curPrice"`
		Timestamp       int64          `json:"timestamp"`
		Title           string         `json:"title"`
		Slug            string         `json:"slug"`
		Icon            string         `json:"icon"`
		EventSlug       string         `json:"eventSlug"`
		Outcome         string         `json:"outcome"`
		OutcomeIndex    int            `json:"outcomeIndex"`
		OppositeOutcome string         `json:"oppositeOutcome"`
		OppositeAsset   types.U256     `json:"oppositeAsset"`
		EndDate         FlexibleTime   `json:"endDate"`
		NegativeRisk    bool           `json:"negativeRisk,omitempty"`
	}
	Trade struct {
		ProxyWallet           common.Address `json:"proxyWallet"`
		Side                  Side           `json:"side"`
		Asset                 types.U256     `json:"asset"`
		ConditionID           common.Hash    `json:"conditionId"`
		Size                  types.Decimal  `json:"size"`
		Price                 types.Decimal  `json:"price"`
		Timestamp             int64          `json:"timestamp"`
		Title                 string         `json:"title"`
		Slug                  string         `json:"slug"`
		Icon                  string         `json:"icon"`
		EventSlug             string         `json:"eventSlug"`
		Outcome               string         `json:"outcome"`
		OutcomeIndex          int            `json:"outcomeIndex"`
		Name                  *string        `json:"name,omitempty"`
		Pseudonym             *string        `json:"pseudonym,omitempty"`
		Bio                   *string        `json:"bio,omitempty"`
		ProfileImage          *string        `json:"profileImage,omitempty"`
		ProfileImageOptimized *string        `json:"profileImageOptimized,omitempty"`
		TransactionHash       common.Hash    `json:"transactionHash"`
		MatchTime             *string        `json:"matchTime,omitempty"`
		Fee                   *types.Decimal `json:"fee,omitempty"`
	}
	Activity struct {
		ProxyWallet           common.Address `json:"proxyWallet"`
		Timestamp             int64          `json:"timestamp"`
		ConditionID           *common.Hash   `json:"conditionId,omitempty"`
		ActivityType          ActivityType   `json:"type"`
		Size                  types.Decimal  `json:"size"`
		USDCSize              types.Decimal  `json:"usdcSize"`
		TransactionHash       common.Hash    `json:"transactionHash"`
		Price                 *types.Decimal `json:"price,omitempty"`
		Asset                 *types.U256    `json:"asset,omitempty"`
		Side                  *Side          `json:"side,omitempty"`
		OutcomeIndex          *int           `json:"outcomeIndex,omitempty"`
		Title                 *string        `json:"title,omitempty"`
		Slug                  *string        `json:"slug,omitempty"`
		Icon                  *string        `json:"icon,omitempty"`
		EventSlug             *string        `json:"eventSlug,omitempty"`
		Outcome               *string        `json:"outcome,omitempty"`
		Name                  *string        `json:"name,omitempty"`
		Pseudonym             *string        `json:"pseudonym,omitempty"`
		Bio                   *string        `json:"bio,omitempty"`
		ProfileImage          *string        `json:"profileImage,omitempty"`
		ProfileImageOptimized *string        `json:"profileImageOptimized,omitempty"`
	}
	Holder struct {
		ProxyWallet           common.Address `json:"proxyWallet"`
		Bio                   *string        `json:"bio,omitempty"`
		Asset                 types.U256     `json:"asset"`
		Pseudonym             *string        `json:"pseudonym,omitempty"`
		Amount                types.Decimal  `json:"amount"`
		DisplayUsernamePublic *bool          `json:"displayUsernamePublic,omitempty"`
		OutcomeIndex          int            `json:"outcomeIndex"`
		Name                  *string        `json:"name,omitempty"`
		ProfileImage          *string        `json:"profileImage,omitempty"`
		ProfileImageOptimized *string        `json:"profileImageOptimized,omitempty"`
		Verified              *bool          `json:"verified,omitempty"`
	}
	MetaHolder struct {
		Token   types.U256 `json:"token"`
		Holders []Holder   `json:"holders"`
	}
	Traded struct {
		User   common.Address `json:"user"`
		Traded int            `json:"traded"`
	}
	Value struct {
		User  common.Address `json:"user"`
		Value types.Decimal  `json:"value"`
	}
	OpenInterest struct {
		Market Market        `json:"market"`
		Value  types.Decimal `json:"value"`
	}
	MarketVolume struct {
		Market Market        `json:"market"`
		Value  types.Decimal `json:"value"`
	}
	LiveVolume struct {
		Total   types.Decimal  `json:"total"`
		Markets []MarketVolume `json:"markets"`
	}
	BuilderLeaderboardEntry struct {
		Rank        IntString     `json:"rank"`
		Builder     string        `json:"builder"`
		Volume      types.Decimal `json:"volume"`
		ActiveUsers int           `json:"activeUsers"`
		Verified    bool          `json:"verified"`
		BuilderLogo *string       `json:"builderLogo,omitempty"`
	}
	BuilderVolumeEntry struct {
		Dt          FlexibleTime  `json:"dt"`
		Builder     string        `json:"builder"`
		BuilderLogo *string       `json:"builderLogo,omitempty"`
		Verified    bool          `json:"verified"`
		Volume      types.Decimal `json:"volume"`
		ActiveUsers int           `json:"activeUsers"`
		Rank        IntString     `json:"rank"`
	}
	TraderLeaderboardEntry struct {
		Rank          IntString      `json:"rank"`
		ProxyWallet   common.Address `json:"proxyWallet"`
		UserName      *string        `json:"userName,omitempty"`
		Vol           types.Decimal  `json:"vol"`
		Pnl           types.Decimal  `json:"pnl"`
		ProfileImage  *string        `json:"profileImage,omitempty"`
		XUsername     *string        `json:"xUsername,omitempty"`
		VerifiedBadge *bool          `json:"verifiedBadge,omitempty"`
	}
)

// Response aliases.
type (
	PositionsResponse           []Position
	TradesResponse              []Trade
	ActivityResponse            []Activity
	HoldersResponse             []MetaHolder
	ValueResponse               []Value
	ClosedPositionsResponse     []ClosedPosition
	TradedResponse              Traded
	OpenInterestResponse        []OpenInterest
	LiveVolumeResponse          []LiveVolume
	LeaderboardResponse         []TraderLeaderboardEntry
	BuildersLeaderboardResponse []BuilderLeaderboardEntry
	BuildersVolumeResponse      []BuilderVolumeEntry
)

func bytesTrimSpace(data []byte) []byte {
	return []byte(strings.TrimSpace(string(data)))
}
