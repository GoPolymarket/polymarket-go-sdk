package clob

import (
	"context"
	"fmt"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

type stubClient struct {
	*clientImpl

	tickSize      float64
	feeRate       int64
	book          clobtypes.OrderBookResponse
	orders        map[string]clobtypes.OrdersResponse
	trades        map[string]clobtypes.TradesResponse
	builderTrades map[string]clobtypes.BuilderTradesResponse
}

func newStubClient() *stubClient {
	return &stubClient{
		clientImpl:    &clientImpl{},
		orders:        make(map[string]clobtypes.OrdersResponse),
		trades:        make(map[string]clobtypes.TradesResponse),
		builderTrades: make(map[string]clobtypes.BuilderTradesResponse),
	}
}

func (s *stubClient) OrderBook(ctx context.Context, req *clobtypes.BookRequest) (clobtypes.OrderBookResponse, error) {
	return s.book, nil
}

func (s *stubClient) TickSize(ctx context.Context, req *clobtypes.TickSizeRequest) (clobtypes.TickSizeResponse, error) {
	return clobtypes.TickSizeResponse{MinimumTickSize: s.tickSize}, nil
}

func (s *stubClient) FeeRate(ctx context.Context, req *clobtypes.FeeRateRequest) (clobtypes.FeeRateResponse, error) {
	return clobtypes.FeeRateResponse{BaseFee: int(s.feeRate)}, nil
}

func (s *stubClient) Orders(ctx context.Context, req *clobtypes.OrdersRequest) (clobtypes.OrdersResponse, error) {
	cursor := cursorFromOrdersRequest(req)
	resp, ok := s.orders[cursor]
	if !ok {
		return clobtypes.OrdersResponse{}, fmt.Errorf("unexpected orders cursor %q", cursor)
	}
	return resp, nil
}

func (s *stubClient) Trades(ctx context.Context, req *clobtypes.TradesRequest) (clobtypes.TradesResponse, error) {
	cursor := cursorFromTradesRequest(req)
	resp, ok := s.trades[cursor]
	if !ok {
		return clobtypes.TradesResponse{}, fmt.Errorf("unexpected trades cursor %q", cursor)
	}
	return resp, nil
}

func (s *stubClient) BuilderTrades(ctx context.Context, req *clobtypes.BuilderTradesRequest) (clobtypes.BuilderTradesResponse, error) {
	cursor := cursorFromBuilderTradesRequest(req)
	resp, ok := s.builderTrades[cursor]
	if !ok {
		return clobtypes.BuilderTradesResponse{}, fmt.Errorf("unexpected builder trades cursor %q", cursor)
	}
	return resp, nil
}

func cursorFromOrdersRequest(req *clobtypes.OrdersRequest) string {
	if req == nil {
		return clobtypes.InitialCursor
	}
	if req.NextCursor != "" {
		return req.NextCursor
	}
	if req.Cursor != "" {
		return req.Cursor
	}
	return clobtypes.InitialCursor
}

func cursorFromTradesRequest(req *clobtypes.TradesRequest) string {
	if req == nil {
		return clobtypes.InitialCursor
	}
	if req.NextCursor != "" {
		return req.NextCursor
	}
	if req.Cursor != "" {
		return req.Cursor
	}
	return clobtypes.InitialCursor
}

func cursorFromBuilderTradesRequest(req *clobtypes.BuilderTradesRequest) string {
	if req == nil {
		return clobtypes.InitialCursor
	}
	if req.NextCursor != "" {
		return req.NextCursor
	}
	if req.Cursor != "" {
		return req.Cursor
	}
	return clobtypes.InitialCursor
}
