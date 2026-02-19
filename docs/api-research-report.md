# Polymarket API Research Report

## Date: 2026-02-19
## SDK: github.com/GoPolymarket/polymarket-go-sdk

---

## 1. API Overview

Polymarket exposes three main APIs:

| API | Base URL | Auth Required | Purpose |
|-----|----------|---------------|---------|
| **CLOB API** | `https://clob.polymarket.com` | Mixed (public market data + authenticated trading) | Order book, pricing, trading, account management |
| **Gamma API** | `https://gamma-api.polymarket.com` | No | Market/event discovery, metadata, search, tags |
| **Data API** | `https://data-api.polymarket.com` | No | Positions, trades, analytics, leaderboards |

### Authentication

**L1 (EIP-712 Signing)** - Used for API key creation/derivation:
- Headers: `POLY_ADDRESS`, `POLY_SIGNATURE`, `POLY_TIMESTAMP`, `POLY_NONCE`
- Signs a `ClobAuth` typed data message

**L2 (HMAC-SHA256)** - Used for all authenticated trading operations:
- Headers: `POLY_ADDRESS`, `POLY_API_KEY`, `POLY_PASSPHRASE`, `POLY_TIMESTAMP`, `POLY_SIGNATURE`
- Message = `{timestamp}{method}{path}{body}`

**Builder Attribution** - Additional headers for builder-attributed trades:
- Headers: `POLY_BUILDER_API_KEY`, `POLY_BUILDER_PASSPHRASE`, `POLY_BUILDER_TIMESTAMP`, `POLY_BUILDER_SIGNATURE`

**Signature Types:**
| Type | Value | Description |
|------|-------|-------------|
| EOA | 0 | Standard Externally Owned Account |
| POLY_PROXY | 1 | Polymarket Proxy / Magic Link wallet |
| GNOSIS_SAFE | 2 | Gnosis Safe multisig |

### Pagination

- **CLOB API**: Cursor-based (`next_cursor` param; `"MA=="` = initial, `"-1"` or `"LTE="` = end)
- **Gamma API**: Offset-based (`limit` + `offset` params)
- **Data API**: Offset-based (`limit` + `offset` params)

---

## 2. CLOB REST Endpoints

### 2.1 System Status

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| GET | `/` | No | `Health()` | Returns health status string |
| GET | `/time` | No | `Time()` | Returns Unix timestamp |

### 2.2 Market Data (Public)

| Method | Path | Auth | SDK Method | Parameters |
|--------|------|------|------------|------------|
| GET | `/markets` | No | `Markets()` | `limit`, `cursor`, `active`, `asset_id` |
| GET | `/markets/{id}` | No | `Market()` | Path param: market condition ID |
| GET | `/simplified-markets` | No | `SimplifiedMarkets()` | Same as `/markets` |
| GET | `/sampling-markets` | No | `SamplingMarkets()` | None |
| GET | `/sampling-simplified-markets` | No | `SamplingSimplifiedMarkets()` | None |

### 2.3 Order Book & Pricing (Public)

| Method | Path | Auth | SDK Method | Parameters |
|--------|------|------|------------|------------|
| GET | `/book` | No | `OrderBook()` | `token_id`, `side` |
| POST | `/books` | No | `OrderBooks()` | Body: array of `{token_id, side}` |
| GET | `/midpoint` | No | `Midpoint()` | `token_id` |
| POST | `/midpoints` | No | `Midpoints()` | Body: array of `{token_id}` |
| GET | `/price` | No | `Price()` | `token_id`, `side` |
| GET | `/prices` | No | `AllPrices()` | None (all active tokens) |
| POST | `/prices` | No | `Prices()` | Body: array of `{token_id, side}` |
| GET | `/spread` | No | `Spread()` | `token_id`, `side` |
| POST | `/spreads` | No | `Spreads()` | Body: array of `{token_id, side}` |
| GET | `/last-trade-price` | No | `LastTradePrice()` | `token_id` |
| GET | `/last-trades-prices` | No | - | Query param: up to 500 `token_id` |
| POST | `/last-trades-prices` | No | `LastTradesPrices()` | Body: array of `{token_id}` |
| GET | `/tick-size` | No | `TickSize()` | `token_id` |
| GET | `/tick-size/{token_id}` | No | - | Path param variant |
| GET | `/neg-risk` | No | `NegRisk()` | `token_id` |
| GET | `/fee-rate` | No | `FeeRate()` | `token_id` |
| GET | `/fee-rate/{token_id}` | No | - | Path param variant |
| GET | `/prices-history` | No | `PricesHistory()` | `market`/`token_id`, `interval`, `start_ts`, `end_ts`, `fidelity` |

