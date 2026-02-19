# Polymarket Go SDK - Code Analysis Report

## Summary

This report covers a deep analysis of the Go SDK codebase. Two packages have confirmed failures: `pkg/clob/ws` (build failure) and `pkg/clob` (test failures). Several additional issues of varying severity were identified across the codebase.

---

## CRITICAL Issues

### 1. Build Failure: `pkg/clob/ws` - PriceEvent vs PriceChangeEvent Type Mismatch

**Files affected:**
- `pkg/clob/ws/impl_test.go` (lines 394, 422, 431, 432, 442, 451, 452)
- `pkg/clob/ws/client_test.go` (lines 72, 73)
- `pkg/clob/ws/subscription_panic_test.go` (line 251)

**Description:**
The production code in `impl.go` defines `priceSubs` as `map[string]*subscriptionEntry[PriceChangeEvent]` (line 77), but multiple test files use `subscriptionEntry[PriceEvent]` instead. `PriceEvent` and `PriceChangeEvent` are different types:
- `PriceEvent` contains `Market`, `PriceChanges []PriceChangeEvent`, `Timestamp` (a container/wrapper type)
- `PriceChangeEvent` contains `AssetId`, `BestAsk`, `BestBid`, `Hash`, `Price`, `Side`, `Size` (individual price change data)

The `SubscribePrices` method and `SubscribePricesStream` both operate on `PriceChangeEvent` channels, so the tests must use `PriceChangeEvent`.

**Specific errors:**
```
impl_test.go:394: cannot use map[string]*subscriptionEntry[PriceEvent] as map[string]*subscriptionEntry[PriceChangeEvent]
impl_test.go:422: cannot use &subscriptionEntry[PriceEvent]{...} as *subscriptionEntry[PriceChangeEvent]
impl_test.go:431: ev.AssetID undefined (type PriceEvent has no field AssetID)
client_test.go:72: event.AssetID undefined (type PriceChangeEvent has no field AssetID, but does have field AssetId)
```

**Suggested fix:**
1. In test files, change all `subscriptionEntry[PriceEvent]` to `subscriptionEntry[PriceChangeEvent]`
2. Change all `ev.AssetID` / `event.AssetID` to `ev.AssetId` / `event.AssetId` (matching the struct field name `AssetId` in `PriceChangeEvent`)
3. Update `newTestClient()` to use `subscriptionEntry[PriceChangeEvent]` for `priceSubs`

### 2. Build Failure: `pkg/clob/ws` - AssetID vs AssetId Field Name Mismatch

**Files affected:**
- `pkg/clob/ws/client_test.go` (lines 72, 73)
- `pkg/clob/ws/impl_test.go` (lines 431, 432, 451, 452)

**Description:**
`PriceChangeEvent` uses `AssetId` (lowercase 'd') as its Go field name (with JSON tag `json:"asset_id"`), but tests reference `AssetID` (uppercase 'D'). In Go, field names are case-sensitive. The correct field name is `AssetId`.

Note: Other event types like `OrderbookEvent`, `MidpointEvent`, etc. use `AssetID` (uppercase). This inconsistency in the codebase is itself a problem - see MEDIUM issue #1.

**Suggested fix:**
- In test files, change `event.AssetID` to `event.AssetId` when accessing `PriceChangeEvent` fields
- Or (preferred): rename the struct field in `types.go` from `AssetId` to `AssetID` to follow Go naming conventions

### 3. Test Failure: `pkg/clob` - OrderResponse JSON Tag Mismatch

**File:** `pkg/clob/clobtypes/types.go` (line 319-322)

**Description:**
```go
OrderResponse struct {
    ID     string `json:"orderID"`
    Status string `json:"status"`
}
```

The `ID` field has JSON tag `json:"orderID"`, meaning it expects JSON key `"orderID"` for deserialization. However:

1. **PostOrder test** (`impl_orders_test.go:24-39`): Mock returns `{"id":"o1","status":"OK"}` - key is `"id"`, not `"orderID"`. Deserialization silently leaves `resp.ID` as empty string.

2. **OrderLookup test** (`impl_orders_test.go:110-121`): Mock returns `{"id":"o1","status":"OK"}` - same mismatch.

The test error messages confirm this: `"PostOrder failed: <nil>"` means err is nil but `resp.ID != "o1"`.

