# Polymarket Trader — Design Document

**Date**: 2026-02-24
**Status**: Approved
**Goal**: Build a standalone auto-trading bot that uses `go-polymarket-sdk` to generate real trading volume on Polymarket, qualify for the Builder Program grant ($2.5M+ pool).

## Context

The SDK (`go-polymarket-sdk`) is feature-complete but produces zero trading volume. Polymarket grants are awarded based on volume, not code quality. A separate project (`polymarket-trader`) will serve as the SDK's first real user and the vehicle for grant qualification.

## Architecture: Single Process + Goroutine Concurrency

```
polymarket-trader/
├── cmd/
│   └── trader/
│       └── main.go           # Entry: load config → start components → graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go         # YAML + env merge, all params centralized
│   ├── feed/
│   │   └── feed.go           # Wraps SDK WS client, aggregates orderbook/price/trade events
│   ├── strategy/
│   │   ├── maker.go          # Market making: two-sided quoting, dynamic spread
│   │   └── taker.go          # Signal-driven: imbalance detection, FAK market orders
│   ├── risk/
│   │   └── manager.go        # Unified risk: position caps, daily loss, order limits
│   └── metrics/
│       └── metrics.go        # Simple log metrics: trade count, success rate, PnL
├── config.yaml               # Default config template
├── Dockerfile                 # Multi-stage build
├── docker-compose.yml         # One-click: trader + optional signer-server
├── go.mod                     # depends on github.com/GoPolymarket/polymarket-go-sdk
└── Makefile                   # build / run / docker shortcuts
```

Single Go binary. Goroutines for concurrency. No message queues, no plugin systems, no frameworks beyond the SDK itself.

## Data Flow

```
              Polymarket CLOB
                    │
         ┌──────────┼──────────┐
         ▼          ▼          ▼
    WS Orderbook  WS Price  WS Trades
         │          │          │
         └──────────┼──────────┘
                    ▼
              ┌───────────┐
              │   Feed    │  Aggregates WS events, maintains local book snapshot
              └─────┬─────┘
                    │ chan FeedEvent
         ┌──────────┼──────────┐
         ▼                     ▼
   ┌───────────┐       ┌───────────┐
   │  Maker    │       │  Taker    │
   │ Strategy  │       │ Strategy  │
   └─────┬─────┘       └─────┬─────┘
         │                    │
         └─────────┬──────────┘
                   ▼
           ┌──────────────┐
           │ Risk Manager │
           └──────┬───────┘
                  ▼
           ┌──────────────┐
           │  Executor    │  SDK OrderBuilder → CreateOrder (auto Builder Code)
           └──────────────┘
```

Key decisions:
- WS push (not REST polling) for low latency
- Maker re-quotes on every book update
- Taker triggers on imbalance threshold
- Maker uses GTC limit orders, Taker uses FAK market orders
- Builder Code mounted at SDK client init, all orders auto-attributed

## Strategy: Maker (Market Making)

On each orderbook update:
1. Compute mid price from best bid/ask
2. Calculate dynamic half-spread: `max(min_spread_bps, volatility * spread_multiplier)`
3. Place buy limit at `mid - halfSpread`, sell limit at `mid + halfSpread`
4. Cancel-and-replace: cancel stale orders before placing new ones
5. Force refresh every N seconds even if book hasn't changed

Auto market selection when no markets specified:
- Scan active markets, filter by depth (>100 shares) and spread range
- Select top N by liquidity
- Re-evaluate hourly

```yaml
maker:
  enabled: true
  markets: []
  auto_select_top: 5
  min_spread_bps: 20
  spread_multiplier: 1.5
  order_size_usdc: 25
  refresh_interval: 5s
  max_orders_per_market: 2
```

## Strategy: Taker (Signal-Driven)

On each orderbook update:
1. Sum top N depth levels for bid/ask sides
2. Compute imbalance: `(bidDepth - askDepth) / totalDepth`
3. If `abs(imbalance) < threshold` → skip
4. If same market traded within cooldown → skip
5. Place FAK market order with slippage guard

Improvements over existing `pkg/bot`:
- WS-driven (not one-shot REST poll)
- Cooldown prevents consecutive losses on same market
- Configurable depth levels (top N, not just top 1)
- In-memory trade history for cooldown checks

```yaml
taker:
  enabled: true
  markets: []
  min_imbalance: 0.15
  depth_levels: 3
  amount_usdc: 20
  max_slippage_bps: 30
  cooldown: 60s
  min_confidence_bps: 25
```

## Risk Management

Three gates before any order is placed:

```
Order Request
    │
    ▼
[Gate 1] Order count: open_orders >= max → reject
    ▼
[Gate 2] Daily loss: realized PnL <= -max_daily_loss → stop all strategies
    ▼
[Gate 3] Per-market position: position >= max_per_market → reject
    ▼
  Pass → Executor
```

PnL tracking:
- Listen to WS `UserTrades` events for real-time fill notifications
- Maintain in-memory: `map[tokenID] → {avgEntryPrice, size, realizedPnL}`
- No database needed; on restart, recover open orders from Polymarket API

```yaml
risk:
  max_open_orders: 20
  max_daily_loss_usdc: 100
  max_position_per_market: 50
  emergency_stop: false
```

## Graceful Shutdown

On SIGINT/SIGTERM:
1. Stop strategies (no new orders)
2. Cancel all open orders (`CancelAll`)
3. Close WS connections
4. Print final PnL summary
5. Exit 0

## Deployment

Multi-stage Docker build (~15MB final image):

```dockerfile
FROM golang:1.24-alpine AS builder
COPY . .
RUN CGO_ENABLED=0 go build -o /trader ./cmd/trader/

FROM alpine:3.19
COPY --from=builder /trader /trader
COPY config.yaml /config.yaml
ENTRYPOINT ["/trader"]
```

docker-compose with optional remote signer:

```yaml
services:
  trader:
    build: .
    restart: unless-stopped
    env_file: .env
    volumes:
      - ./config.yaml:/config.yaml
  signer:
    build:
      context: ../go-polymarket-sdk
      dockerfile: cmd/signer-server/Dockerfile
    env_file: .env.signer
```

## Builder Code Integration

Zero additional code. SDK handles everything at init:

```go
client := polymarket.NewClient().CLOB.
    WithAuth(signer, apiKey).
    WithBuilderConfig(&auth.BuilderConfig{
        Local: &auth.BuilderCredentials{
            Key:        cfg.Builder.Key,
            Secret:     cfg.Builder.Secret,
            Passphrase: cfg.Builder.Passphrase,
        },
    })
// All subsequent CreateOrder calls auto-include builder attribution headers
```

## Config Loading Priority

```
CLI flags > Environment variables > config.yaml > Defaults
```

## Go-Live Checklist

1. Register Builder at `polymarket.com/settings?tab=builder`, get API keys
2. Run with `DRY_RUN=true`, verify full flow works
3. Small-amount live test ($5-10 per trade)
4. Confirm volume visible on `builders.polymarket.com`
5. Gradually increase amounts
6. Apply for Verified tier at `builder@polymarket.com`
7. Submit grant application with volume data

## Dependencies

- `github.com/GoPolymarket/polymarket-go-sdk` — CLOB, WS, Auth, Builder
- `gopkg.in/yaml.v3` — config parsing
- No other frameworks