### 2.4 Trading (Authenticated - L2)

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| POST | `/order` | L2 | `PostOrder()` | Submit single signed order |
| POST | `/orders` | L2 | `PostOrders()` | Submit up to 15 signed orders |
| DELETE | `/order` | L2 | `CancelOrder()` | Body: `{"orderId": "..."}` |
| DELETE | `/orders` | L2 | `CancelOrders()` | Body: array of order IDs (max 3000) |
| DELETE | `/cancel-all` | L2 | `CancelAll()` | Cancel all open orders |
| DELETE | `/cancel-market-orders` | L2 | `CancelMarketOrders()` | Body: `{"market": "...", "asset_id": "..."}` |

### 2.5 Order & Trade Queries (Authenticated - L2)

| Method | Path | Auth | SDK Method | Parameters |
|--------|------|------|------------|------------|
| GET | `/data/order/{orderID}` | L2 | `Order()` | Path param: order ID |
| GET | `/data/orders` | L2 | `Orders()` | `id`, `market`, `asset_id`, `limit`, `next_cursor` |
| GET | `/data/trades` | L2 | `Trades()` | `id`, `taker`, `maker`, `market`, `asset_id`, `before`, `after`, `limit`, `next_cursor` |
| GET | `/builder/trades` | L2 (Builder) | `BuilderTrades()` | Same as trades + builder key |

### 2.6 Account (Authenticated)

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| GET | `/balance-allowance` | L2 | `BalanceAllowance()` | `asset_type`, `token_id`, `signature_type` |
| GET | `/balance-allowance/update` | L2 | `UpdateBalanceAllowance()` | Internal use |
| GET | `/notifications` | L2 | `Notifications()` | `limit` |
| DELETE | `/notifications` | L2 | `DropNotifications()` | `id` (comma-separated) |

### 2.7 Rewards (Authenticated)

| Method | Path | Auth | SDK Method | Parameters |
|--------|------|------|------------|------------|
| GET | `/rewards/user` | L2 | `UserEarnings()` | `date`, `signature_type`, `next_cursor` |
| GET | `/rewards/user/total` | L2 | `UserTotalEarnings()` | `date`, `signature_type` |
| GET | `/rewards/user/percentages` | L2 | `UserRewardPercentages()` | None |
| GET | `/rewards/user/by-market` | L2 | `UserRewardsByMarket()` | `date`, `order_by`, `position`, `no_competition`, `signature_type`, `next_cursor` |
| GET | `/rewards/markets/current` | L2 | `RewardsMarketsCurrent()` | `next_cursor` |
| GET | `/rewards/markets/{id}` | L2 | `RewardsMarkets()` | Path param: market ID |

### 2.8 Scoring

| Method | Path | Auth | SDK Method | Parameters |
|--------|------|------|------------|------------|
| GET | `/order-scoring` | L2 | `OrderScoring()` | `order_id` |
| POST | `/orders-scoring` | L2 | `OrdersScoring()` | Body: array of order IDs |

### 2.9 API Key Management

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| POST | `/auth/api-key` | L1 | `CreateAPIKey()` | Creates new L2 credentials |
| GET | `/auth/derive-api-key` | L1 | `DeriveAPIKey()` | Derives existing L2 credentials |
| GET | `/auth/api-keys` | L2 | `ListAPIKeys()` | Lists all active keys |
| DELETE | `/auth/api-key` | L2 | `DeleteAPIKey()` | `api_key` query param |
| GET | `/auth/ban-status/closed-only` | L2 | `ClosedOnlyStatus()` | Check close-only status |
| POST | `/auth/readonly-api-key` | L2 | `CreateReadonlyAPIKey()` | |
| GET | `/auth/readonly-api-keys` | L2 | `ListReadonlyAPIKeys()` | |
| DELETE | `/auth/readonly-api-key` | L2 | `DeleteReadonlyAPIKey()` | Body: `{"key": "..."}` |
| GET | `/auth/validate-readonly-api-key` | No | `ValidateReadonlyAPIKey()` | `address`, `key` |
| POST | `/auth/builder-api-key` | L2 | `CreateBuilderAPIKey()` | |
| GET | `/auth/builder-api-key` | L2 | `ListBuilderAPIKeys()` | |
| DELETE | `/auth/builder-api-key` | L2 | `RevokeBuilderAPIKey()` | |

