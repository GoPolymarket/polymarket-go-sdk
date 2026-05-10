package rfq

// RFQ filters/sort enums.
type RFQState string
type RFQSortBy string
type RFQSortDir string

const (
	RFQStateActive   RFQState = "active"
	RFQStateInactive RFQState = "inactive"
)

const (
	RFQSortByPrice   RFQSortBy = "price"
	RFQSortByExpiry  RFQSortBy = "expiry"
	RFQSortBySize    RFQSortBy = "size"
	RFQSortByCreated RFQSortBy = "created"
)

const (
	RFQSortDirAsc  RFQSortDir = "asc"
	RFQSortDirDesc RFQSortDir = "desc"
)

// Request types.
type RFQRequest struct {
	MarketID string `json:"market_id,omitempty"`
	Side     string `json:"side,omitempty"`
	Size     string `json:"size,omitempty"`

	AssetIn   string `json:"assetIn,omitempty"`
	AssetOut  string `json:"assetOut,omitempty"`
	AmountIn  string `json:"amountIn,omitempty"`
	AmountOut string `json:"amountOut,omitempty"`
	UserType  string `json:"userType,omitempty"`
}

type RFQCancelRequest struct {
	ID        string `json:"id,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

type RFQRequestsQuery struct {
	Limit       int        `json:"limit,omitempty"`
	Cursor      string     `json:"cursor,omitempty"`
	Offset      string     `json:"offset,omitempty"`
	State       RFQState   `json:"state,omitempty"`
	RequestIDs  []string   `json:"requestIds,omitempty"`
	Markets     []string   `json:"markets,omitempty"`
	SizeMin     string     `json:"sizeMin,omitempty"`
	SizeMax     string     `json:"sizeMax,omitempty"`
	SizeUsdcMin string     `json:"sizeUsdcMin,omitempty"`
	SizeUsdcMax string     `json:"sizeUsdcMax,omitempty"`
	PriceMin    string     `json:"priceMin,omitempty"`
	PriceMax    string     `json:"priceMax,omitempty"`
	SortBy      RFQSortBy  `json:"sortBy,omitempty"`
	SortDir     RFQSortDir `json:"sortDir,omitempty"`
}

type RFQQuote struct {
	RequestID string `json:"request_id,omitempty"`
	Price     string `json:"price,omitempty"`

	RequestIDV2 string `json:"requestId,omitempty"`
	AssetIn     string `json:"assetIn,omitempty"`
	AssetOut    string `json:"assetOut,omitempty"`
	AmountIn    string `json:"amountIn,omitempty"`
	AmountOut   string `json:"amountOut,omitempty"`
	UserType    string `json:"userType,omitempty"`
}

type RFQCancelQuote struct {
	ID      string `json:"id,omitempty"`
	QuoteID string `json:"quoteId,omitempty"`
}

type RFQQuotesQuery struct {
	RequestIDs  []string   `json:"requestIds,omitempty"`
	QuoteIDs    []string   `json:"quoteIds,omitempty"`
	Markets     []string   `json:"markets,omitempty"`
	State       RFQState   `json:"state,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Cursor      string     `json:"cursor,omitempty"`
	Offset      string     `json:"offset,omitempty"`
	SizeMin     string     `json:"sizeMin,omitempty"`
	SizeMax     string     `json:"sizeMax,omitempty"`
	SizeUsdcMin string     `json:"sizeUsdcMin,omitempty"`
	SizeUsdcMax string     `json:"sizeUsdcMax,omitempty"`
	PriceMin    string     `json:"priceMin,omitempty"`
	PriceMax    string     `json:"priceMax,omitempty"`
	SortBy      RFQSortBy  `json:"sortBy,omitempty"`
	SortDir     RFQSortDir `json:"sortDir,omitempty"`
}

type RFQBestQuoteQuery struct {
	RequestID  string   `json:"request_id,omitempty"`
	RequestIDs []string `json:"requestIds,omitempty"`
}

