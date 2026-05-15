package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
)

type fakeCLOBClient struct {
	createResp clobtypes.OrderResponse
	createErr  error

	cancelResp clobtypes.CancelResponse
	cancelErr  error

	orderResp clobtypes.OrderResponse
	orderErr  error

	ordersResp clobtypes.OrdersResponse
	ordersErr  error

	gotCreateOrder *clobtypes.Order
	gotCancelReq   *clobtypes.CancelOrderRequest
	gotOrderID     string
	gotOrdersReq   *clobtypes.OrdersRequest
}

func (f *fakeCLOBClient) CreateOrder(_ context.Context, order *clobtypes.Order) (clobtypes.OrderResponse, error) {
	f.gotCreateOrder = order
	return f.createResp, f.createErr
}

func (f *fakeCLOBClient) CancelOrder(_ context.Context, req *clobtypes.CancelOrderRequest) (clobtypes.CancelResponse, error) {
	f.gotCancelReq = req
	return f.cancelResp, f.cancelErr
}

func (f *fakeCLOBClient) Order(_ context.Context, id string) (clobtypes.OrderResponse, error) {
	f.gotOrderID = id
	return f.orderResp, f.orderErr
}

func (f *fakeCLOBClient) Orders(_ context.Context, req *clobtypes.OrdersRequest) (clobtypes.OrdersResponse, error) {
	f.gotOrdersReq = req
	return f.ordersResp, f.ordersErr
}

func TestNewCLOBEngineRejectsNilClient(t *testing.T) {
	_, err := NewCLOBEngine(nil)
	if err == nil {
		t.Fatalf("expected error for nil client")
	}
}

func TestCLOBEnginePlaceRequiresOrder(t *testing.T) {
	engine, err := NewCLOBEngine(&fakeCLOBClient{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	_, err = engine.Place(context.Background(), PlaceRequest{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestCLOBEnginePlaceCallsCreateOrder(t *testing.T) {
	fake := &fakeCLOBClient{
		createResp: clobtypes.OrderResponse{ID: "ord-1", Status: "live"},
	}
	engine, err := NewCLOBEngine(fake)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	order := &clobtypes.Order{}
	resp, err := engine.Place(context.Background(), PlaceRequest{
		Order: order,
		Attribution: Attribution{
			Builder: " builder-a ",
			Funder:  " 0xabc ",
			Source:  " PMX-Gateway ",
		},
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	if fake.gotCreateOrder != order {
		t.Fatalf("expected order pointer to be forwarded")
	}
	if resp.Order.ID != "ord-1" {
		t.Fatalf("expected order id ord-1, got %q", resp.Order.ID)
	}
	if resp.Attribution.Builder != "builder-a" || resp.Attribution.Funder != "0xabc" || resp.Attribution.Source != "pmx-gateway" {
		t.Fatalf("expected normalized attribution passthrough, got %+v", resp.Attribution)
	}
}

func TestCLOBEngineCancelRequiresOrderID(t *testing.T) {
	engine, err := NewCLOBEngine(&fakeCLOBClient{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	_, err = engine.Cancel(context.Background(), CancelRequest{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestCLOBEngineCancelCallsCancelOrder(t *testing.T) {
	fake := &fakeCLOBClient{
		cancelResp: clobtypes.CancelResponse{Canceled: []string{"ord-1"}},
	}
	engine, err := NewCLOBEngine(fake)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	resp, err := engine.Cancel(context.Background(), CancelRequest{OrderID: "ord-1"})
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if fake.gotCancelReq == nil || fake.gotCancelReq.OrderID != "ord-1" {
		t.Fatalf("expected cancel request for ord-1")
	}
	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %q", resp.Status)
	}
}

func TestCLOBEngineQueryRequiresOrderID(t *testing.T) {
	engine, err := NewCLOBEngine(&fakeCLOBClient{})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	_, err = engine.Query(context.Background(), QueryRequest{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestCLOBEngineQueryCallsOrder(t *testing.T) {
	fake := &fakeCLOBClient{
		orderResp: clobtypes.OrderResponse{ID: "ord-2", Status: "matched"},
	}
	engine, err := NewCLOBEngine(fake)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	resp, err := engine.Query(context.Background(), QueryRequest{OrderID: "ord-2"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if fake.gotOrderID != "ord-2" {
		t.Fatalf("expected query for ord-2, got %q", fake.gotOrderID)
	}
	if resp.Order.Status != "matched" {
		t.Fatalf("expected matched status, got %q", resp.Order.Status)
	}
}

func TestCLOBEngineReplayCallsOrdersWithFilters(t *testing.T) {
	fake := &fakeCLOBClient{
		ordersResp: clobtypes.OrdersResponse{
			Data:       []clobtypes.OrderResponse{{ID: "ord-1"}, {ID: "ord-2"}},
			Count:      2,
			NextCursor: "cursor-2",
		},
	}
	engine, err := NewCLOBEngine(fake)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	resp, err := engine.Replay(context.Background(), ReplayRequest{
		Market: "market-1",
		Cursor: "cursor-1",
		Limit:  0, // use default
	})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if fake.gotOrdersReq == nil {
		t.Fatalf("expected replay request to call Orders")
	}
	if fake.gotOrdersReq.Market != "market-1" || fake.gotOrdersReq.Cursor != "cursor-1" {
		t.Fatalf("unexpected replay request: %+v", fake.gotOrdersReq)
	}
	if fake.gotOrdersReq.Limit != defaultReplayLimit {
		t.Fatalf("expected default limit %d, got %d", defaultReplayLimit, fake.gotOrdersReq.Limit)
	}
	if len(resp.Orders) != 2 || resp.NextCursor != "cursor-2" {
		t.Fatalf("unexpected replay response: %+v", resp)
	}
}

func TestCLOBEngineForwardsUpstreamErrors(t *testing.T) {
	boom := errors.New("boom")
	fake := &fakeCLOBClient{createErr: boom}
	engine, err := NewCLOBEngine(fake)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	_, err = engine.Place(context.Background(), PlaceRequest{Order: &clobtypes.Order{}})
	if !errors.Is(err, boom) {
		t.Fatalf("expected upstream error, got %v", err)
	}
}