**Impact:** The actual Polymarket API returns `"orderID"` for POST /order responses. The JSON tag is likely correct for API alignment, but the test mocks use the wrong key. Additionally, for GET /data/order/{id}, the API may return a different response shape. This needs verification.

**Suggested fix:**
- Fix test mocks to use `"orderID"` key: `{"orderID":"o1","status":"OK"}`
- For the Order lookup endpoint, verify what key the API actually returns and adjust accordingly

---

## HIGH Severity Issues

### 1. Potential Reentrant RLock Deadlock in subscriptionEntry

**File:** `pkg/clob/ws/subscription_manager.go` (lines 53-83)

**Description:**
`trySend()` acquires `s.mu.RLock()` (line 54), then on the full-channel path calls `s.notifyLag(1)` (line 65). `notifyLag()` also acquires `s.mu.RLock()` (line 73). This is a reentrant read lock from the same goroutine.

In Go, `sync.RWMutex` is not reentrant. If a writer (`close()` which calls `s.mu.Lock()`) is pending between the two RLock calls, the second RLock in `notifyLag` will block (waiting for the writer), while the writer blocks on the first RLock held by `trySend`. This creates a deadlock.

**Reproduction scenario:**
1. Goroutine A calls `trySend`, acquires RLock, channel is full
2. Goroutine B calls `close`, tries to acquire Lock (writer), blocks on A's RLock
3. Goroutine A calls `notifyLag`, tries to acquire RLock, blocks because writer B is waiting
4. Deadlock: A waits for B, B waits for A