type RFQAcceptRequest struct {
	QuoteID     string `json:"quote_id,omitempty"`
	RequestID   string `json:"requestId,omitempty"`
	QuoteIDV2   string `json:"quoteId,omitempty"`
	MakerAmount string `json:"makerAmount,omitempty"`
	TakerAmount string `json:"takerAmount,omitempty"`
	TokenID     string `json:"tokenId,omitempty"`
	Maker       string `json:"maker,omitempty"`
	Signer      string `json:"signer,omitempty"`
	Expiration  string `json:"expiration,omitempty"`
	Side        string `json:"side,omitempty"`
	Signature   string `json:"signature,omitempty"`
	Salt        string `json:"salt,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Timestamp   int64  `json:"timestamp,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
	Builder     string `json:"builder,omitempty"`
}

type RFQApproveQuote struct {
	QuoteID     string `json:"quote_id,omitempty"`
	RequestID   string `json:"requestId,omitempty"`
	QuoteIDV2   string `json:"quoteId,omitempty"`
	MakerAmount string `json:"makerAmount,omitempty"`
	TakerAmount string `json:"takerAmount,omitempty"`
	TokenID     string `json:"tokenId,omitempty"`
	Maker       string `json:"maker,omitempty"`
	Signer      string `json:"signer,omitempty"`
	Expiration  string `json:"expiration,omitempty"`
	Side        string `json:"side,omitempty"`
	Signature   string `json:"signature,omitempty"`
	Salt        string `json:"salt,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Timestamp   int64  `json:"timestamp,omitempty"`
	Metadata    string `json:"metadata,omitempty"`
	Builder     string `json:"builder,omitempty"`
}

// Response types.
type RFQRequestResponse struct {
	ID        string `json:"id,omitempty"`
	RequestID string `json:"requestId,omitempty"`
	Expiry    int64  `json:"expiry,omitempty"`
}

type RFQCancelResponse struct {
	Status string `json:"status"`
}

type RFQRequestsResponse []RFQRequestItem

type RFQQuoteResponse struct {
	ID      string `json:"id,omitempty"`
	QuoteID string `json:"quoteId,omitempty"`
}

type RFQQuotesResponse []RFQQuoteItem
type RFQBestQuoteResponse RFQQuoteItem

type RFQAcceptResponse struct {
	Status   string   `json:"status,omitempty"`
	TradeIDs []string `json:"tradeIds,omitempty"`
}

type RFQApproveResponse struct {
	Status   string   `json:"status,omitempty"`
	TradeIDs []string `json:"tradeIds,omitempty"`
}

type RFQConfigResponse struct {
	MinSize string `json:"min_size"`
}

type RFQRequestItem struct {
	ID           string `json:"id,omitempty"`
	RequestID    string `json:"requestId,omitempty"`
	UserAddress  string `json:"userAddress,omitempty"`
	ProxyAddress string `json:"proxyAddress,omitempty"`
	Condition    string `json:"condition,omitempty"`
	Token        string `json:"token,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Side         string `json:"side,omitempty"`
	SizeIn       string `json:"sizeIn,omitempty"`
	SizeOut      string `json:"sizeOut,omitempty"`
	Price        string `json:"price,omitempty"`
	Expiry       int64  `json:"expiry,omitempty"`
}

type RFQQuoteItem struct {
	ID           string `json:"id,omitempty"`
	QuoteID      string `json:"quoteId,omitempty"`
	RequestID    string `json:"requestId,omitempty"`
	UserAddress  string `json:"userAddress,omitempty"`
	ProxyAddress string `json:"proxyAddress,omitempty"`
	Condition    string `json:"condition,omitempty"`
	Token        string `json:"token,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Side         string `json:"side,omitempty"`
	SizeIn       string `json:"sizeIn,omitempty"`
	SizeOut      string `json:"sizeOut,omitempty"`
	Price        string `json:"price,omitempty"`
}