### 2.10 RFQ (Request For Quote)

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| POST | `/rfq/request` | L2 | `CreateRFQRequest()` | Create RFQ |
| DELETE | `/rfq/request` | L2 | `CancelRFQRequest()` | Cancel RFQ |
| GET | `/rfq/data/requests` | L2 | `RFQRequests()` | Query RFQ requests |
| POST | `/rfq/quote` | L2 | `CreateRFQQuote()` | Submit quote |
| DELETE | `/rfq/quote` | L2 | `CancelRFQQuote()` | Cancel quote |
| GET | `/rfq/data/quotes` | L2 | `RFQQuotes()` | Query quotes |
| GET | `/rfq/data/best-quote` | L2 | `RFQBestQuote()` | Get best quote |
| POST | `/rfq/request/accept` | L2 | `RFQRequestAccept()` | Accept request |
| POST | `/rfq/quote/approve` | L2 | `RFQQuoteApprove()` | Approve quote |
| GET | `/rfq/config` | No | `RFQConfig()` | Get RFQ config |

### 2.11 Heartbeat

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| POST | `/heartbeats` | L2 | `Heartbeat()` | Keep session alive |

### 2.12 Market Trades Events

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| GET | `/v1/market-trades-events/{id}` | No | `MarketTradesEvents()` | Recent trade events |

### 2.13 Geoblock

| Method | Path | Auth | SDK Method | Notes |
|--------|------|------|------------|-------|
| GET | `https://polymarket.com/api/geoblock` | No | `Geoblock()` | Uses different host |

---

## 3. WebSocket Channels and Events

### 3.1 Connection Details

| Channel | URL | Auth Required |
|---------|-----|---------------|
| Market Data | `wss://ws-subscriptions-clob.polymarket.com/ws/market` | No |
| User Events | `wss://ws-subscriptions-clob.polymarket.com/ws/user` | Yes (API key in subscription message) |

### 3.2 Subscription Protocol

**Market subscription:**
```json
{
  "type": "market",
  "operation": "subscribe",
  "assets_ids": ["token_id_1", "token_id_2"],
  "initial_dump": true
}
```

**User subscription:**
```json
{
  "type": "subscribe",
  "operation": "subscribe",
  "markets": ["condition_id_1"],
  "initial_dump": true,
  "auth": {
    "apiKey": "...",
    "secret": "...",
    "passphrase": "..."
  }
}
```

**Heartbeat:** Send `"PING"` text message, receive `"PONG"` response.

### 3.3 Market Data Events

| Event Type | SDK Type | Description | Key Fields |
|------------|----------|-------------|------------|
| `book` / `orderbook` | `OrderbookEvent` | L2 order book snapshot | `asset_id`, `bids`, `asks`, `hash`, `timestamp` |
| `price` / `price_change` | `PriceEvent` / `PriceChangeEvent` | Price updates | `market`, `price_changes[]` with `asset_id`, `best_ask`, `best_bid`, `price`, `side`, `size` |
| `midpoint` | `MidpointEvent` | Mid-price update | `asset_id`, `midpoint` |
| `last_trade_price` | `LastTradePriceEvent` | Last trade price | `asset_id`, `price`, `side`, `size`, `fee_rate_bps`, `timestamp` |
| `tick_size_change` | `TickSizeChangeEvent` | Tick size change | `asset_id`, `tick_size`, `minimum_tick_size` |
| `best_bid_ask` | `BestBidAskEvent` | Top-of-book update | `asset_id`, `best_bid`, `best_ask`, `spread` |
| `new_market` | `NewMarketEvent` | New market created | `id`, `question`, `assets_ids`, `outcomes` |
| `market_resolved` | `MarketResolvedEvent` | Market resolved | `id`, `winning_asset_id`, `winning_outcome` |

