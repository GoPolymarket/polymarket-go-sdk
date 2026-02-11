package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/shopspring/decimal"
)

// RiskSnapshot tracks live constraints before an order can be sent.
type RiskSnapshot struct {
	OpenOrders   int
	DailyPnLUSDC decimal.Decimal
	CanTrade     bool
	Reason       string
}

func (e *Engine) EvaluateRisk(ctx context.Context) (RiskSnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.RequestTimeout)
	defer cancel()

	orders, err := e.client.Orders(ctx, &clobtypes.OrdersRequest{Limit: 200})
	if err != nil {
		return RiskSnapshot{}, fmt.Errorf("orders query failed: %w", err)
	}

	open := len(orders.Data)
	if open >= e.cfg.MaxOpenTrades {
		return RiskSnapshot{OpenOrders: open, DailyPnLUSDC: decimal.Zero, CanTrade: false, Reason: "max open trades reached"}, nil
	}

	// PnL is intentionally conservative here: until trade-level pnl attribution is wired,
	// we gate only on open-order count and per-trade sizing caps.
	return RiskSnapshot{OpenOrders: open, DailyPnLUSDC: decimal.Zero, CanTrade: true}, nil
}

func (e *Engine) ValidatePlanAgainstRisk(plan *TradePlan, risk RiskSnapshot) error {
	if plan == nil {
		return fmt.Errorf("plan is required")
	}
	if !risk.CanTrade {
		return fmt.Errorf("risk blocked: %s", risk.Reason)
	}
	if plan.AmountUSDC.GreaterThan(e.cfg.MaxPerTradeUSDC) {
		return fmt.Errorf("amount %s exceeds per-trade cap %s", plan.AmountUSDC.String(), e.cfg.MaxPerTradeUSDC.String())
	}
	if !strings.EqualFold(plan.Side, "BUY") && !strings.EqualFold(plan.Side, "SELL") {
		return fmt.Errorf("unsupported side %q", plan.Side)
	}
	if plan.MaxAcceptedPrice.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max accepted price must be > 0")
	}
	return nil
}
