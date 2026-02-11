package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Config controls scanning, risk, and execution behavior.
type Config struct {
	ScanLimit             int
	TopN                  int
	MinSpreadBps          decimal.Decimal
	MinBookDepthShares    decimal.Decimal
	MinImbalance          decimal.Decimal
	DefaultAmountUSDC     decimal.Decimal
	MaxPerTradeUSDC       decimal.Decimal
	MaxDailyLossUSDC      decimal.Decimal
	MaxOpenTrades         int
	MaxSlippageBps        decimal.Decimal
	MinConfidenceBps      decimal.Decimal
	RequestTimeout        time.Duration
	DryRun                bool
	AllowExecution        bool
	RequireExplicitConfirm bool
}

func DefaultConfig() Config {
	return Config{
		ScanLimit:              60,
		TopN:                   8,
		MinSpreadBps:           decimal.NewFromInt(15),
		MinBookDepthShares:     decimal.NewFromInt(100),
		MinImbalance:           decimal.RequireFromString("0.08"),
		DefaultAmountUSDC:      decimal.NewFromInt(25),
		MaxPerTradeUSDC:        decimal.NewFromInt(100),
		MaxDailyLossUSDC:       decimal.NewFromInt(150),
		MaxOpenTrades:          5,
		MaxSlippageBps:         decimal.NewFromInt(25),
		MinConfidenceBps:       decimal.NewFromInt(25),
		RequestTimeout:         12 * time.Second,
		DryRun:                 true,
		AllowExecution:         false,
		RequireExplicitConfirm: true,
	}
}

func (c Config) Validate() error {
	if c.ScanLimit <= 0 {
		return fmt.Errorf("scan limit must be > 0")
	}
	if c.TopN <= 0 {
		return fmt.Errorf("top-n must be > 0")
	}
	if c.DefaultAmountUSDC.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("default amount must be > 0")
	}
	if c.MaxPerTradeUSDC.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max per trade must be > 0")
	}
	if c.MaxDailyLossUSDC.LessThan(decimal.Zero) {
		return fmt.Errorf("max daily loss must be >= 0")
	}
	if c.MaxOpenTrades <= 0 {
		return fmt.Errorf("max open trades must be > 0")
	}
	if c.MaxSlippageBps.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max slippage bps must be > 0")
	}
	if c.MinConfidenceBps.LessThan(decimal.Zero) {
		return fmt.Errorf("min confidence bps must be >= 0")
	}
	return nil
}

// MergeEnv allows easy ops without recompiling.
func (c Config) MergeEnv() Config {
	if v := strings.TrimSpace(os.Getenv("BOT_SCAN_LIMIT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.ScanLimit = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("BOT_TOP_N")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.TopN = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("BOT_DEFAULT_AMOUNT_USDC")); v != "" {
		if d, err := decimal.NewFromString(v); err == nil && d.GreaterThan(decimal.Zero) {
			c.DefaultAmountUSDC = d
		}
	}
	if v := strings.TrimSpace(os.Getenv("BOT_MAX_PER_TRADE_USDC")); v != "" {
		if d, err := decimal.NewFromString(v); err == nil && d.GreaterThan(decimal.Zero) {
			c.MaxPerTradeUSDC = d
		}
	}
	if v := strings.TrimSpace(os.Getenv("BOT_MAX_SLIPPAGE_BPS")); v != "" {
		if d, err := decimal.NewFromString(v); err == nil && d.GreaterThan(decimal.Zero) {
			c.MaxSlippageBps = d
		}
	}
	if v := strings.TrimSpace(os.Getenv("BOT_DRY_RUN")); v != "" {
		c.DryRun = strings.EqualFold(v, "1") || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}
	return c
}