### 3.4 User Events (Authenticated)

| Event Type | SDK Type | Description | Key Fields |
|------------|----------|-------------|------------|
| `trade` | `TradeEvent` | User trade execution | `asset_id`, `price`, `size`, `side`, `timestamp`, `status` |
| `order` | `OrderEvent` | Order status update | `id`, `asset_id`, `side`, `price`, `original_size`, `size_matched`, `status`, `order_type` |

### 3.5 Order Event Statuses
- `LIVE` - resting on order book
- `CANCELED` - cancelled
- `MATCHED` - filled

### 3.6 Trade Event Statuses
- `MATCHED` -> `MINED` -> `CONFIRMED` (success path)
- `RETRYING` -> `FAILED` (failure path)

---

## 4. Gamma API Endpoints

### 4.1 Events

| Method | Path | SDK Method | Parameters |
|--------|------|------------|------------|
| GET | `/events` | `Events()` | `limit`, `offset`, `order`, `ascending`, `id[]`, `tag_id`, `exclude_tag_id[]`, `slug[]`, `tag_slug`, `related_tags`, `active`, `archived`, `featured`, `cyom`, `closed`, `liquidity_min/max`, `volume_min/max`, `start_date_min/max`, `end_date_min/max` |
| GET | `/events/{id}` | `EventByID()` | `include_chat`, `include_template` |
| GET | `/events/slug/{slug}` | `EventBySlug()` | `include_chat`, `include_template` |
| GET | `/events/{id}/tags` | `EventTags()` | None |

### 4.2 Markets

| Method | Path | SDK Method | Parameters |
|--------|------|------------|------------|
| GET | `/markets` | `Markets()` | `limit`, `offset`, `order`, `ascending`, `slug`, `slug[]`, `id[]`, `clob_token_ids[]`, `condition_ids[]`, `market_maker_address[]`, `active`, `closed`, `tag_id`, `tag_slug`, `related_tags`, `cyom`, `uma_resolution_status`, `game_id`, `sports_market_types[]`, `volume_min/max`, `liquidity_min/max`, `liquidity_num_min/max`, `volume_num_min/max`, `start_date_min/max`, `end_date_min/max`, `rewards_min_size`, `rewards_max_size` |
| GET | `/markets/{id}` | `MarketByID()` | `include_tag` |
| GET | `/markets/slug/{slug}` | `MarketBySlug()` | `include_tag` |
| GET | `/markets/{id}/tags` | `MarketTags()` | None |

### 4.3 Tags

| Method | Path | SDK Method |
|--------|------|------------|
| GET | `/tags` | `Tags()` |
| GET | `/tags/{id}` | `TagByID()` |
| GET | `/tags/slug/{slug}` | `TagBySlug()` |
| GET | `/tags/{id}/related-tags` | `RelatedTagsByID()` |
| GET | `/tags/slug/{slug}/related-tags` | `RelatedTagsBySlug()` |
| GET | `/tags/{id}/related-tags/tags` | `TagsRelatedToTagByID()` |
| GET | `/tags/slug/{slug}/related-tags/tags` | `TagsRelatedToTagBySlug()` |

### 4.4 Other

| Method | Path | SDK Method |
|--------|------|------------|
| GET | `/series` | `Series()` |
| GET | `/series/{id}` | `SeriesByID()` |
| GET | `/teams` | `Teams()` |
| GET | `/sports` | `Sports()` |
| GET | `/sports/market-types` | `SportsMarketTypes()` |
| GET | `/comments` | `Comments()` |
| GET | `/comments/{id}` | `CommentByID()` |
| GET | `/comments/user_address/{address}` | `CommentsByUserAddress()` |
| GET | `/public-profile` | `PublicProfile()` |
| GET | `/public-search` | `PublicSearch()` |
| GET | `/status` | `Status()` |

---

## 5. Data API Endpoints

