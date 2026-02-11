package bot

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestSlippageGuardPrice(t *testing.T) {
	op := Opportunity{Mid: decimal.RequireFromString("0.50"), Recommended: "BUY"}
	got := slippageGuardPrice(op, decimal.NewFromInt(20))
	want := decimal.RequireFromString("0.501")
	if !got.Equal(want) {
		t.Fatalf("buy guard mismatch got=%s want=%s", got, want)
	}

	op = Opportunity{Mid: decimal.RequireFromString("0.50"), Recommended: "SELL"}
	got = slippageGuardPrice(op, decimal.NewFromInt(20))
	want = decimal.RequireFromString("0.499")
	if !got.Equal(want) {
		t.Fatalf("sell guard mismatch got=%s want=%s", got, want)
	}
}
