// Package combo provides the client for interacting with Polymarket combo markets.
// Combo markets combine multiple conditions into a single prediction market.
package combo

import (
	"context"
)

// Client defines the interface for the Polymarket Combo Markets API.
type Client interface {
	// ComboMarkets retrieves a list of combo markets.
	ComboMarkets(ctx context.Context, req *ComboMarketsRequest) ([]ComboMarket, error)
}