| Method | Path | SDK Method | Key Parameters |
|--------|------|------------|----------------|
| GET | `/` or `/` | `Health()` | None |
| GET | `/positions` | `Positions()` | `user` (required), `market`, `eventId`, `sizeThreshold`, `redeemable`, `mergeable`, `limit` (0-500), `offset` (0-10000), `sortBy`, `sortDirection`, `title` |
| GET | `/closed-positions` | `ClosedPositions()` | `user` (required), `market`, `title`, `limit` (0-50), `offset`, `sortBy`, `sortDirection` |
| GET | `/trades` | `Trades()` | `user`, `market`, `eventId`, `limit` (0-10000), `offset`, `takerOnly`, `filterType`, `filterAmount`, `side` |
| GET | `/activity` | `Activity()` | `user` (required), `market`, `type[]`, `limit` (0-500), `offset`, `start`, `end`, `sortBy`, `sortDirection`, `side` |
| GET | `/holders` | `Holders()` | `market[]`, `limit` (0-20), `minBalance` |
| GET | `/value` | `Value()` | `user`, `market[]` |
| GET | `/traded` | `Traded()` | `user` |
| GET | `/oi` | `OpenInterest()` | `market[]` |
| GET | `/live-volume` | `LiveVolume()` | `id` (event ID) |
| GET | `/v1/leaderboard` | `Leaderboard()` | `category`, `timePeriod`, `orderBy`, `limit` (1-50), `offset` (0-1000), `user`, `userName` |
| GET | `/v1/builders/leaderboard` | `BuildersLeaderboard()` | `timePeriod`, `limit`, `offset` |
| GET | `/v1/builders/volume` | `BuildersVolume()` | `timePeriod` |

---

## 6. RTDS (Real-Time Data Service) WebSocket

The RTDS service provides additional real-time streams beyond the CLOB WebSocket:

| Stream Type | SDK Method | Description |
|-------------|------------|-------------|
| `crypto_prices` | `SubscribeCryptoPrices()` | Binance crypto price feed |
| `crypto_prices_chainlink` | `SubscribeChainlinkPrices()` | Chainlink oracle prices |
| `comments` | `SubscribeComments()` | Comment stream (supports auth) |
| `activity` | `SubscribeOrdersMatched()` | Matched order activity |

---

## 7. Key Data Types

### Order Types (Time-in-Force)
| Type | Description | SDK Constant |
|------|-------------|-------------|
| GTC | Good Till Cancelled | `OrderTypeGTC` |
| GTD | Good Till Date (requires expiration) | `OrderTypeGTD` |
| FOK | Fill Or Kill (complete fill or reject) | `OrderTypeFOK` |
| FAK | Fill And Kill (partial fill allowed, remainder cancelled) | `OrderTypeFAK` |

### Price History Intervals
| Interval | Description | SDK Constant |
|----------|-------------|-------------|
| `1m` | 1 minute | `PriceHistoryInterval1m` |
| `1h` | 1 hour | `PriceHistoryInterval1h` |
| `6h` | 6 hours | `PriceHistoryInterval6h` |
| `1d` | 1 day | `PriceHistoryInterval1d` |
| `1w` | 1 week | `PriceHistoryInterval1w` |
| `max` | Maximum available | `PriceHistoryIntervalMax` |

### EIP-712 Order Structure
```
Order {
  salt: uint256
  maker: address
  signer: address
  taker: address
  tokenId: uint256
  makerAmount: uint256
  takerAmount: uint256
  expiration: uint256
  nonce: uint256
  feeRateBps: uint256
  side: uint8        // 0=BUY, 1=SELL
  signatureType: uint8  // 0=EOA, 1=Proxy, 2=Safe
}
```

### Exchange Contract
- Mainnet: `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E`

### Tick Sizes
Prices must be multiples of the tick size (e.g., 0.01 for most markets, 0.001 for higher-precision markets). Valid price range: `[0.00, 1.00]`.

---

## 8. Mismatches Found (SDK vs API Docs)

### 8.1 Market Type Field Mismatch (CLOB vs Gamma)

The **CLOB** `Market` type (`pkg/clob/clobtypes/types.go:409`) uses `snake_case` JSON tags:
```go
type Market struct {
    ConditionID string `json:"condition_id"`
    EndDate     string `json:"end_date"`
}
```

The **Gamma** `Market` type (`pkg/gamma/types.go:206`) uses `camelCase` JSON tags:
```go
type Market struct {
    ConditionID string `json:"conditionId"`
    EndDate     string `json:"endDate"`
}
```

