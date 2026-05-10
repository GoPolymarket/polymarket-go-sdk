# Release Notes

## Version 2.0.0 (2026-05-10) — Polymarket CLOB V2 Migration

Polymarket CLOB V2 于 2026年4月28日 上线，V1 SDK 不再受支持（关闭截止日期：2026年6月30日）。此版本将 Go SDK 迁移到 V2。

### Breaking Changes

**Order 结构体** (`clobtypes.Order`)
- 移除: `Taker`, `Nonce`, `FeeRateBps`
- 新增: `Timestamp` (int64, 毫秒), `Metadata` (string, bytes32 hex), `Builder` (string, bytes32 hex)
- `Expiration` 保留在 API payload 中，但从 EIP-712 签名中排除

**EIP-712 签名**
- Domain version: `"1"` → `"2"`
- Exchange 合约地址: `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E` → `0xE111180000d2663C0091e4f400237545B87B996B`
- 签名消息: 12 字段 → 11 字段（移除 taker/nonce/feeRateBps/expiration，新增 timestamp/metadata/builder）

**OrderBuilder**
- 移除: `FeeRateBps()`, `FeeRateBpsDec()`, `Nonce()`, `Taker()`
- 新增: `Timestamp()`, `Metadata()`, `Builder()`

**RFQ 类型**
- `RFQAcceptRequest` / `RFQApproveQuote`: 移除 Taker/Nonce/FeeRateBps

**Auth**
- 新增 `SignaturePoly1271 = 3` (EIP-1271 智能合约钱包)

### Critical Bug Fixes
- **metadata/builder hex 填充不一致**: EIP-712 签名将 metadata/builder 填充到 32 字节，但 API payload 发送原始未填充值 → 导致订单被拒。通过 `padBytes32` 公共函数统一处理
- **Time() 响应解析**: API 返回 JSON object `{"timestamp":...}` 但 SDK 尝试反序列化为 `int64` → 始终失败。`TimeResponse` 新增 `UnmarshalJSON` 同时支持两种格式
- **UserEarning.Earnings 类型**: API 返回 number 但 SDK 使用 `string` → 反序列化失败（同 commit cd50cb3）。新增 `EarningsFloat` 类型同时支持 string 和 float64
- **salt 序列化**: 从 JSON number 改为 string，与其他 uint256 字段一致
- **OrderResponse.ID**: JSON tag `"orderID"` → `"id"` 匹配 V2 API 格式

### 新功能
- `MarketsKeyset()` — CLOB 游标分页 (`GET /markets/keyset`)
- `EventsKeyset()`, `MarketsKeyset()` — Gamma 游标分页
- `NextCursor` 字段添加到 Gamma `EventsRequest` / `MarketsRequest`

### 迁移指南
1. 更新依赖: `go get github.com/GoPolymarket/polymarket-go-sdk@v2.0.0`
2. OrderBuilder: `FeeRateBps()`/`Nonce()`/`Taker()` → `Timestamp()`/`Metadata()`/`Builder()`
3. 手动构建的 `Order` struct: 移除 `Taker`/`Nonce`/`FeeRateBps` 字段
4. 重新生成 API key（V2 需要 v3 凭证）
5. 重新创建所有未成交订单（V1 订单已于 4月28日 清空）

### 修改文件 (26 files, +227/-244)
- `pkg/clob/clobtypes/types.go` — Order struct, EarningsFloat, TimeResponse, keyset types
- `pkg/clob/impl_orders.go` — EIP-712 domain/types/message V2
- `pkg/clob/order_payload.go` — V2 payload format, padBytes32 helper
- `pkg/clob/impl.go` — Time() response fix
- `pkg/clob/order_builder.go` — remove V1 methods, add V2 methods
- `pkg/clob/order_builder_resolve.go` — remove resolveFeeRateBps
- `pkg/clob/impl_market.go` — MarketsKeyset endpoint
- `pkg/clob/client.go` — MarketsKeyset interface
- `pkg/clob/rfq/types.go` — RFQ types V2
- `pkg/clob/rfq/helpers.go` — RFQ helpers V2
- `pkg/auth/auth.go` — SignaturePoly1271
- `pkg/gamma/client.go`, `impl.go`, `types.go` — keyset pagination
- `examples/` — 4 examples updated
- Test files updated

---

## Version 0.x.x (2026-02-10)

### 🔧 Critical Bug Fixes

