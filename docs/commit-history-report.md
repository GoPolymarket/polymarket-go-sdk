# Commit History Analysis Report

**Repository:** github.com/GoPolymarket/polymarket-go-sdk
**Analysis Date:** 2026-02-19
**Branch:** main

---

## 1. Summary Statistics

| Metric | Count |
|--------|-------|
| Total Commits | 87 |
| Bug Fix Commits | 37 (42.5%) |
| Feature Commits | 6 (6.9%) |
| Refactor Commits | 6 (6.9%) |
| Test-Only Commits | 6 (6.9%) |
| Documentation Commits | 11 (12.6%) |
| Chore/CI Commits | 13 (14.9%) |
| Merge Commits (PRs) | 11 |

### Contributor Breakdown

| Contributor | Commits | Notable Contributions |
|-------------|---------|----------------------|
| dongowu | 66 | Core maintainer, initial SDK, most bug fixes |
| donowu (same person) | 8 | Concurrency fixes (dual email) |
| lwtsn | 6 | External: tick size fix, order type fix, WS type changes |
| Vali Malinoiu (0x4139) | 3 | External: nil transport fix, gamma market tokens, orders_matched |
| minjieh (scream7) | 2 | External: Timestamp type fix (int64->string) |
| dependabot | 2 | Dependency bumps |

### External Contributor PRs

| PR | Author | Issue Fixed |
|----|--------|-------------|
| #14 | 0x4139 | Nil transport in clob.NewClient caused panic |
| #15 | 0x4139 | Missing Outcomes/OutcomePrices fields in gamma.Market |
| #13 | 0x4139 | Added typed subscription for orders_matched events |
| #21 | lwtsn | Incorrect tick size type (string vs float64) |
| #22 | lwtsn | Various SDK fixes (order type, test cases, OrderEvent expansion) |
| #24 | scream7 | TradeEvent.Timestamp type wrong (int64 vs string) |

---

## 2. Bug Pattern Analysis

### Category A: Type Mismatches with Upstream API

**Frequency:** 6 instances
**Severity:** High (causes runtime failures or data corruption)

| Commit | File | Bug | Fix |
|--------|------|-----|-----|
| `4366b65` | `clobtypes/types.go` | `TickSizeResponse.MinimumTickSize` was `string`, API returns `float64` | Changed to `float64` |
| `8346d55` | `clobtypes/types.go` | `PriceHistoryPoint.Price` was `string`, API returns `float64` | Changed to `float64` |
| `5fab2da` | `ws/types.go` | `TradeEvent.Timestamp` was `int64`, API sends `string` | Changed to `string` |
| `90b4304` | `clobtypes/types.go` | `OrderResponse.ID` JSON tag was `"id"`, API sends `"orderID"` | Fixed JSON tag |
| `dcdc3e5` | `gamma/types.go` | `Market` struct missing `Outcomes`/`OutcomePrices` fields entirely | Added missing fields |
| `c8f3ffd` | `ws/types.go` | `PriceEvent` struct didn't match API's nested `PriceChangeEvent` format | Introduced `PriceChangeEvent` wrapper |

**Root Cause:** Types were defined from documentation or assumptions rather than from actual API response inspection. No contract testing or response validation exists.

### Category B: Concurrency / Data Race Bugs

**Frequency:** 8 commits across 4 fix iterations
**Severity:** Critical (crashes, data corruption under load)

**Evolution of the subscription channel race fix:**

1. **`89b9bed`** (Feb 10): Initial fix - added `time.Sleep(10ms)` grace period before closing channels, removed `recover()` panic guard
2. **`b8ad5cc`** (Feb 11): Removed the sleep heuristic (it caused races itself)
3. **`013e6b7`** (Feb 11): Added `closed.Load()` check to `notifyLag()` to prevent panic
4. **`cd86d50`** (Feb 11): Replaced `atomic.Bool` with `sync.RWMutex` (the definitive fix)

This shows a classic "fix the fix" chain: 4 commits over 2 days to properly fix one race condition.

