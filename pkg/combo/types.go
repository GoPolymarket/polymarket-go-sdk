package combo

import (
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/gamma"
)

// ComboMarketsRequest represents the query parameters for listing combo markets.
type ComboMarketsRequest struct {
	Limit   *int   `json:"limit,omitempty"`
	Offset  *int   `json:"offset,omitempty"`
	Active  *bool  `json:"active,omitempty"`
	Closed  *bool  `json:"closed,omitempty"`
	TagID   string `json:"tag_id,omitempty"`
	TagSlug string `json:"tag_slug,omitempty"`
}

// ComboMarket represents a combo market from the Gamma API.
type ComboMarket struct {
	ID            string       `json:"id"`
	Question      string       `json:"question"`
	ConditionIDs  []string     `json:"conditionIds"`
	Slug          string       `json:"slug"`
	StartDate     string       `json:"startDate"`
	EndDate       string       `json:"endDate"`
	Tags          []gamma.Tag  `json:"tags"`
	Active        bool         `json:"active"`
	Closed        bool         `json:"closed"`
}
