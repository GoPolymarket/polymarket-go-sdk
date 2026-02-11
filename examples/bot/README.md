# Polymarket Bot (scan + analyze + guarded execution)

This bot is a complete end-to-end workflow built on the SDK:

1. Scan active markets
2. Analyze token books (spread + imbalance + confidence)
3. Build a risk-capped trade plan
4. Execute only when explicitly enabled and confirmed

## Run

```bash
go run ./cmd/polymarket-bot
```

Live execution (requires manual confirmation):

```bash
go run ./cmd/polymarket-bot --execute
```

Skip prompt confirmation:

```bash
go run ./cmd/polymarket-bot --execute --yes
```

## Required env

- `POLYMARKET_PK`
- `POLYMARKET_API_KEY`
- `POLYMARKET_API_SECRET`
- `POLYMARKET_API_PASSPHRASE`

## Optional bot env

- `BOT_SCAN_LIMIT` (default `60`)
- `BOT_TOP_N` (default `8`)
- `BOT_DEFAULT_AMOUNT_USDC` (default `25`)
- `BOT_MAX_PER_TRADE_USDC` (default `100`)
- `BOT_MAX_SLIPPAGE_BPS` (default `25`)
- `BOT_DRY_RUN` (`true`/`false`)

## Safety

- Dry-run by default
- `--execute` required for live orders
- Interactive confirmation required unless `--yes`
- Per-trade amount cap + open-order cap checks