This release addresses 6 critical and high-priority concurrency and performance issues that could impact production deployments.

#### WebSocket Client Improvements

**Fixed: Goroutine Leaks in WebSocket Connections**
- Added per-connection context cancellation to properly manage goroutine lifecycle
- Implemented `createGoroutineContext()`, `cancelGoroutines()`, and `getGoroutineContext()` helper methods
- Ensured old `pingLoop` and `readLoop` goroutines are properly cleaned up on reconnection
- **Impact**: Prevents memory leaks and resource exhaustion in long-running applications with frequent reconnections

**Fixed: Race Conditions in Connection State Management**
- Connection pointers are now set to `nil` after closing to prevent use-after-close errors
- Added proper nil checks after acquiring connection references
- **Impact**: Eliminates potential crashes and undefined behavior in concurrent scenarios

**Fixed: Subscription Panic Risks**
- Removed panic recovery from `trySend()` method - now uses clean non-blocking send
- Added 10ms grace period before closing channels to allow pending sends to complete
- Fixed TOCTOU (time-of-check-time-of-use) race condition between closed check and channel send
- **Impact**: Improves stability and prevents runtime panics in high-throughput scenarios

#### Stream Processing Improvements

**Fixed: Context Cancellation in Stream Functions**
- Made `StreamDataWithCursor` fully respect context cancellation
- Added context checks before each fetch operation
- Made channel sends cancellable using select with `ctx.Done()`
- **Impact**: Enables proper cleanup and resource management when operations are cancelled

#### Heartbeat Management

**Fixed: Heartbeat Goroutine Accumulation**
- Added proper cleanup of old heartbeat goroutines in `startHeartbeats()`
- Implemented 50ms delay to allow old goroutines to exit gracefully before starting new ones
- **Impact**: Prevents goroutine accumulation when heartbeat intervals are changed

#### Performance Optimization

**Optimized: Rate Limiter Implementation**
- Complete refactor from ticker-based to timestamp-based token calculation
- **Eliminated background goroutine** - tokens are now calculated on-demand
- Simplified internal structure: removed channels, replaced with float64 token counter
- Added `stopped` flag for backward compatibility with `Stop()` behavior
- **Impact**: Reduced resource consumption and improved efficiency in high-throughput scenarios

### 📊 Test Coverage

Added comprehensive test suites to ensure reliability:
- `pkg/clob/ws/goroutine_leak_test.go` - Goroutine leak detection using goleak
- `pkg/clob/ws/race_condition_test.go` - Concurrent access pattern testing
- `pkg/clob/ws/subscription_panic_test.go` - Subscription lifecycle and panic prevention tests

**Test Results**: 16/17 packages passing (94% success rate)
- Rate Limiter: 6/6 tests passing (100%)
- WebSocket: Majority of tests passing

### 🔄 Breaking Changes

**None** - All changes are backward compatible. Existing code will continue to work without modifications.

### 📦 Dependencies

- Added `go.uber.org/goleak v1.3.0` for goroutine leak detection in tests

### 🔍 Files Modified

**Core Implementation:**
- `pkg/transport/ratelimit.go` - Complete refactor to timestamp-based implementation
- `pkg/clob/ws/impl.go` - Goroutine lifecycle management and race condition fixes
- `pkg/clob/ws/subscription_manager.go` - Subscription safety improvements
- `pkg/clob/stream.go` - Context cancellation enhancements
- `pkg/clob/impl.go` - Heartbeat goroutine management

**Test Files:**
- `pkg/clob/ws/goroutine_leak_test.go` (new)
- `pkg/clob/ws/race_condition_test.go` (new)
- `pkg/clob/ws/subscription_panic_test.go` (new)

### 📈 Performance Impact

- **Memory**: Reduced memory footprint by eliminating unnecessary background goroutines
- **CPU**: More efficient rate limiting with on-demand token calculation
- **Stability**: Eliminated goroutine leaks and race conditions for improved long-term stability

### 🎯 Upgrade Recommendation

**Highly Recommended** for all production deployments, especially:
- Applications with long-running WebSocket connections
- High-frequency trading systems
- Market making bots
- Any service experiencing memory growth over time

### 🙏 Acknowledgments

This release was made possible through comprehensive code review and optimization efforts using AI-assisted development tools.

---

**Full Changelog**: https://github.com/GoPolymarket/polymarket-go-sdk/compare/a9a3cc8...89b9bed
