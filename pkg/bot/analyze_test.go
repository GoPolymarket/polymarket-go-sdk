package bot

import (
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/shopspring/decimal"
)

func TestTopOfBook(t *testing.T) {
	price, depth, err := topOfBook([]clobtypes.PriceLevel{{Price: "0.53", Size: "220.12"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !price.Equal(decimal.RequireFromString("0.53")) {
		t.Fatalf("unexpected price: %s", price)
	}
	if !depth.Equal(decimal.RequireFromString("220.12")) {
		t.Fatalf("unexpected depth: %s", depth)
	}
}