**Other concurrency fixes:**
- `9bbe467`: WebSocket `ReadTimeout` needed atomic access
- `601776d`: `CloneWithBaseURL` lost resilience settings (circuit breaker, rate limiter)
- `ecd851b`: Circuit breaker half-open accounting was wrong; rate limiter `Stop()` wasn't idempotent
- `013e6b7`: RateLimiter over-admission after sleep, WebSocket `closeConn` race, goroutine leak in `StreamDataWithCursor`

**Root Cause:** Channel-based concurrency in Go requires careful lifecycle management. The pattern of checking a flag then operating on a channel has an inherent TOCTOU (time-of-check-time-of-use) race that was repeatedly fixed with increasingly correct approaches.

### Category C: Field Naming / ID Confusion

**Frequency:** 5 commits
**Severity:** Medium (compilation errors, wrong API calls)

**The CancelOrderRequest ID oscillation (Jan 30):**

1. `840e1c0`: Initial implementation with `req.ID` and backward compatibility for `req.OrderID`
2. `5b655cf`: Changed body key from `"orderId"` to `"id"`
3. `28e2214`: Added `//lint:ignore` for deprecated field access
4. `543ff33`: Removed deprecated fields entirely (`ID`, `IDs`, `MarketID`)
5. `5400cc2`: Changed back to use `req.ID` (but field was removed!)
6. `c957e59`: Changed to use `req.OrderID` with `"orderId"` key

All 5 of these were committed within 30 minutes on the same day. This was a rapid trial-and-error cycle of deprecation handling.

**Root Cause:** Unclear which field name the upstream API expects (`id` vs `orderId`), compounded by an attempt to maintain backward compatibility that was then removed.

### Category D: Enum Value Collisions

**Frequency:** 1 instance
**Severity:** Critical (wrong wallet type used for signing)

- `8346d55`: `SignatureType` used `iota` which made `SignatureMagic = 1` and `SignatureProxy = 1` (both value 1). Fixed by removing `SignatureMagic` and keeping `SignatureProxy = 1` explicitly.

**Root Cause:** Go's `iota` auto-increment was combined with explicit value assignment, causing a silent value collision. The comment said `SignatureMagic` was different from `SignatureProxy` but they shared the same value.

### Category E: Nil Pointer / Missing Initialization

**Frequency:** 3 instances
**Severity:** High (panics)

| Commit | Bug | Fix |
|--------|-----|-----|
| `eaa0c07` | `clob.NewClient(nil)` stored nil transport, panicking on first API call | Added nil guard with default `BaseURL` |
| `d9da7c4` | Root constructor didn't initialize RTDS client | Added RTDS initialization |
| `8346d55` | RTDS `subscribeRaw` sent messages before WebSocket was connected | Added `connReady` channel to gate subscriptions |

**Root Cause:** Inconsistent nil-handling patterns across packages. `gamma.NewClient` and `data.NewClient` handled nil, but `clob.NewClient` didn't.

### Category F: EIP-712 Signing Errors

**Frequency:** 1 instance
**Severity:** Critical (orders cannot be placed)

- `b942cf2`: EIP-712 type definition used `"clobtypes.Order"` instead of `"Order"` as the primary type name. This caused `SignTypedData` to fail, preventing all order placements.

**Root Cause:** Go package-qualified type name leaked into EIP-712 domain definition, which expects plain type names.

---

## 3. Type Evolution Timeline

### `pkg/clob/ws/types.go`

| Date | Commit | Change | Breaking? |
|------|--------|--------|-----------|
| Jan 28 | `205108b` | Initial types created | N/A |
| Feb 14 | `c8f3ffd` | `PriceEvent` restructured into `PriceEvent` + `PriceChangeEvent` wrapper | YES |
| Feb 14 | `c8f3ffd` | `PriceEvent.AssetID` renamed to `PriceChangeEvent.AssetId` (lowercase d) | YES |
| Feb 16 | `b9e6ff1` | `OrderEvent` completely rewritten: 9 fields -> 19 fields, all field names changed | YES |
| Feb 16 | `b9e6ff1` | `OrderEvent.Timestamp` changed from `int64` to `string` | YES |
| Feb 16 | `b9e6ff1` | `NewUserSubscription` changed from `ChannelUser` to `ChannelSubscribe` | YES |
| Feb 19 | `5fab2da` | `TradeEvent.Timestamp` changed from `int64` to `string` | YES |