This is **correct** - the CLOB and Gamma APIs actually use different JSON naming conventions. The SDK correctly handles both.

### 8.2 OrderResponse Type is Incomplete

The `OrderResponse` struct (`pkg/clob/clobtypes/types.go:319`) only has:
```go
type OrderResponse struct {
    ID     string `json:"orderID"`
    Status string `json:"status"`
}
```

The actual API response for `GET /data/order/{id}` and `GET /data/orders` returns many more fields including: `asset_id`, `market`, `side`, `price`, `original_size`, `size_matched`, `owner`, `maker_address`, `order_type`, `expiration`, `created_at`, `timestamp`, etc.

**Impact:** Users cannot access full order details from the SDK response.

### 8.3 Trade Type is Minimal

The CLOB `Trade` struct (`pkg/clob/clobtypes/types.go:461`) only has:
```go
type Trade struct {
    ID        string `json:"id"`
    Price     string `json:"price"`
    Size      string `json:"size"`
    Side      string `json:"side"`
    Timestamp int64  `json:"timestamp"`
}
```

The API returns additional fields: `taker_order_id`, `maker_order_id`, `market`, `asset_id`, `owner`, `maker_address`, `match_time`, `status`, `fee_rate_bps`, `transaction_hash`, etc.

**Impact:** Missing critical fields like `market`, `asset_id`, `status`, `transaction_hash`.

### 8.4 Gamma Market Type Missing Fields

The Gamma `Market` struct is missing several fields from the API:
- `negRisk` (boolean) - critical for negative risk market handling
- `negRiskMarketId` (string)
- `negRiskRequestId` (string)
- `enableOrderBook` (boolean) - whether CLOB trading is available
- `questionId` (string) - resolution hash
- `volume24hr` (string)
- `spread` (string)
- `bestBid` / `bestAsk` (string)
- `lastTradePrice` (string)
- `commentCount` (int)
- `cyom` (boolean)

### 8.5 Event Type Missing Fields

The Gamma `Event` struct is missing:
- `negRisk` (boolean)
- `enableNegRisk` (boolean)
- `negRiskAugmented` (boolean)
- `commentCount` (int)
- `competitionState` (string)
- `cyom` (boolean)

### 8.6 CancelOrders Body Format

The `CancelOrders()` implementation sends the order IDs array directly as the body:
```go
body = ids  // sends: ["id1", "id2"]
```

But the API docs show the expected format is:
```json
["id1", "id2"]
```

This appears correct, but note the API docs also mention a max of 3000 order IDs per request - the SDK does not validate this limit.

### 8.7 SubscriptionRequest `assets_ids` Naming

In `pkg/clob/ws/types.go:67`:
```go
AssetIDs []string `json:"assets_ids,omitempty"`
```

The field name `assets_ids` (with trailing 's' on 'assets') matches what the WS API expects. However, the `NewMarketEvent` and `MarketResolvedEvent` structs also use `assets_ids`, while the WS server sometimes sends `asset_ids` (without the trailing 's'). The SDK handles this with a fallback (`AssetIDsAlt` field) in `processEvent()`.

### 8.8 Tick Size and Fee Rate Path Param Variants

The API supports both query param and path param variants:
- `GET /tick-size?token_id=X` and `GET /tick-size/X`
- `GET /fee-rate?token_id=X` and `GET /fee-rate/X`

The SDK only uses the query param variant. Not a bug, but the path param variant could be more efficient.

### 8.9 `GET /last-trades-prices` Query Param Support Missing

The API supports `GET /last-trades-prices?token_id=X&token_id=Y` (up to 500 tokens via query params). The SDK only implements the `POST` variant. The `GET` variant is missing from the SDK.

### 8.10 Prices History Response Format

The API returns `{"history": [...]}` but the SDK's `PricesHistoryResponse` has a custom `UnmarshalJSON` that handles both the array form and the wrapper form. This is correctly defensive.

---

## 9. Missing Features in SDK

### 9.1 Missing CLOB Endpoints
1. **`GET /last-trades-prices` (query param version)** - Batch last trade prices via GET
2. **`GET /tick-size/{token_id}` (path param version)** - Direct tick size lookup
3. **`GET /fee-rate/{token_id}` (path param version)** - Direct fee rate lookup

