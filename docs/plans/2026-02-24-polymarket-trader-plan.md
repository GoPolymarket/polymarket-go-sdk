# Polymarket Trader Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build `polymarket-trader`, a standalone auto-trading bot that generates real trading volume on Polymarket using `go-polymarket-sdk`, to qualify for the Builder Program grant.

**Architecture:** Single Go binary, goroutine concurrency. WS data feed drives two strategies (Maker + Taker) through a shared risk manager. Builder Code attribution is mounted at SDK client init.

**Tech Stack:** Go 1.24, `go-polymarket-sdk` (CLOB + WS + Auth + Builder), `gopkg.in/yaml.v3`, Docker

**New project directory:** `/Users/dongowu/code/project/project_ploymarket/polymarket-trader`

---

### Task 1: Scaffold Project

**Files:**
- Create: `go.mod`
- Create: `cmd/trader/main.go`
- Create: `Makefile`

**Step 1: Create project directory and go.mod**

```bash
mkdir -p /Users/dongowu/code/project/project_ploymarket/polymarket-trader
cd /Users/dongowu/code/project/project_ploymarket/polymarket-trader
git init
go mod init github.com/GoPolymarket/polymarket-trader
```

**Step 2: Add SDK dependency**

```bash
go get github.com/GoPolymarket/polymarket-go-sdk@latest
go get gopkg.in/yaml.v3
```

**Step 3: Create minimal main.go**

```go
// cmd/trader/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("polymarket-trader starting...")
	os.Exit(0)
}
```

**Step 4: Create Makefile**

```makefile
.PHONY: build run test clean

build:
	go build -o bin/trader ./cmd/trader/

run:
	go run ./cmd/trader/

test:
	go test ./... -v -race -count=1

clean:
	rm -rf bin/
```

**Step 5: Verify it compiles**

Run: `go build ./cmd/trader/`
Expected: no errors

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: scaffold polymarket-trader project"
```

---

### Task 2: Config Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `config.yaml`

**Step 1: Write the failing test**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Maker.MinSpreadBps <= 0 {
		t.Fatal("expected positive min spread bps")
	}
	if cfg.Risk.MaxOpenOrders <= 0 {
		t.Fatal("expected positive max open orders")
	}
	if cfg.ScanInterval <= 0 {
		t.Fatal("expected positive scan interval")
	}
}

func TestLoadFromYAML(t *testing.T) {
	yaml := `
scan_interval: 30s
maker:
  enabled: false
  order_size_usdc: 50
taker:
  min_imbalance: 0.2
risk:
  max_daily_loss_usdc: 200
`
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write([]byte(yaml))
	f.Close()

	cfg, err := LoadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Maker.Enabled {
		t.Fatal("expected maker disabled")
	}
	if cfg.Maker.OrderSizeUSDC != 50 {
		t.Fatalf("expected order size 50, got %f", cfg.Maker.OrderSizeUSDC)
	}
	if cfg.Taker.MinImbalance != 0.2 {
		t.Fatalf("expected min imbalance 0.2, got %f", cfg.Taker.MinImbalance)
	}
	if cfg.Risk.MaxDailyLossUSDC != 200 {
		t.Fatalf("expected max daily loss 200, got %f", cfg.Risk.MaxDailyLossUSDC)
	}
	if cfg.ScanInterval != 30*time.Second {
		t.Fatalf("expected 30s scan interval, got %v", cfg.ScanInterval)
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("TRADER_DRY_RUN", "false")
	cfg := Default()
	cfg.ApplyEnv()
	if cfg.DryRun {
		t.Fatal("expected dry run false from env")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v`