### `pkg/clob/clobtypes/types.go`

| Date | Commit | Change | Breaking? |
|------|--------|--------|-----------|
| Jan 30 | `840e1c0` | Major expansion: added PriceHistoryInterval, BatchRequest forms, funder/salt fields | YES |
| Jan 30 | `5b655cf`-`c957e59` | CancelOrderRequest field oscillation (ID/OrderID, 5 commits in 30 min) | YES |
| Feb 9 | `8346d55` | `PriceHistoryPoint.Price`: `string` -> `float64` | YES |
| Feb 14 | `4366b65` | `TickSizeResponse.MinimumTickSize`/`TickSize`: `string` -> `float64` | YES |
| Feb 16 | `90b4304` | `OrderResponse.ID` JSON tag: `"id"` -> `"orderID"` | YES |

### `pkg/auth/auth.go`

| Date | Commit | Change | Breaking? |
|------|--------|--------|-----------|
| Feb 1 | `3f7dc6c` | Error vars re-exported from `pkg/errors` | No |
| Feb 9 | `8346d55` | `SignatureMagic` removed, `SignatureProxy = 1` made explicit | YES |

---

## 4. Incomplete Fix Analysis

### 4.1 Subscription Manager Race: 4 Attempts to Fix

The subscription channel close race required 4 sequential commits to properly fix:

- **Attempt 1** (`89b9bed`): Added `time.Sleep(10ms)` before closing - **introduced new race**
- **Attempt 2** (`b8ad5cc`): Removed sleep, relied on atomic + non-blocking send - **still racy (TOCTOU)**
- **Attempt 3** (`013e6b7`): Added closed check to `notifyLag` - **partial fix**
- **Attempt 4** (`cd86d50`): Replaced atomic.Bool with RWMutex - **complete fix**

### 4.2 CancelOrder Field Name: 5 Attempts

Within 30 minutes on Jan 30, the CancelOrder request body went through:

1. `req.OrderID || req.ID` with `"orderId"` key
2. Changed to `"id"` key
3. Added lint ignore
4. Removed deprecated fields (breaking references)
5. Changed back to `req.ID` (field no longer exists)
6. Changed to `req.OrderID` with `"orderId"` key

### 4.3 Rate Limiter: 3 Iterations

- **v1** (`3f7dc6c`): Channel-based token bucket with background goroutine
- **v2** (`89b9bed`): Timestamp-based (eliminated goroutine) but had over-admission bug
- **v3** (`013e6b7`): Added re-check loop after sleep to fix over-admission

### 4.4 Tick Size Type Change: Incomplete Initial Fix

- `4366b65` changed `TickSizeResponse` fields from `string` to `float64`
- But `order_builder.go` still used `decimal.NewFromInt(int64(tickStr))` which truncates float64
- `d71bb96` followed to fix the order builder to use `decimal.NewFromFloat()`

### 4.5 Duplicate Commit

- `8346d55` and `a9a3cc8` have identical commit messages and are sequential, suggesting a rebase or amend issue. The first (`8346d55`) contains the actual changes, the second (`a9a3cc8`) only removes 2 lines from a test file.

---

## 5. Package Reliability Ranking

Ranked from **most problematic** to **least problematic** based on bug-fix-to-total-commit ratio and severity of bugs found:

| Rank | Package | Total Commits | Fix Commits | Fix Ratio | Critical Bugs |
|------|---------|---------------|-------------|-----------|---------------|
| 1 | `pkg/clob/ws` | 18 | 11 | 61% | Race conditions, goroutine leaks, type mismatches |
| 2 | `pkg/transport` | 15 | 9 | 60% | Rate limiter over-admission, circuit breaker accounting, data races |
| 3 | `pkg/clob` (core) | 34 | 20 | 59% | Nil transport panic, field naming oscillation, order signing failure |
| 4 | `pkg/auth` | 8 | 3 | 38% | SignatureType enum collision |
| 5 | `pkg/rtds` | 8 | 1 | 13% | Subscribe before connection ready |
| 6 | `pkg/gamma` | 5 | 0 | 0% | Missing struct fields (found by external contributor) |
| 7 | `pkg/data` | 6 | 0 | 0% | No bugs found |
| 8 | `pkg/bridge` | 5 | 0 | 0% | No bugs found |
| 9 | `pkg/ctf` | 4 | 0 | 0% | No bugs found |

---

## 6. Root Cause Analysis

### 6.1 Systemic Issue: No API Contract Validation

**Impact:** 6 type mismatch bugs
**Evidence:** Every external contributor PR fixed a type mismatch with the upstream API.

The SDK types were defined based on assumptions rather than actual API response schemas. There is no schema validation, no integration test against a real/mock API, and no type-generation from OpenAPI specs.

### 6.2 Systemic Issue: Concurrency Patterns Not Established Before Implementation

**Impact:** 8+ concurrency fix commits
**Evidence:** The subscription manager went through 4 iterations. The rate limiter went through 3 iterations. Both converged on standard Go patterns (RWMutex, timestamp-based tokens) that could have been the initial design.

The channel-based approach with `recover()` panic guards and `time.Sleep()` heuristics are anti-patterns that were iteratively replaced with proper mutex-based synchronization.

### 6.3 Systemic Issue: Rapid-Fire Fix Commits Without Testing

**Impact:** "Fix the fix" chains
**Evidence:** 5 commits in 30 minutes for CancelOrder fields. 4 commits in 2 days for subscription races. Multiple commits with messages just saying "fix" without description.

Changes are being committed before running the full test suite or verifying against the API.

### 6.4 Systemic Issue: Breaking Changes Without Versioning

**Impact:** Consumer compatibility
**Evidence:** At least 7 breaking changes to public types (PriceEvent, OrderEvent, TickSizeResponse, etc.) with no semver bumps, no changelog entries, and no deprecation period.

### 6.5 Systemic Issue: Tests Don't Catch Real Bugs

**Impact:** All major bugs were found by manual testing or external contributors
**Evidence:**
- Type mismatches (string vs float64) weren't caught because tests used string values
- Nil transport panic wasn't caught because tests always passed a non-nil client
- SignatureType collision wasn't caught because tests didn't verify enum values
- Race conditions weren't caught until the `-race` flag was enabled in CI

---

## 7. Recommendations

### High Priority

1. **Add API contract tests**: Record actual API responses and validate SDK types against them. Consider generating types from the Polymarket OpenAPI spec if available.

2. **Enable `-race` in CI permanently**: Multiple race conditions were only found after enabling the race detector. It should always be on.

3. **Establish concurrency review checklist**: Before any channel-based code is merged, verify:
   - No TOCTOU races on channel state
   - No `recover()` for closed channel panics (use mutex instead)
   - No `time.Sleep()` for synchronization
   - Goroutine lifecycle managed by context cancellation

4. **Run full test suite before committing**: Avoid rapid-fire fix chains by running `go test -race ./...` before each commit.

### Medium Priority

5. **Adopt semantic versioning**: Track breaking changes properly. Any struct field type change or removal should bump the minor version.

6. **Add integration test suite**: Create a mock server that returns actual API response shapes to catch type mismatches early.

7. **Standardize field naming**: Establish a convention for JSON field names (camelCase vs snake_case) and enforce it with linting.

### Low Priority

8. **Improve commit hygiene**: Avoid commits with just "fix" as the message. Every commit should describe what changed and why.

9. **Add pre-commit hooks**: Run `go vet`, `staticcheck`, and `go test -race` before allowing commits.

10. **Document the type oscillation pattern**: When the upstream API changes field names or types, document the mapping explicitly so future contributors know the canonical names.