### 9.2 Incomplete Response Types
1. **`OrderResponse`** - Missing most fields (asset_id, market, side, price, original_size, size_matched, owner, maker_address, order_type, expiration, created_at, etc.)
2. **`Trade` (CLOB)** - Missing market, asset_id, status, transaction_hash, fee_rate_bps, etc.
3. **`Market` (Gamma)** - Missing negRisk, enableOrderBook, questionId, volume24hr, spread, bestBid/bestAsk, lastTradePrice, etc.
4. **`Event` (Gamma)** - Missing negRisk, enableNegRisk, negRiskAugmented, commentCount, etc.
5. **`TradeEvent` (WS)** - The struct has a placeholder `type TradeEvent struct { // ... }` in `clobtypes` - it's empty.

### 9.3 Missing Validation
1. No max batch size validation for `POST /orders` (limit: 15)
2. No max batch size validation for `DELETE /orders` (limit: 3000)
3. No max token count validation for `GET /last-trades-prices` (limit: 500)
4. No rate limit awareness or retry-after handling

### 9.4 Missing Data API Endpoints
The Data API docs mention these additional endpoints not in the SDK:
- **No missing endpoints** - The SDK covers: positions, closed-positions, trades, activity, holders, value, traded, oi, live-volume, leaderboard, builders/leaderboard, builders/volume.

### 9.5 Bridge/CTF Operations
The `pkg/bridge` package only implements `Deposit()` and returns `ErrWithdrawUnsupported` for `Withdraw()`. The API docs mention bridge operations (deposits, withdrawals, quotes) that could be expanded.

### 9.6 WebSocket Event Type Mapping Gaps
The WS `processEvent()` handles event types `"order"` but the API sends `"orders"` (plural). The SDK's `UserOrders` const is `"orders"` while the dispatch matches on `"order"`. This could lead to missed order events if the server sends `"orders"` as the event type rather than `"order"`.

---

## 10. Rate Limits

| Endpoint | Limit |
|----------|-------|
| CLOB General | 9,000 req/10s |
| `POST /order` | 3,500 req/10s burst; 36,000 req/10min sustained |
| `DELETE /order` | 3,000 req/10s burst |
| `/book` | 1,500 req/10s |
| Gamma General | 4,000 req/10s |
| `/markets` (Gamma) | 300 req/10s |
| Data API General | 1,000 req/10s |

The SDK has a `transport/ratelimit.go` module but rate limits are not pre-configured per-endpoint.

---

## 11. Contract Addresses

| Contract | Address |
|----------|---------|
| CTF Exchange (Mainnet) | `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E` |
| Proxy Factory | `0xaB45c5A4B0c941a2F231C04C3f49182e1A254052` |
| Safe Factory | `0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b` |
| USDC (Polygon) | `0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174` |
| CTF Contract | `0x4D97DCd97eC945f40cF65F87097ACe5EA0476045` |
| Neg Risk Adapter | `0xC5d563A36AE78145C45a50134d48A1215220f80a` |
| Neg Risk Exchange | `0xC5d563A36AE78145C45a50134d48A1215220f80a` |

---

## 12. Summary of Priority Issues

### High Priority
1. **Incomplete `OrderResponse`** - Users cannot access full order data (missing ~15 fields)
2. **Incomplete CLOB `Trade`** - Missing market, asset_id, status, transaction_hash
3. **Missing Gamma `Market.NegRisk`** - Critical for negative risk market support
4. **WS event type `"order"` vs `"orders"` mismatch** - Could cause missed events

### Medium Priority
5. **Missing Gamma Market fields** - enableOrderBook, questionId, volume24hr, spread, etc.
6. **Missing Gamma Event fields** - negRisk flags, commentCount, etc.
7. **No batch size validation** - orders (15), cancels (3000), last-trades-prices (500)
8. **Empty `TradeEvent` in clobtypes** - Placeholder struct with no fields

### Low Priority
9. **Missing GET `/last-trades-prices` query param variant**
10. **Missing path param variants for tick-size and fee-rate**
11. **Rate limit values not pre-configured**
12. **Bridge Withdraw not implemented**
