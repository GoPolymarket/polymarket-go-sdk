package rfq

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
)

type Client interface {
	CreateRFQRequest(ctx context.Context, req *RFQRequest) (RFQRequestResponse, error)
	CancelRFQRequest(ctx context.Context, req *RFQCancelRequest) (RFQCancelResponse, error)
	RFQRequests(ctx context.Context, req *RFQRequestsQuery) (RFQRequestsResponse, error)
	CreateRFQQuote(ctx context.Context, req *RFQQuote) (RFQQuoteResponse, error)
	CancelRFQQuote(ctx context.Context, req *RFQCancelQuote) (RFQCancelResponse, error)
	RFQQuotes(ctx context.Context, req *RFQQuotesQuery) (RFQQuotesResponse, error)
	RFQBestQuote(ctx context.Context, req *RFQBestQuoteQuery) (RFQBestQuoteResponse, error)
	RFQRequestAccept(ctx context.Context, req *RFQAcceptRequest) (RFQAcceptResponse, error)
	RFQQuoteApprove(ctx context.Context, req *RFQApproveQuote) (RFQApproveResponse, error)
	RFQConfig(ctx context.Context) (RFQConfigResponse, error)
}

type clientImpl struct {
	httpClient *transport.Client
}

func NewClient(httpClient *transport.Client) Client {
	return &clientImpl{httpClient: httpClient}
}

func (c *clientImpl) CreateRFQRequest(ctx context.Context, req *RFQRequest) (RFQRequestResponse, error) {
	var resp RFQRequestResponse
	err := c.httpClient.Post(ctx, "/rfq/request", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelRFQRequest(ctx context.Context, req *RFQCancelRequest) (RFQCancelResponse, error) {
	var resp RFQCancelResponse
	err := c.httpClient.Delete(ctx, "/rfq/request", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQRequests(ctx context.Context, req *RFQRequestsQuery) (RFQRequestsResponse, error) {
	var resp RFQRequestsResponse
	q := url.Values{}
	if req != nil {
		applyRFQPagination(&q, req.Limit, req.Offset, req.Cursor)
		applyRFQFilters(&q, req.State, req.RequestIDs, nil, req.Markets, req.SizeMin, req.SizeMax, req.SizeUsdcMin, req.SizeUsdcMax, req.PriceMin, req.PriceMax, req.SortBy, req.SortDir)
	}
	err := c.httpClient.Get(ctx, "/rfq/data/requests", q, &resp)
	return resp, err
}

func (c *clientImpl) CreateRFQQuote(ctx context.Context, req *RFQQuote) (RFQQuoteResponse, error) {
	var resp RFQQuoteResponse
	err := c.httpClient.Post(ctx, "/rfq/quote", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelRFQQuote(ctx context.Context, req *RFQCancelQuote) (RFQCancelResponse, error) {
	var resp RFQCancelResponse
	err := c.httpClient.Delete(ctx, "/rfq/quote", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQQuotes(ctx context.Context, req *RFQQuotesQuery) (RFQQuotesResponse, error) {
	var resp RFQQuotesResponse
	q := url.Values{}
	if req != nil {
		applyRFQPagination(&q, req.Limit, req.Offset, req.Cursor)
		applyRFQFilters(&q, req.State, req.RequestIDs, req.QuoteIDs, req.Markets, req.SizeMin, req.SizeMax, req.SizeUsdcMin, req.SizeUsdcMax, req.PriceMin, req.PriceMax, req.SortBy, req.SortDir)
	}
	err := c.httpClient.Get(ctx, "/rfq/data/quotes", q, &resp)
	return resp, err
}

func (c *clientImpl) RFQBestQuote(ctx context.Context, req *RFQBestQuoteQuery) (RFQBestQuoteResponse, error) {
	var resp RFQBestQuoteResponse
	q := url.Values{}
	if req != nil {
		requestIDs := req.RequestIDs
		if req.RequestID != "" {
			q.Set("request_id", req.RequestID)
			if len(requestIDs) == 0 {
				requestIDs = []string{req.RequestID}
			}
		}
		if len(requestIDs) > 0 {
			q.Set("requestIds", strings.Join(requestIDs, ","))
		}
	}
	err := c.httpClient.Get(ctx, "/rfq/data/best-quote", q, &resp)
	return resp, err
}

func (c *clientImpl) RFQRequestAccept(ctx context.Context, req *RFQAcceptRequest) (RFQAcceptResponse, error) {
	var resp RFQAcceptResponse
	err := c.httpClient.Post(ctx, "/rfq/request/accept", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQQuoteApprove(ctx context.Context, req *RFQApproveQuote) (RFQApproveResponse, error) {
	var resp RFQApproveResponse
	err := c.httpClient.Post(ctx, "/rfq/quote/approve", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQConfig(ctx context.Context) (RFQConfigResponse, error) {
	var resp RFQConfigResponse
	err := c.httpClient.Get(ctx, "/rfq/config", nil, &resp)
	return resp, err
}

func applyRFQPagination(q *url.Values, limit int, offset, cursor string) {
	if q == nil {
		return
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset == "" && cursor != "" {
		offset = cursor
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if offset != "" {
		q.Set("offset", offset)
	}
}

func applyRFQFilters(q *url.Values, state RFQState, requestIDs, quoteIDs, markets []string, sizeMin, sizeMax, sizeUsdcMin, sizeUsdcMax, priceMin, priceMax string, sortBy RFQSortBy, sortDir RFQSortDir) {
	if q == nil {
		return
	}
	if state != "" {
		q.Set("state", string(state))
	}
	if len(requestIDs) > 0 {
		q.Set("request_ids", strings.Join(requestIDs, ","))
	}
	if len(quoteIDs) > 0 {
		q.Set("quote_ids", strings.Join(quoteIDs, ","))
	}
	if len(markets) > 0 {
		q.Set("markets", strings.Join(markets, ","))
	}
	if sizeMin != "" {
		q.Set("size_min", sizeMin)
	}
	if sizeMax != "" {
		q.Set("size_max", sizeMax)
	}
	if sizeUsdcMin != "" {
		q.Set("size_usdc_min", sizeUsdcMin)
	}
	if sizeUsdcMax != "" {
		q.Set("size_usdc_max", sizeUsdcMax)
	}
	if priceMin != "" {
		q.Set("price_min", priceMin)
	}
	if priceMax != "" {
		q.Set("price_max", priceMax)
	}
	if sortBy != "" {
		q.Set("sort_by", string(sortBy))
	}
	if sortDir != "" {
		q.Set("sort_dir", string(sortDir))
	}
}
