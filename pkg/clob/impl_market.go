package clob

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func (c *clientImpl) Markets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Cursor != "" {
			q.Set("cursor", req.Cursor)
		}
		if req.Active != nil {
			q.Set("active", strconv.FormatBool(*req.Active))
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
	}

	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/markets", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) MarketsAll(ctx context.Context, req *clobtypes.MarketsRequest) ([]clobtypes.Market, error) {
	var results []clobtypes.Market
	cursor := clobtypes.InitialCursor
	if req != nil && req.Cursor != "" {
		cursor = req.Cursor
	}

	for cursor != clobtypes.EndCursor {
		nextReq := clobtypes.MarketsRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.Cursor = cursor

		resp, err := c.Markets(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Data...)

		if resp.NextCursor == "" || resp.NextCursor == cursor {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}

func (c *clientImpl) Market(ctx context.Context, id string) (clobtypes.MarketResponse, error) {
	var resp clobtypes.MarketResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/markets/%s", id), nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) SimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Cursor != "" {
			q.Set("cursor", req.Cursor)
		}
		if req.Active != nil {
			q.Set("active", strconv.FormatBool(*req.Active))
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
	}
	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/simplified-markets", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) SamplingMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/sampling-markets", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) SamplingSimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/sampling-simplified-markets", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) OrderBook(ctx context.Context, req *clobtypes.BookRequest) (clobtypes.OrderBookResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.Side != "" {
			q.Set("side", req.Side)
		}
	}
	var resp clobtypes.OrderBookResponse
	err := c.httpClient.Get(ctx, "/book", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) OrderBooks(ctx context.Context, req *clobtypes.BooksRequest) (clobtypes.OrderBooksResponse, error) {
	var resp clobtypes.OrderBooksResponse
	var body interface{}
	if req != nil {
		if len(req.Requests) > 0 {
			body = req.Requests
		} else if len(req.TokenIDs) > 0 {
			requests := make([]clobtypes.BookRequest, 0, len(req.TokenIDs))
			for _, id := range req.TokenIDs {
				requests = append(requests, clobtypes.BookRequest{TokenID: id})
			}
			body = requests
		}
	}
	err := c.httpClient.Post(ctx, "/books", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Midpoint(ctx context.Context, req *clobtypes.MidpointRequest) (clobtypes.MidpointResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp clobtypes.MidpointResponse
	err := c.httpClient.Get(ctx, "/midpoint", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Midpoints(ctx context.Context, req *clobtypes.MidpointsRequest) (clobtypes.MidpointsResponse, error) {
	var resp clobtypes.MidpointsResponse
	var body []map[string]string
	if req != nil {
		body = make([]map[string]string, 0, len(req.TokenIDs))
		for _, id := range req.TokenIDs {
			body = append(body, map[string]string{"token_id": id})
		}
	}
	err := c.httpClient.Post(ctx, "/midpoints", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Price(ctx context.Context, req *clobtypes.PriceRequest) (clobtypes.PriceResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.Side != "" {
			q.Set("side", req.Side)
		}
	}
	var resp clobtypes.PriceResponse
	err := c.httpClient.Get(ctx, "/price", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Prices(ctx context.Context, req *clobtypes.PricesRequest) (clobtypes.PricesResponse, error) {
	var resp clobtypes.PricesResponse
	var body interface{}
	if req != nil {
		if len(req.Requests) > 0 {
			body = req.Requests
		} else if len(req.TokenIDs) > 0 {
			requests := make([]clobtypes.PriceRequest, 0, len(req.TokenIDs))
			for _, id := range req.TokenIDs {
				requests = append(requests, clobtypes.PriceRequest{TokenID: id, Side: req.Side})
			}
			body = requests
		}
	}
	err := c.httpClient.Post(ctx, "/prices", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) AllPrices(ctx context.Context) (clobtypes.PricesResponse, error) {
	var resp clobtypes.PricesResponse
	err := c.httpClient.Get(ctx, "/prices", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Spread(ctx context.Context, req *clobtypes.SpreadRequest) (clobtypes.SpreadResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.Side != "" {
			q.Set("side", req.Side)
		}
	}
	var resp clobtypes.SpreadResponse
	err := c.httpClient.Get(ctx, "/spread", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Spreads(ctx context.Context, req *clobtypes.SpreadsRequest) (clobtypes.SpreadsResponse, error) {
	var resp clobtypes.SpreadsResponse
	var body interface{}
	if req != nil {
		if len(req.Requests) > 0 {
			body = req.Requests
		} else if len(req.TokenIDs) > 0 {
			requests := make([]clobtypes.SpreadRequest, 0, len(req.TokenIDs))
			for _, id := range req.TokenIDs {
				requests = append(requests, clobtypes.SpreadRequest{TokenID: id})
			}
			body = requests
		}
	}
	err := c.httpClient.Post(ctx, "/spreads", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) LastTradePrice(ctx context.Context, req *clobtypes.LastTradePriceRequest) (clobtypes.LastTradePriceResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp clobtypes.LastTradePriceResponse
	err := c.httpClient.Get(ctx, "/last-trade-price", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) LastTradesPrices(ctx context.Context, req *clobtypes.LastTradesPricesRequest) (clobtypes.LastTradesPricesResponse, error) {
	var resp clobtypes.LastTradesPricesResponse
	var body []map[string]string
	if req != nil {
		body = make([]map[string]string, 0, len(req.TokenIDs))
		for _, id := range req.TokenIDs {
			body = append(body, map[string]string{"token_id": id})
		}
	}
	err := c.httpClient.Post(ctx, "/last-trades-prices", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) TickSize(ctx context.Context, req *clobtypes.TickSizeRequest) (clobtypes.TickSizeResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.tickSizes[req.TokenID]; ok && cached != 0 {
			c.cache.mu.RUnlock()
			return clobtypes.TickSizeResponse{MinimumTickSize: cached}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp clobtypes.TickSizeResponse
	err := c.httpClient.Get(ctx, "/tick-size", q, &resp)
	if err == nil && req != nil && req.TokenID != "" && c.cache != nil {
		tickSize := resp.MinimumTickSize
		if tickSize == 0 {
			tickSize = resp.TickSize
		}
		if tickSize != 0 {
			c.cache.mu.Lock()
			c.cache.tickSizes[req.TokenID] = tickSize
			c.cache.mu.Unlock()
		}
	}
	return resp, mapError(err)
}

func (c *clientImpl) NegRisk(ctx context.Context, req *clobtypes.NegRiskRequest) (clobtypes.NegRiskResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.negRisk[req.TokenID]; ok {
			c.cache.mu.RUnlock()
			return clobtypes.NegRiskResponse{NegRisk: cached}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp clobtypes.NegRiskResponse
	err := c.httpClient.Get(ctx, "/neg-risk", q, &resp)
	if err == nil && req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.Lock()
		c.cache.negRisk[req.TokenID] = resp.NegRisk
		c.cache.mu.Unlock()
	}
	return resp, mapError(err)
}

func (c *clientImpl) FeeRate(ctx context.Context, req *clobtypes.FeeRateRequest) (clobtypes.FeeRateResponse, error) {
	q := url.Values{}
	if req != nil && req.TokenID != "" {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.feeRates[req.TokenID]; ok {
			c.cache.mu.RUnlock()
			return clobtypes.FeeRateResponse{BaseFee: int(cached)}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp clobtypes.FeeRateResponse
	err := c.httpClient.Get(ctx, "/fee-rate", q, &resp)
	if err == nil && req != nil && req.TokenID != "" && c.cache != nil {
		fee := int64(resp.BaseFee)
		if fee == 0 && resp.FeeRate != "" {
			if parsed, parseErr := strconv.ParseInt(resp.FeeRate, 10, 64); parseErr == nil {
				fee = parsed
			}
		}
		if fee > 0 {
			c.cache.mu.Lock()
			c.cache.feeRates[req.TokenID] = fee
			c.cache.mu.Unlock()
		}
	}
	return resp, mapError(err)
}

func (c *clientImpl) PricesHistory(ctx context.Context, req *clobtypes.PricesHistoryRequest) (clobtypes.PricesHistoryResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Market != "" {
			q.Set("market", req.Market)
		} else if req.TokenID != "" {
			q.Set("token_id", req.TokenID)
		}

		interval := ""
		if req.Interval != "" {
			interval = string(req.Interval)
		} else if req.Resolution != "" {
			interval = req.Resolution
		}

		if interval != "" {
			q.Set("interval", interval)
		} else {
			if req.StartTs > 0 {
				q.Set("start_ts", strconv.FormatInt(req.StartTs, 10))
			}
			if req.EndTs > 0 {
				q.Set("end_ts", strconv.FormatInt(req.EndTs, 10))
			}
		}

		if req.Fidelity > 0 {
			q.Set("fidelity", strconv.Itoa(req.Fidelity))
		}
	}
	var resp clobtypes.PricesHistoryResponse
	err := c.httpClient.Get(ctx, "/prices-history", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) MarketTradesEvents(ctx context.Context, id string) (clobtypes.MarketTradesEventsResponse, error) {
	var resp clobtypes.MarketTradesEventsResponse
	err := c.httpClient.Get(ctx, "/v1/market-trades-events/"+id, nil, &resp)
	return resp, mapError(err)
}
