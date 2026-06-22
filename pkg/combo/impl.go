// Package combo provides the implementation for the Polymarket Combo Markets API.
package combo

import (
	"context"
	"net/url"
	"strconv"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/gamma"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
)

type clientImpl struct {
	httpClient *transport.Client
}

// NewClient creates a new Combo Markets API client.
func NewClient(httpClient *transport.Client) Client {
	if httpClient == nil {
		httpClient = transport.NewClient(nil, gamma.BaseURL)
	}
	return &clientImpl{
		httpClient: httpClient,
	}
}

func addInt(q url.Values, key string, val *int) {
	if val != nil {
		q.Set(key, strconv.Itoa(*val))
	}
}

func addBool(q url.Values, key string, val *bool) {
	if val != nil {
		q.Set(key, strconv.FormatBool(*val))
	}
}

func addString(q url.Values, key, val string) {
	if val != "" {
		q.Set(key, val)
	}
}

func (c *clientImpl) ComboMarkets(ctx context.Context, req *ComboMarketsRequest) ([]ComboMarket, error) {
	q := url.Values{}
	if req != nil {
		addInt(q, "limit", req.Limit)
		addInt(q, "offset", req.Offset)
		addBool(q, "active", req.Active)
		addBool(q, "closed", req.Closed)
		addString(q, "tag_id", req.TagID)
		addString(q, "tag_slug", req.TagSlug)
	}
	var resp []ComboMarket
	err := c.httpClient.Get(ctx, "/combo-markets", q, &resp)
	return resp, err
}