Expected: FAIL (package doesn't exist yet)

**Step 3: Write implementation**

```go
// internal/config/config.go
package config

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Auth
	PrivateKey      string `yaml:"private_key"`
	APIKey          string `yaml:"api_key"`
	APISecret       string `yaml:"api_secret"`
	APIPassphrase   string `yaml:"api_passphrase"`
	BuilderKey      string `yaml:"builder_key"`
	BuilderSecret   string `yaml:"builder_secret"`
	BuilderPassphrase string `yaml:"builder_passphrase"`

	// General
	ScanInterval time.Duration `yaml:"scan_interval"`
	DryRun       bool          `yaml:"dry_run"`
	LogLevel     string        `yaml:"log_level"`

	Maker  MakerConfig  `yaml:"maker"`
	Taker  TakerConfig  `yaml:"taker"`
	Risk   RiskConfig   `yaml:"risk"`
}

type MakerConfig struct {
	Enabled           bool          `yaml:"enabled"`
	Markets           []string      `yaml:"markets"`
	AutoSelectTop     int           `yaml:"auto_select_top"`
	MinSpreadBps      float64       `yaml:"min_spread_bps"`
	SpreadMultiplier  float64       `yaml:"spread_multiplier"`
	OrderSizeUSDC     float64       `yaml:"order_size_usdc"`
	RefreshInterval   time.Duration `yaml:"refresh_interval"`
	MaxOrdersPerMarket int          `yaml:"max_orders_per_market"`
}

type TakerConfig struct {
	Enabled         bool          `yaml:"enabled"`
	Markets         []string      `yaml:"markets"`
	MinImbalance    float64       `yaml:"min_imbalance"`
	DepthLevels     int           `yaml:"depth_levels"`
	AmountUSDC      float64       `yaml:"amount_usdc"`
	MaxSlippageBps  float64       `yaml:"max_slippage_bps"`
	Cooldown        time.Duration `yaml:"cooldown"`
	MinConfidenceBps float64      `yaml:"min_confidence_bps"`
}

type RiskConfig struct {
	MaxOpenOrders       int     `yaml:"max_open_orders"`
	MaxDailyLossUSDC    float64 `yaml:"max_daily_loss_usdc"`
	MaxPositionPerMarket float64 `yaml:"max_position_per_market"`
	EmergencyStop       bool    `yaml:"emergency_stop"`
}

func Default() Config {
	return Config{
		ScanInterval: 10 * time.Second,
		DryRun:       true,
		LogLevel:     "info",
		Maker: MakerConfig{
			Enabled:            true,
			AutoSelectTop:      5,
			MinSpreadBps:       20,
			SpreadMultiplier:   1.5,
			OrderSizeUSDC:      25,
			RefreshInterval:    5 * time.Second,
			MaxOrdersPerMarket: 2,
		},
		Taker: TakerConfig{
			Enabled:          true,
			MinImbalance:     0.15,
			DepthLevels:      3,
			AmountUSDC:       20,
			MaxSlippageBps:   30,
			Cooldown:         60 * time.Second,
			MinConfidenceBps: 25,
		},
		Risk: RiskConfig{
			MaxOpenOrders:        20,
			MaxDailyLossUSDC:     100,
			MaxPositionPerMarket: 50,
		},
	}
}

func LoadFile(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c *Config) ApplyEnv() {
	if v := os.Getenv("POLYMARKET_PK"); v != "" {
		c.PrivateKey = v
	}
	if v := os.Getenv("POLYMARKET_API_KEY"); v != "" {
		c.APIKey = v
	}
	if v := os.Getenv("POLYMARKET_API_SECRET"); v != "" {
		c.APISecret = v
	}
	if v := os.Getenv("POLYMARKET_API_PASSPHRASE"); v != "" {
		c.APIPassphrase = v
	}
	if v := os.Getenv("BUILDER_KEY"); v != "" {
		c.BuilderKey = v
	}
	if v := os.Getenv("BUILDER_SECRET"); v != "" {
		c.BuilderSecret = v
	}
	if v := os.Getenv("BUILDER_PASSPHRASE"); v != "" {
		c.BuilderPassphrase = v
	}
	if v := os.Getenv("TRADER_DRY_RUN"); v != "" {
		c.DryRun = strings.EqualFold(v, "true") || v == "1"
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/config/ -v`
Expected: PASS

**Step 5: Create config.yaml template**

```yaml
# config.yaml — polymarket-trader default configuration
scan_interval: 10s
dry_run: true
log_level: info

maker:
  enabled: true
  markets: []           # empty = auto-select
  auto_select_top: 5
  min_spread_bps: 20
  spread_multiplier: 1.5
  order_size_usdc: 25
  refresh_interval: 5s
  max_orders_per_market: 2

taker:
  enabled: true
  markets: []
  min_imbalance: 0.15
  depth_levels: 3
  amount_usdc: 20
  max_slippage_bps: 30
  cooldown: 60s
  min_confidence_bps: 25

risk:
  max_open_orders: 20
  max_daily_loss_usdc: 100
  max_position_per_market: 50
  emergency_stop: false
```

**Step 6: Commit**

```bash
git add internal/config/ config.yaml
git commit -m "feat: add config package with YAML + env support"
```

---

### Task 3: Feed Package (WS Data Aggregation)

**Files:**
- Create: `internal/feed/feed.go`
- Create: `internal/feed/feed_test.go`

**Step 1: Write the failing test**

```go
// internal/feed/feed_test.go
package feed

import (
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

func TestBookSnapshotUpdate(t *testing.T) {
	snap := NewBookSnapshot()
	event := ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "100"}, {Price: "0.49", Size: "200"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "150"}, {Price: "0.53", Size: "250"}},
	}
	snap.Update(event)

	book, ok := snap.Get("token-1")
	if !ok {
		t.Fatal("expected book for token-1")
	}
	if len(book.Bids) != 2 {
		t.Fatalf("expected 2 bid levels, got %d", len(book.Bids))
	}
	if book.Bids[0].Price != "0.50" {
		t.Fatalf("expected best bid 0.50, got %s", book.Bids[0].Price)
	}
	if len(book.Asks) != 2 {
		t.Fatalf("expected 2 ask levels, got %d", len(book.Asks))
	}
}

func TestBookSnapshotMid(t *testing.T) {
	snap := NewBookSnapshot()
	snap.Update(ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "100"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "100"}},
	})
	mid, err := snap.Mid("token-1")
	if err != nil {
		t.Fatal(err)
	}
	// mid = (0.50 + 0.52) / 2 = 0.51
	expected := 0.51
	if mid < expected-0.001 || mid > expected+0.001 {
		t.Fatalf("expected mid ~0.51, got %f", mid)
	}
}

func TestBookSnapshotDepth(t *testing.T) {
	snap := NewBookSnapshot()
	snap.Update(ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "100"}, {Price: "0.49", Size: "200"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "150"}, {Price: "0.53", Size: "250"}},
	})
	bidDepth, askDepth := snap.Depth("token-1", 2)
	if bidDepth != 300 {
		t.Fatalf("expected bid depth 300, got %f", bidDepth)
	}
	if askDepth != 400 {
		t.Fatalf("expected ask depth 400, got %f", askDepth)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/feed/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/feed/feed.go
package feed

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

// BookSnapshot maintains an in-memory orderbook snapshot per asset.
type BookSnapshot struct {
	mu    sync.RWMutex
	books map[string]ws.OrderbookEvent
}

func NewBookSnapshot() *BookSnapshot {
	return &BookSnapshot{books: make(map[string]ws.OrderbookEvent)}
}

func (s *BookSnapshot) Update(event ws.OrderbookEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.books[event.AssetID] = event
}

func (s *BookSnapshot) Get(assetID string) (ws.OrderbookEvent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.books[assetID]
	return b, ok
}

func (s *BookSnapshot) Mid(assetID string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.books[assetID]
	if !ok || len(b.Bids) == 0 || len(b.Asks) == 0 {
		return 0, fmt.Errorf("no book for %s", assetID)
	}
	bid, err := strconv.ParseFloat(b.Bids[0].Price, 64)
	if err != nil {
		return 0, err
	}
	ask, err := strconv.ParseFloat(b.Asks[0].Price, 64)
	if err != nil {
		return 0, err
	}
	return (bid + ask) / 2, nil
}

// Depth returns total bid and ask depth for the top n levels.
func (s *BookSnapshot) Depth(assetID string, levels int) (bidDepth, askDepth float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.books[assetID]
	if !ok {
		return 0, 0
	}
	for i := 0; i < levels && i < len(b.Bids); i++ {
		size, _ := strconv.ParseFloat(b.Bids[i].Size, 64)
		bidDepth += size
	}
	for i := 0; i < levels && i < len(b.Asks); i++ {
		size, _ := strconv.ParseFloat(b.Asks[i].Size, 64)
		askDepth += size
	}
	return bidDepth, askDepth
}

// AssetIDs returns all tracked assets.
func (s *BookSnapshot) AssetIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.books))
	for id := range s.books {
		ids = append(ids, id)
	}
	return ids
}
```

**Step 4: Run tests**

Run: `go test ./internal/feed/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/feed/
git commit -m "feat: add feed package with orderbook snapshot"
```

---

### Task 4: Risk Manager

**Files:**
- Create: `internal/risk/manager.go`
- Create: `internal/risk/manager_test.go`

**Step 1: Write the failing test**

```go
// internal/risk/manager_test.go
package risk

import (
	"testing"
)

func TestAllowOrderBasic(t *testing.T) {
	m := New(Config{MaxOpenOrders: 5, MaxDailyLossUSDC: 100, MaxPositionPerMarket: 50})
	if err := m.Allow("token-1", 25); err != nil {
		t.Fatalf("expected allow, got %v", err)
	}
}

func TestBlockOnMaxOrders(t *testing.T) {
	m := New(Config{MaxOpenOrders: 2, MaxDailyLossUSDC: 100, MaxPositionPerMarket: 50})
	m.SetOpenOrders(2)
	if err := m.Allow("token-1", 25); err == nil {
		t.Fatal("expected block on max orders")
	}
}

func TestBlockOnDailyLoss(t *testing.T) {
	m := New(Config{MaxOpenOrders: 20, MaxDailyLossUSDC: 100, MaxPositionPerMarket: 50})
	m.RecordPnL(-101)
	if err := m.Allow("token-1", 25); err == nil {
		t.Fatal("expected block on daily loss")
	}
}

func TestBlockOnPositionLimit(t *testing.T) {
	m := New(Config{MaxOpenOrders: 20, MaxDailyLossUSDC: 100, MaxPositionPerMarket: 50})
	m.AddPosition("token-1", 30)
	if err := m.Allow("token-1", 25); err == nil {
		t.Fatal("expected block on position limit")
	}
}

func TestEmergencyStop(t *testing.T) {
	m := New(Config{MaxOpenOrders: 20, MaxDailyLossUSDC: 100, MaxPositionPerMarket: 50})
	m.SetEmergencyStop(true)
	if err := m.Allow("token-1", 10); err == nil {
		t.Fatal("expected block on emergency stop")
	}
}

func TestRecordPnLAndReset(t *testing.T) {
	m := New(Config{MaxOpenOrders: 20, MaxDailyLossUSDC: 100, MaxPositionPerMarket: 50})
	m.RecordPnL(-50)
	m.RecordPnL(-40)
	if m.DailyPnL() != -90 {
		t.Fatalf("expected -90, got %f", m.DailyPnL())
	}
	m.ResetDaily()
	if m.DailyPnL() != 0 {
		t.Fatalf("expected 0 after reset, got %f", m.DailyPnL())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/risk/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/risk/manager.go
package risk

import (
	"fmt"
	"sync"
)

type Config struct {
	MaxOpenOrders        int
	MaxDailyLossUSDC     float64
	MaxPositionPerMarket float64
}

type Manager struct {
	mu            sync.RWMutex
	cfg           Config
	openOrders    int
	dailyPnL      float64
	positions     map[string]float64 // tokenID → USDC exposure
	emergencyStop bool
}

func New(cfg Config) *Manager {
	return &Manager{
		cfg:       cfg,
		positions: make(map[string]float64),
	}
}

func (m *Manager) Allow(tokenID string, amountUSDC float64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.emergencyStop {
		return fmt.Errorf("emergency stop active")
	}
	if m.openOrders >= m.cfg.MaxOpenOrders {
		return fmt.Errorf("max open orders reached: %d/%d", m.openOrders, m.cfg.MaxOpenOrders)
	}
	if m.dailyPnL <= -m.cfg.MaxDailyLossUSDC {
		return fmt.Errorf("daily loss limit reached: %.2f/%.2f", m.dailyPnL, -m.cfg.MaxDailyLossUSDC)
	}
	pos := m.positions[tokenID]
	if pos+amountUSDC > m.cfg.MaxPositionPerMarket {
		return fmt.Errorf("position limit for %s: %.2f+%.2f > %.2f", tokenID, pos, amountUSDC, m.cfg.MaxPositionPerMarket)
	}
	return nil
}

func (m *Manager) SetOpenOrders(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.openOrders = n
}

func (m *Manager) RecordPnL(amount float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dailyPnL += amount
}

func (m *Manager) AddPosition(tokenID string, amountUSDC float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.positions[tokenID] += amountUSDC
}

func (m *Manager) RemovePosition(tokenID string, amountUSDC float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.positions[tokenID] -= amountUSDC
	if m.positions[tokenID] <= 0 {
		delete(m.positions, tokenID)
	}
}

func (m *Manager) SetEmergencyStop(stop bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emergencyStop = stop
}

func (m *Manager) DailyPnL() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dailyPnL
}

func (m *Manager) ResetDaily() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dailyPnL = 0
}
```

**Step 4: Run tests**

Run: `go test ./internal/risk/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/risk/
git commit -m "feat: add risk manager with three-gate checks"
```

---

### Task 5: Maker Strategy

**Files:**
- Create: `internal/strategy/maker.go`
- Create: `internal/strategy/maker_test.go`

**Step 1: Write the failing test**

```go
// internal/strategy/maker_test.go
package strategy

import (
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

func TestMakerQuote(t *testing.T) {
	m := NewMaker(MakerConfig{
		MinSpreadBps:     20,
		SpreadMultiplier: 1.5,
		OrderSizeUSDC:    25,
	})

	book := ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "100"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "100"}},
	}

	quote, err := m.ComputeQuote(book)
	if err != nil {
		t.Fatal(err)
	}
	if quote.AssetID != "token-1" {
		t.Fatalf("expected token-1, got %s", quote.AssetID)
	}
	if quote.BuyPrice >= quote.SellPrice {
		t.Fatalf("buy %f should be less than sell %f", quote.BuyPrice, quote.SellPrice)
	}
	if quote.Size != 25 {
		t.Fatalf("expected size 25, got %f", quote.Size)
	}
}

func TestMakerSkipsEmptyBook(t *testing.T) {
	m := NewMaker(MakerConfig{MinSpreadBps: 20, SpreadMultiplier: 1.5, OrderSizeUSDC: 25})
	book := ws.OrderbookEvent{AssetID: "token-1"}
	_, err := m.ComputeQuote(book)
	if err == nil {
		t.Fatal("expected error on empty book")
	}
}

func TestMakerMinSpreadEnforced(t *testing.T) {
	m := NewMaker(MakerConfig{MinSpreadBps: 100, SpreadMultiplier: 1.0, OrderSizeUSDC: 25})
	// spread = (0.52-0.50)/0.51 = ~392 bps, but min is 100 bps = 1%
	book := ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.505", Size: "100"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.506", Size: "100"}},
	}
	quote, err := m.ComputeQuote(book)
	if err != nil {
		t.Fatal(err)
	}
	// half spread should be at least 100/2 = 50 bps of mid
	mid := (0.505 + 0.506) / 2
	minHalfSpread := mid * 50 / 10000
	actualHalf := (quote.SellPrice - quote.BuyPrice) / 2
	if actualHalf < minHalfSpread-0.0001 {
		t.Fatalf("half spread %f less than min %f", actualHalf, minHalfSpread)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/strategy/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/strategy/maker.go
package strategy

import (
	"fmt"
	"math"
	"strconv"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

type MakerConfig struct {
	MinSpreadBps       float64
	SpreadMultiplier   float64
	OrderSizeUSDC      float64
	MaxOrdersPerMarket int
}

type Quote struct {
	AssetID   string
	BuyPrice  float64
	SellPrice float64
	Size      float64
}

type Maker struct {
	cfg MakerConfig
}

func NewMaker(cfg MakerConfig) *Maker {
	return &Maker{cfg: cfg}
}

func (m *Maker) ComputeQuote(book ws.OrderbookEvent) (Quote, error) {
	if len(book.Bids) == 0 || len(book.Asks) == 0 {
		return Quote{}, fmt.Errorf("empty book for %s", book.AssetID)
	}

	bestBid, err := strconv.ParseFloat(book.Bids[0].Price, 64)
	if err != nil {
		return Quote{}, fmt.Errorf("bad bid price: %w", err)
	}
	bestAsk, err := strconv.ParseFloat(book.Asks[0].Price, 64)
	if err != nil {
		return Quote{}, fmt.Errorf("bad ask price: %w", err)
	}
	if bestAsk <= bestBid {
		return Quote{}, fmt.Errorf("crossed book: bid=%f ask=%f", bestBid, bestAsk)
	}

	mid := (bestBid + bestAsk) / 2
	marketSpreadBps := (bestAsk - bestBid) / mid * 10000

	// Dynamic half spread: use market spread * multiplier, but enforce minimum
	halfSpreadBps := math.Max(m.cfg.MinSpreadBps/2, marketSpreadBps*m.cfg.SpreadMultiplier/2)
	halfSpread := mid * halfSpreadBps / 10000

	buyPrice := mid - halfSpread
	sellPrice := mid + halfSpread

	// Clamp to valid range
	if buyPrice <= 0 {
		buyPrice = 0.01
	}
	if sellPrice >= 1 {
		sellPrice = 0.99
	}

	return Quote{
		AssetID:   book.AssetID,
		BuyPrice:  buyPrice,
		SellPrice: sellPrice,
		Size:      m.cfg.OrderSizeUSDC,
	}, nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/strategy/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/strategy/
git commit -m "feat: add maker strategy with dynamic spread"
```

---

### Task 6: Taker Strategy

**Files:**
- Create: `internal/strategy/taker.go`
- Create: `internal/strategy/taker_test.go`

**Step 1: Write the failing test**

```go
// internal/strategy/taker_test.go
package strategy

import (
	"testing"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

func TestTakerSignal(t *testing.T) {
	tk := NewTaker(TakerConfig{
		MinImbalance:   0.15,
		DepthLevels:    2,
		AmountUSDC:     20,
		MaxSlippageBps: 30,
		Cooldown:       1 * time.Second,
	})

	// Strong buy signal: bid depth >> ask depth
	book := ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "300"}, {Price: "0.49", Size: "200"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "50"}, {Price: "0.53", Size: "50"}},
	}
	// imbalance = (500-100)/600 = 0.667 > 0.15

	sig, err := tk.Evaluate(book)
	if err != nil {
		t.Fatal(err)
	}
	if sig == nil {
		t.Fatal("expected signal")
	}
	if sig.Side != "BUY" {
		t.Fatalf("expected BUY, got %s", sig.Side)
	}
	if sig.AmountUSDC != 20 {
		t.Fatalf("expected amount 20, got %f", sig.AmountUSDC)
	}
}

func TestTakerNoSignalLowImbalance(t *testing.T) {
	tk := NewTaker(TakerConfig{
		MinImbalance: 0.15,
		DepthLevels:  2,
		AmountUSDC:   20,
		Cooldown:     1 * time.Second,
	})

	// Balanced book
	book := ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "100"}, {Price: "0.49", Size: "100"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "100"}, {Price: "0.53", Size: "100"}},
	}

	sig, err := tk.Evaluate(book)
	if err != nil {
		t.Fatal(err)
	}
	if sig != nil {
		t.Fatal("expected no signal on balanced book")
	}
}

func TestTakerCooldown(t *testing.T) {
	tk := NewTaker(TakerConfig{
		MinImbalance: 0.10,
		DepthLevels:  1,
		AmountUSDC:   20,
		Cooldown:     100 * time.Millisecond,
	})

	book := ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "300"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "50"}},
	}

	sig1, _ := tk.Evaluate(book)
	if sig1 == nil {
		t.Fatal("expected first signal")
	}
	tk.RecordTrade("token-1")

	sig2, _ := tk.Evaluate(book)
	if sig2 != nil {
		t.Fatal("expected cooldown block")
	}

	time.Sleep(150 * time.Millisecond)
	sig3, _ := tk.Evaluate(book)
	if sig3 == nil {
		t.Fatal("expected signal after cooldown")
	}
}

func TestTakerSellSignal(t *testing.T) {
	tk := NewTaker(TakerConfig{
		MinImbalance: 0.15,
		DepthLevels:  1,
		AmountUSDC:   20,
		Cooldown:     1 * time.Second,
	})

	// Strong sell signal: ask depth >> bid depth
	book := ws.OrderbookEvent{
		AssetID: "token-1",
		Bids:    []ws.OrderbookLevel{{Price: "0.50", Size: "50"}},
		Asks:    []ws.OrderbookLevel{{Price: "0.52", Size: "300"}},
	}

	sig, _ := tk.Evaluate(book)
	if sig == nil {
		t.Fatal("expected signal")
	}
	if sig.Side != "SELL" {
		t.Fatalf("expected SELL, got %s", sig.Side)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/strategy/ -v -run TestTaker`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/strategy/taker.go
package strategy

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

type TakerConfig struct {
	MinImbalance     float64
	DepthLevels      int
	AmountUSDC       float64
	MaxSlippageBps   float64
	Cooldown         time.Duration
	MinConfidenceBps float64
}

type Signal struct {
	AssetID       string
	Side          string // BUY or SELL
	AmountUSDC    float64
	MaxPrice      float64 // slippage-guarded price
	Mid           float64
	Imbalance     float64
}

type Taker struct {
	cfg        TakerConfig
	mu         sync.Mutex
	lastTrades map[string]time.Time
}

func NewTaker(cfg TakerConfig) *Taker {
	return &Taker{
		cfg:        cfg,
		lastTrades: make(map[string]time.Time),
	}
}

func (tk *Taker) Evaluate(book ws.OrderbookEvent) (*Signal, error) {
	if len(book.Bids) == 0 || len(book.Asks) == 0 {
		return nil, fmt.Errorf("empty book for %s", book.AssetID)
	}

	// Check cooldown
	tk.mu.Lock()
	if last, ok := tk.lastTrades[book.AssetID]; ok && time.Since(last) < tk.cfg.Cooldown {
		tk.mu.Unlock()
		return nil, nil
	}
	tk.mu.Unlock()

	// Sum depth for top N levels
	var bidDepth, askDepth float64
	for i := 0; i < tk.cfg.DepthLevels && i < len(book.Bids); i++ {
		size, _ := strconv.ParseFloat(book.Bids[i].Size, 64)
		bidDepth += size
	}
	for i := 0; i < tk.cfg.DepthLevels && i < len(book.Asks); i++ {
		size, _ := strconv.ParseFloat(book.Asks[i].Size, 64)
		askDepth += size
	}

	totalDepth := bidDepth + askDepth
	if totalDepth == 0 {
		return nil, nil
	}

	imbalance := (bidDepth - askDepth) / totalDepth
	if math.Abs(imbalance) < tk.cfg.MinImbalance {
		return nil, nil
	}

	bestBid, _ := strconv.ParseFloat(book.Bids[0].Price, 64)
	bestAsk, _ := strconv.ParseFloat(book.Asks[0].Price, 64)
	mid := (bestBid + bestAsk) / 2

	side := "BUY"
	if imbalance < 0 {
		side = "SELL"
	}

	// Slippage guard
	delta := mid * tk.cfg.MaxSlippageBps / 10000
	maxPrice := mid + delta
	if side == "SELL" {
		maxPrice = mid - delta
		if maxPrice <= 0 {
			maxPrice = 0.01
		}
	}

	return &Signal{
		AssetID:    book.AssetID,
		Side:       side,
		AmountUSDC: tk.cfg.AmountUSDC,
		MaxPrice:   maxPrice,
		Mid:        mid,
		Imbalance:  imbalance,
	}, nil
}

func (tk *Taker) RecordTrade(assetID string) {
	tk.mu.Lock()
	defer tk.mu.Unlock()
	tk.lastTrades[assetID] = time.Now()
}
```

**Step 4: Run tests**

Run: `go test ./internal/strategy/ -v`
Expected: PASS (all maker + taker tests)

**Step 5: Commit**

```bash
git add internal/strategy/taker.go internal/strategy/taker_test.go
git commit -m "feat: add taker strategy with imbalance detection and cooldown"
```

---

### Task 7: Market Selector

**Files:**
- Create: `internal/strategy/selector.go`
- Create: `internal/strategy/selector_test.go`

**Step 1: Write the failing test**

```go
// internal/strategy/selector_test.go
package strategy

import (
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func TestSelectTopMarkets(t *testing.T) {
	markets := []clobtypes.Market{
		{ID: "m1", Tokens: []clobtypes.MarketToken{{TokenID: "t1", Price: 0.50}}, Active: true},
		{ID: "m2", Tokens: []clobtypes.MarketToken{{TokenID: "t2", Price: 0.90}}, Active: true},
		{ID: "m3", Tokens: []clobtypes.MarketToken{{TokenID: "t3", Price: 0.50}}, Active: true},
		{ID: "m4", Tokens: []clobtypes.MarketToken{{TokenID: "t4", Price: 0.50}}, Active: false}, // inactive
	}

	books := map[string]clobtypes.OrderBook{
		"t1": {Bids: []clobtypes.PriceLevel{{Price: "0.49", Size: "500"}}, Asks: []clobtypes.PriceLevel{{Price: "0.51", Size: "500"}}},
		"t2": {Bids: []clobtypes.PriceLevel{{Price: "0.89", Size: "10"}}, Asks: []clobtypes.PriceLevel{{Price: "0.91", Size: "10"}}},   // low depth
		"t3": {Bids: []clobtypes.PriceLevel{{Price: "0.49", Size: "200"}}, Asks: []clobtypes.PriceLevel{{Price: "0.51", Size: "200"}}},
	}

	selected := SelectMarkets(markets, books, 2, 50) // top 2, min depth 50
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(selected))
	}
	// t1 (depth 1000) should be first, t3 (depth 400) second, t2 excluded (depth 20)
	if selected[0] != "t1" {
		t.Fatalf("expected t1 first, got %s", selected[0])
	}
	if selected[1] != "t3" {
		t.Fatalf("expected t3 second, got %s", selected[1])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/strategy/ -v -run TestSelect`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/strategy/selector.go
package strategy

import (
	"sort"
	"strconv"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

type marketScore struct {
	tokenID    string
	totalDepth float64
}

func SelectMarkets(markets []clobtypes.Market, books map[string]clobtypes.OrderBook, topN int, minDepth float64) []string {
	var scores []marketScore

	for _, m := range markets {
		if !m.Active {
			continue
		}
		for _, tok := range m.Tokens {
			book, ok := books[tok.TokenID]
			if !ok || len(book.Bids) == 0 || len(book.Asks) == 0 {
				continue
			}
			var depth float64
			for _, lvl := range book.Bids {
				s, _ := strconv.ParseFloat(lvl.Size, 64)
				depth += s
			}
			for _, lvl := range book.Asks {
				s, _ := strconv.ParseFloat(lvl.Size, 64)
				depth += s
			}
			if depth >= minDepth {
				scores = append(scores, marketScore{tokenID: tok.TokenID, totalDepth: depth})
			}
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].totalDepth > scores[j].totalDepth
	})

	result := make([]string, 0, topN)
	for i := 0; i < topN && i < len(scores); i++ {
		result = append(result, scores[i].tokenID)
	}
	return result
}
```

**Step 4: Run tests**

Run: `go test ./internal/strategy/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/strategy/selector.go internal/strategy/selector_test.go
git commit -m "feat: add market selector by liquidity depth"
```

---

### Task 8: Main Entry Point (Full Lifecycle)

**Files:**
- Modify: `cmd/trader/main.go`

**Step 1: Write the full main.go**

This is the glue that connects all components. No separate test — integration tested by `make run` with `DRY_RUN=true`.

```go
// cmd/trader/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"

	"github.com/GoPolymarket/polymarket-trader/internal/config"
	"github.com/GoPolymarket/polymarket-trader/internal/feed"
	"github.com/GoPolymarket/polymarket-trader/internal/risk"
	"github.com/GoPolymarket/polymarket-trader/internal/strategy"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Load config
	cfg, err := config.LoadFile(*cfgPath)
	if err != nil {
		log.Printf("warning: config file: %v, using defaults", err)
		cfg = config.Default()
	}
	cfg.ApplyEnv()

	if cfg.PrivateKey == "" || cfg.APIKey == "" {
		log.Fatal("POLYMARKET_PK and POLYMARKET_API_KEY are required")
	}

	log.Printf("polymarket-trader starting (dry_run=%t)", cfg.DryRun)

	// Init SDK client
	signer, err := auth.NewPrivateKeySigner(strings.TrimSpace(cfg.PrivateKey), 137)
	if err != nil {
		log.Fatalf("signer: %v", err)
	}

	apiKey := &auth.APIKey{
		Key:        strings.TrimSpace(cfg.APIKey),
		Secret:     strings.TrimSpace(cfg.APISecret),
		Passphrase: strings.TrimSpace(cfg.APIPassphrase),
	}

	sdkClient := polymarket.NewClient()
	clobClient := sdkClient.CLOB.WithAuth(signer, apiKey)

	// Attach Builder Code if configured
	if cfg.BuilderKey != "" && cfg.BuilderSecret != "" {
		clobClient = clobClient.WithBuilderConfig(&auth.BuilderConfig{
			Local: &auth.BuilderCredentials{
				Key:        strings.TrimSpace(cfg.BuilderKey),
				Secret:     strings.TrimSpace(cfg.BuilderSecret),
				Passphrase: strings.TrimSpace(cfg.BuilderPassphrase),
			},
		})
		log.Println("builder attribution enabled")
	}

	// Init WS + authenticate
	wsClient := sdkClient.CLOBWS.Authenticate(signer, apiKey)

	// Init components
	books := feed.NewBookSnapshot()
	riskMgr := risk.New(risk.Config{
		MaxOpenOrders:        cfg.Risk.MaxOpenOrders,
		MaxDailyLossUSDC:     cfg.Risk.MaxDailyLossUSDC,
		MaxPositionPerMarket: cfg.Risk.MaxPositionPerMarket,
	})
	maker := strategy.NewMaker(strategy.MakerConfig{
		MinSpreadBps:       cfg.Maker.MinSpreadBps,
		SpreadMultiplier:   cfg.Maker.SpreadMultiplier,
		OrderSizeUSDC:      cfg.Maker.OrderSizeUSDC,
		MaxOrdersPerMarket: cfg.Maker.MaxOrdersPerMarket,
	})
	taker := strategy.NewTaker(strategy.TakerConfig{
		MinImbalance:   cfg.Taker.MinImbalance,
		DepthLevels:    cfg.Taker.DepthLevels,
		AmountUSDC:     cfg.Taker.AmountUSDC,
		MaxSlippageBps: cfg.Taker.MaxSlippageBps,
		Cooldown:       cfg.Taker.Cooldown,
	})

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Select markets
	assetIDs := cfg.Maker.Markets
	if len(assetIDs) == 0 {
		log.Println("auto-selecting markets...")
		assetIDs, err = autoSelectMarkets(ctx, clobClient, cfg.Maker.AutoSelectTop)
		if err != nil {
			log.Fatalf("market selection: %v", err)
		}
	}
	if len(assetIDs) == 0 {
		log.Fatal("no markets selected")
	}
	log.Printf("monitoring %d assets: %v", len(assetIDs), assetIDs)

	// Subscribe to orderbook WS
	bookCh, err := wsClient.SubscribeOrderbook(ctx, assetIDs)
	if err != nil {
		log.Fatalf("ws subscribe: %v", err)
	}

	// Track active maker orders for cancel-replace
	activeOrders := make(map[string][]string) // assetID → []orderID

	log.Println("trading loop started")

	// Stats
	var totalOrders, totalFills int

	for {
		select {
		case <-sigCh:
			log.Println("shutdown signal received")
			goto shutdown
		case event, ok := <-bookCh:
			if !ok {
				log.Println("book channel closed, reconnecting...")
				time.Sleep(2 * time.Second)
				bookCh, err = wsClient.SubscribeOrderbook(ctx, assetIDs)
				if err != nil {
					log.Printf("reconnect failed: %v", err)
					goto shutdown
				}
				continue
			}
			books.Update(event)

			// --- Maker ---
			if cfg.Maker.Enabled {
				quote, err := maker.ComputeQuote(event)
				if err != nil {
					continue
				}
				// Cancel old orders for this asset
				if old, ok := activeOrders[event.AssetID]; ok && len(old) > 0 {
					_, _ = clobClient.CancelOrders(ctx, &clobtypes.CancelOrdersRequest{OrderIDs: old})
					delete(activeOrders, event.AssetID)
				}

				if !cfg.DryRun {
					if err := riskMgr.Allow(event.AssetID, quote.Size); err != nil {
						continue
					}
					// Place buy
					buyResp := placeLimit(ctx, clobClient, signer, event.AssetID, "BUY", quote.BuyPrice, quote.Size)
					if buyResp.ID != "" {
						activeOrders[event.AssetID] = append(activeOrders[event.AssetID], buyResp.ID)
						totalOrders++
					}
					// Place sell
					sellResp := placeLimit(ctx, clobClient, signer, event.AssetID, "SELL", quote.SellPrice, quote.Size)
					if sellResp.ID != "" {
						activeOrders[event.AssetID] = append(activeOrders[event.AssetID], sellResp.ID)
						totalOrders++
					}
				} else {
					log.Printf("[DRY] maker %s: buy=%.4f sell=%.4f size=%.2f",
						event.AssetID, quote.BuyPrice, quote.SellPrice, quote.Size)
				}
			}

			// --- Taker ---
			if cfg.Taker.Enabled {
				sig, err := taker.Evaluate(event)
				if err != nil || sig == nil {
					continue
				}
				if !cfg.DryRun {
					if err := riskMgr.Allow(event.AssetID, sig.AmountUSDC); err != nil {
						continue
					}
					resp := placeMarket(ctx, clobClient, signer, sig.AssetID, sig.Side, sig.AmountUSDC, sig.MaxPrice)
					if resp.ID != "" {
						taker.RecordTrade(sig.AssetID)
						totalFills++
					}
				} else {
					log.Printf("[DRY] taker %s: side=%s amount=%.2f imbalance=%.4f",
						sig.AssetID, sig.Side, sig.AmountUSDC, sig.Imbalance)
				}
			}
		}
	}

shutdown:
	log.Println("shutting down...")
	if !cfg.DryRun {
		log.Println("cancelling all open orders...")
		resp, err := clobClient.CancelAll(ctx)
		if err != nil {
			log.Printf("cancel all error: %v", err)
		} else {
			log.Printf("cancelled %d orders", resp.Count)
		}
	}
	_ = wsClient.Close()
	log.Printf("session complete: orders=%d fills=%d pnl=%.2f", totalOrders, totalFills, riskMgr.DailyPnL())
}

func autoSelectMarkets(ctx context.Context, client clob.Client, topN int) ([]string, error) {
	resp, err := client.Markets(ctx, &clobtypes.MarketsRequest{Active: boolPtr(true), Limit: 50})
	if err != nil {
		return nil, err
	}
	// Fetch books for each token
	booksMap := make(map[string]clobtypes.OrderBook)
	for _, m := range resp.Data {
		for _, tok := range m.Tokens {
			book, err := client.OrderBook(ctx, &clobtypes.BookRequest{TokenID: tok.TokenID})
			if err != nil {
				continue
			}
			booksMap[tok.TokenID] = clobtypes.OrderBook(book)
		}
	}
	return strategy.SelectMarkets(resp.Data, booksMap, topN, 50), nil
}

func placeLimit(ctx context.Context, client clob.Client, signer auth.Signer, tokenID, side string, price, sizeUSDC float64) clobtypes.OrderResponse {
	builder := clob.NewOrderBuilder(client, signer).
		TokenID(tokenID).
		Side(side).
		Price(price).
		AmountUSDC(sizeUSDC).
		OrderType(clobtypes.OrderTypeGTC)

	signable, err := builder.BuildSignableWithContext(ctx)
	if err != nil {
		log.Printf("build limit %s %s: %v", side, tokenID, err)
		return clobtypes.OrderResponse{}
	}
	resp, err := client.CreateOrderFromSignable(ctx, signable)
	if err != nil {
		log.Printf("place limit %s %s: %v", side, tokenID, err)
		return clobtypes.OrderResponse{}
	}
	log.Printf("limit %s %s @ %.4f: id=%s", side, tokenID, price, resp.ID)
	return resp
}

func placeMarket(ctx context.Context, client clob.Client, signer auth.Signer, tokenID, side string, amountUSDC, maxPrice float64) clobtypes.OrderResponse {
	builder := clob.NewOrderBuilder(client, signer).
		TokenID(tokenID).
		Side(side).
		AmountUSDC(amountUSDC).
		OrderType(clobtypes.OrderTypeFAK)

	signable, err := builder.BuildMarketWithContext(ctx)
	if err != nil {
		log.Printf("build market %s %s: %v", side, tokenID, err)
		return clobtypes.OrderResponse{}
	}
	resp, err := client.CreateOrderFromSignable(ctx, signable)
	if err != nil {
		log.Printf("place market %s %s: %v", side, tokenID, err)
		return clobtypes.OrderResponse{}
	}
	log.Printf("market %s %s amount=%.2f: id=%s", side, tokenID, amountUSDC, resp.ID)
	return resp
}

func boolPtr(v bool) *bool { return &v }
```

**Step 2: Verify it compiles**

Run: `go build ./cmd/trader/`
Expected: no errors

**Step 3: Commit**

```bash
git add cmd/trader/main.go
git commit -m "feat: add main entry point with full trading loop"
```

---

### Task 9: Dockerfile & docker-compose

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.env.example`
- Create: `.gitignore`

**Step 1: Create Dockerfile**

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /trader ./cmd/trader/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /trader /trader
COPY config.yaml /config.yaml
ENTRYPOINT ["/trader"]
```

**Step 2: Create docker-compose.yml**

```yaml
services:
  trader:
    build: .
    restart: unless-stopped
    env_file: .env
    volumes:
      - ./config.yaml:/config.yaml:ro
```

**Step 3: Create .env.example**

```env
POLYMARKET_PK=0xYOUR_PRIVATE_KEY
POLYMARKET_API_KEY=your-api-key
POLYMARKET_API_SECRET=your-api-secret
POLYMARKET_API_PASSPHRASE=your-api-passphrase
BUILDER_KEY=your-builder-key
BUILDER_SECRET=your-builder-secret
BUILDER_PASSPHRASE=your-builder-passphrase
TRADER_DRY_RUN=true
```

**Step 4: Create .gitignore**

```gitignore
bin/
.env
*.out
```

**Step 5: Verify Docker build**

Run: `docker build -t polymarket-trader .`
Expected: successful build

**Step 6: Commit**

```bash
git add Dockerfile docker-compose.yml .env.example .gitignore
git commit -m "feat: add Docker deployment setup"
```

---

### Task 10: Dry-Run Smoke Test

**Step 1: Run with dry_run=true**

```bash
POLYMARKET_PK=0x_test_key \
POLYMARKET_API_KEY=test \
POLYMARKET_API_SECRET=test \
POLYMARKET_API_PASSPHRASE=test \
TRADER_DRY_RUN=true \
go run ./cmd/trader/ -config config.yaml
```

Expected: starts up, selects markets, prints `[DRY]` log lines, no actual orders placed.

**Step 2: Run all unit tests**

```bash
go test ./... -v -race -count=1
```

Expected: all PASS

**Step 3: Final commit**

```bash
git add -A
git commit -m "chore: verify dry-run smoke test passes"
```

---

## Summary

| Task | What | Est. |
|------|------|------|
| 1 | Scaffold project | 5 min |
| 2 | Config package | 10 min |
| 3 | Feed (book snapshot) | 10 min |
| 4 | Risk manager | 10 min |
| 5 | Maker strategy | 10 min |
| 6 | Taker strategy | 10 min |
| 7 | Market selector | 5 min |
| 8 | Main entry point | 15 min |
| 9 | Docker deployment | 5 min |
| 10 | Smoke test | 5 min |

**Total: ~10 tasks, ~85 minutes of implementation**

After this plan is complete, the go-live path is:
1. Set real API keys in `.env`
2. Set `TRADER_DRY_RUN=false`, `order_size_usdc: 5` (small amounts)
3. `docker-compose up -d`
4. Verify volume on `builders.polymarket.com`
5. Scale up amounts, apply for grant