**Suggested fix:**
Make `notifyLag` an internal method that assumes the caller already holds the RLock (don't re-acquire):
```go
func (s *subscriptionEntry[T]) trySend(msg T) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.closed { return }
    select {
    case s.ch <- msg: return
    default:
        // Inline lag notification without re-acquiring lock
        s.notifyLagLocked(1)
    }
}

func (s *subscriptionEntry[T]) notifyLagLocked(count int) {
    if count <= 0 || s.closed { return }
    select {
    case s.errCh <- LaggedError{Count: count, Channel: s.channel, EventType: s.event}:
    default:
    }
}
```

### 2. PriceChangeEvent.AssetId Naming Inconsistency

**File:** `pkg/clob/ws/types.go` (line 143)

**Description:**
`PriceChangeEvent.AssetId` uses `AssetId` while every other event type uses `AssetID`:
- `OrderbookEvent.AssetID`
- `MidpointEvent.AssetID`
- `TickSizeChangeEvent.AssetID`
- `LastTradePriceEvent.AssetID`
- `BestBidAskEvent.AssetID`
- etc.

Go convention is to capitalize acronyms: `ID` not `Id`. This is a breaking API change if renamed, but the JSON tag `json:"asset_id"` means the wire format is unaffected.

**Suggested fix:**
Rename `AssetId` to `AssetID` in `PriceChangeEvent` struct and update all references.

### 3. SubscribeOrders/SubscribeTrades Always Return Error

**File:** `pkg/clob/ws/impl.go` (lines 802-808)

**Description:**
```go
func (c *clientImpl) SubscribeOrdersStream(...) (*Stream[OrderEvent], error) {
    return nil, errors.New("markets required: use SubscribeUserOrdersStream")
}
func (c *clientImpl) SubscribeTradesStream(...) (*Stream[TradeEvent], error) {
    return nil, errors.New("markets required: use SubscribeUserTradesStream")
}
```

These methods are on the `Client` interface but always error. This is misleading API design. The interface suggests they're usable, but they're stubs that redirect to other methods.

**Suggested fix:**
Either remove from the interface, or make them work by requiring markets as parameters.

### 4. SubscriptionRequest.Type Uses ChannelSubscribe for User Subscriptions

**File:** `pkg/clob/ws/types.go` (lines 91-99)

**Description:**
`NewUserSubscription` uses `ChannelSubscribe` ("subscribe") as the type:
```go
func NewUserSubscription(markets []string) *SubscriptionRequest {
    return &SubscriptionRequest{
        Type:        ChannelSubscribe,  // "subscribe"
        ...
    }
}
```
But `NewUserUnsubscribe` uses `ChannelUser` ("user"):
```go
func NewUserUnsubscribe(markets []string) *SubscriptionRequest {
    return &SubscriptionRequest{
        Type:      ChannelUser,  // "user"
        ...
    }
}
```

This inconsistency may cause issues depending on what the Polymarket WS server expects. If "subscribe" and "user" are different channel types, subscribing and unsubscribing would target different channels.

**Impact:** Potential failure to correctly unsubscribe from user channels. Needs API documentation verification.

---

## MEDIUM Severity Issues

### 1. Test newTestClient() Creates priceSubs with Wrong Type

**File:** `pkg/clob/ws/impl_test.go` (line 394)

**Description:**
`newTestClient()` at line 394 initializes `priceSubs` as `map[string]*subscriptionEntry[PriceEvent]` but the actual `clientImpl` uses `map[string]*subscriptionEntry[PriceChangeEvent]`. This is part of the CRITICAL build failure.

### 2. processEvent Price Tests Mismatch Subscription Type

**File:** `pkg/clob/ws/impl_test.go` (lines 419-457)

**Description:**
`TestProcessEvent_Price` and `TestProcessEvent_PriceChange` create `subscriptionEntry[PriceEvent]` for `priceSubs`, but `dispatchPrice` in `impl.go` sends `PriceChangeEvent` to `priceSubs` entries. The test expects to receive `PriceEvent` with `AssetID` field, but the dispatch actually sends individual `PriceChangeEvent` items (which have `AssetId`).

Looking at `dispatchPrice` (line 661-673):
```go
func (c *clientImpl) dispatchPrice(event PriceEvent) {
    for _, sub := range subs {
        for _, priceChange := range event.PriceChanges {
            if sub.matchesAsset(priceChange.AssetId) {
                sub.trySend(priceChange)  // sends PriceChangeEvent, not PriceEvent
            }
        }
    }
}
```

So `priceSubs` receives `PriceChangeEvent` items, and tests must match this.

### 3. Missing Validation in Data Package Endpoints

**File:** `pkg/data/impl.go`

**Description:**
- `Value()` (line 174-185): Does not check if `req.User` is zero address, unlike `Positions()` and `Activity()` which do
- `Traded()` (line 219-229): Does not check if `req.User` is zero address

### 4. Client Interface Method Naming: SubscribeOrders vs SubscribeUserOrders

**File:** `pkg/clob/ws/client.go`

**Description:**
The interface has `SubscribeUserOrders` and `SubscribeUserTrades` but also `SubscribeOrders` and `SubscribeTrades` (via the hidden stream methods). The non-User versions always error. This is redundant and confusing. If the intent was to have convenience methods without the "User" prefix, they should delegate to the User versions.

### 5. NewUserSubscription Uses ChannelSubscribe Instead of ChannelUser

**File:** `pkg/clob/ws/types.go` (line 93)

**Description:**
The `applySubscription` method in `impl.go` routes based on `req.Type`:
- `ChannelMarket` -> market channel handling
- `ChannelUser` -> user channel handling

But `NewUserSubscription` sets `Type: ChannelSubscribe` which is `"subscribe"`, not `"user"`. The `applySubscription` handler at line 954-965 won't match either `ChannelMarket` or `ChannelUser` for a raw `NewUserSubscription` request, potentially hitting the default `"unknown subscription channel"` error.

However, `subscribeUserStream` calls `addUserRefs` and `ensureConn` directly without going through `applySubscription`, so the stream API works. The issue only affects direct `Subscribe()` calls with `NewUserSubscription`.

---

## LOW Severity Issues

### 1. Hardcoded Exchange Contract Address

**File:** `pkg/clob/impl_orders.go` (line 101)

**Description:**
```go
VerifyingContract: "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E",
```
This is hardcoded. If users need to target a different exchange (testnet, NegRisk exchange), they can't override it.

### 2. Inconsistent Error Handling in readLoop

**File:** `pkg/clob/ws/impl.go` (lines 384-461)

**Description:**
The `readLoop` uses `break` to exit the for loop on errors, but then checks `c.closing.Load()` after the loop. If the break was caused by a nil connection (after the sleep), the flow falls through to `c.shutdown()` which might be premature.

### 3. EIP-712 ClobAuthDomain Missing ChainId

**File:** `pkg/auth/auth.go` (lines 30-33)

**Description:**
```go
var ClobAuthDomain = &apitypes.TypedDataDomain{
    Name:    "ClobAuthDomain",
    Version: "1",
}
```
The `ClobAuthTypes` includes `chainId` in the EIP712Domain type, but `ClobAuthDomain` doesn't set `ChainId`. This means the domain separator hash will use a zero chain ID, which is incorrect for signature verification against a specific chain. The chain ID should be set dynamically based on the signer's chain.

### 4. UserSubscription Channel Type Confusion

**File:** `pkg/clob/ws/types.go` (lines 32-36, 91-107)

**Description:**
Three channel constants exist: `ChannelMarket`, `ChannelUser`, and `ChannelSubscribe`. The relationship between `ChannelSubscribe` and `ChannelUser` is unclear. `NewUserSubscription` uses `ChannelSubscribe` while `NewUserUnsubscribe` uses `ChannelUser`. Both should use the same channel type for consistency.

### 5. Duplicate GeoblockResponse Definition

**File:** `pkg/clob/clobtypes/types.go` (lines 312-317)

**Description:**
`GeoblockResponse` appears to be defined in the response types block (line 312) which shadows or duplicates the earlier definition. Go won't allow duplicate type names in the same package, so if there's only one actual definition this is fine, but the placement in the code suggests there may have been a copy-paste issue during development.

### 6. Missing Context Cancellation in Pagination Methods

**Files:** `pkg/clob/impl_orders.go`, `pkg/gamma/impl.go`

**Description:**
`OrdersAll()`, `TradesAll()`, `BuilderTradesAll()`, `EventsAll()`, `MarketsAll()` loop fetching pages but don't check context cancellation between iterations. A cancelled context will only be caught on the next HTTP request, not between pages. For long pagination sequences, this could delay cancellation.

---

## Test Quality Issues

### 1. processEvent Price Tests Won't Receive Expected Events

The `TestProcessEvent_Price` and `TestProcessEvent_PriceChange` tests at impl_test.go lines 419-457 will never receive events even after the type fix, because:

- `processEvent` for `"price"` or `"price_change"` unmarshals into `PriceEvent` which has `PriceChanges []PriceChangeEvent`
- The raw event `{"event_type": "price", "asset_id": "tok1", "price": "0.55"}` will unmarshal into a `PriceEvent` with empty `PriceChanges` slice
- `dispatchPrice` iterates over `event.PriceChanges` and sends each `PriceChangeEvent` to subscribers
- Since `PriceChanges` is empty, nothing gets dispatched

The test data needs to include a `price_changes` array for events to reach subscribers.

### 2. subscription_panic_test.go Uses PriceEvent for ConcurrentDispatchAndClose

**File:** `pkg/clob/ws/subscription_panic_test.go` (lines 244-253, 269)

The test creates `subscriptionEntry[PriceEvent]` for `priceSubs` (type mismatch, won't compile) and dispatches `PriceEvent{AssetID: "test", Price: "0.5"}` which has no `AssetID` field (should use `PriceChangeEvent`).

---

## Package-by-Package Summary

| Package | Status | Issues |
|---------|--------|--------|
| `pkg/clob/ws` | **BUILD FAILURE** | PriceEvent/PriceChangeEvent mismatch, AssetId/AssetID mismatch |
| `pkg/clob` | **TEST FAILURE** | OrderResponse JSON tag mismatch (`orderID` vs `id`) |
| `pkg/clob/rfq` | OK | No issues found |
| `pkg/clob/clobtypes` | Has bugs | OrderResponse JSON tag may not match all API responses |
| `pkg/transport` | OK | Well-structured with retry, circuit breaker, rate limiting |
| `pkg/auth` | OK | SignatureType uses explicit values (fixed from previous iota bug) |
| `pkg/gamma` | OK | Clean implementation |
| `pkg/data` | Minor | Missing zero-address validation on some endpoints |
| `pkg/rtds` | OK | Proper atomic/sync usage |
| `pkg/bridge` | OK | Withdraw returns unsupported (by design) |
| `pkg/ctf` | OK | Needs real backend for testing |
| `pkg/bot` | OK | No issues found |

---

## Recommended Fix Priority

1. **Immediate**: Fix `pkg/clob/ws` build failure (PriceEvent -> PriceChangeEvent, AssetId naming)
2. **Immediate**: Fix `pkg/clob` test failure (OrderResponse JSON tag or test mocks)
3. **High**: Fix reentrant RLock deadlock in subscriptionEntry
4. **Medium**: Standardize AssetId -> AssetID across PriceChangeEvent
5. **Medium**: Fix processEvent price test data to include price_changes array
6. **Low**: Address other issues as part of ongoing maintenance
