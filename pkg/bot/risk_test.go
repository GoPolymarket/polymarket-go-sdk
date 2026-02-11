package bot

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidatePlanAgainstRisk(t *testing.T) {
	cfg := DefaultConfig()
	e := &Engine{cfg: cfg}

	plan := &TradePlan{Side: "BUY", AmountUSDC: decimal.NewFromInt(10), MaxAcceptedPrice: decimal.RequireFromString("0.51")}
	risk := RiskSnapshot{CanTrade: true}
	if err := e.ValidatePlanAgainstRisk(plan, risk); err != nil {
		t.Fatalf("expected valid plan, got err: %v", err)
	}

	bad := *plan
	bad.AmountUSDC = cfg.MaxPerTradeUSDC.Add(decimal.NewFromInt(1))
	if err := e.ValidatePlanAgainstRisk(&bad, risk); err == nil {
		t.Fatal("expected per-trade cap error")
	}
}
