package execution

import (
	"context"
	"errors"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
)

const defaultReplayLimit = 100

var (
	errNilClient       = errors.New("execution: client is required")
	errOrderRequired   = errors.New("execution: order is required")
	errOrderIDRequired = errors.New("execution: order_id is required")
)

// Engine defines a unified execution contract for strategy/runtime layers.
//
// The contract intentionally stays minimal:
//   - Place submits an order
//   - Cancel cancels an order
//   - Query fetches a single order state
//   - Replay fetches paginated order history for audit/replay loops
type Engine interface {
	Place(ctx context.Context, req PlaceRequest) (PlaceResponse, error)
	Cancel(ctx context.Context, req CancelRequest) (CancelResponse, error)
	Query(ctx context.Context, req QueryRequest) (QueryResponse, error)
	Replay(ctx context.Context, req ReplayRequest) (ReplayResponse, error)
}

// PlaceRequest wraps a CLOB order for submission.
type PlaceRequest struct {
	Order       *clobtypes.Order
	Attribution Attribution
}

// PlaceResponse returns upstream order details after submission.
type PlaceResponse struct {
	Order       clobtypes.OrderResponse
	Attribution Attribution
}

// CancelRequest identifies the order to cancel.
type CancelRequest struct {
	OrderID     string
	Attribution Attribution
}

// CancelResponse returns upstream cancel status.
type CancelResponse struct {
	Status      string
	Attribution Attribution
}

// QueryRequest identifies the order to fetch.
type QueryRequest struct {
	OrderID     string
	Attribution Attribution
}

// QueryResponse returns current upstream order state.
type QueryResponse struct {
	Order       clobtypes.OrderResponse
	Attribution Attribution
}

// ReplayRequest controls paginated order replay.
type ReplayRequest struct {
	Market      string
	Cursor      string
	Limit       int
	Attribution Attribution
}

// ReplayResponse returns replay pages and cursor.
type ReplayResponse struct {
	Orders      []clobtypes.OrderResponse
	NextCursor  string
	Count       int
	Limit       int
	Attribution Attribution
}

type clobExecutionClient interface {
	CreateOrder(ctx context.Context, order *clobtypes.Order) (clobtypes.OrderResponse, error)
	CancelOrder(ctx context.Context, req *clobtypes.CancelOrderRequest) (clobtypes.CancelResponse, error)
	Order(ctx context.Context, id string) (clobtypes.OrderResponse, error)
	Orders(ctx context.Context, req *clobtypes.OrdersRequest) (clobtypes.OrdersResponse, error)
}

// CLOBEngine adapts a CLOB client to the unified execution Engine contract.
type CLOBEngine struct {
	client clobExecutionClient
}

var _ Engine = (*CLOBEngine)(nil)

// NewCLOBEngine creates an Engine backed by CLOB REST APIs.
func NewCLOBEngine(client clobExecutionClient) (*CLOBEngine, error) {
	if client == nil {
		return nil, errNilClient
	}
	return &CLOBEngine{client: client}, nil
}

// Place submits one order to the exchange.
func (e *CLOBEngine) Place(ctx context.Context, req PlaceRequest) (PlaceResponse, error) {
	if req.Order == nil {
		return PlaceResponse{}, errOrderRequired
	}
	attr := NormalizeAttribution(req.Attribution)
	order, err := e.client.CreateOrder(ctx, req.Order)
	if err != nil {
		return PlaceResponse{}, err
	}
	return PlaceResponse{Order: order, Attribution: attr}, nil
}

// Cancel requests cancellation for an existing order.
func (e *CLOBEngine) Cancel(ctx context.Context, req CancelRequest) (CancelResponse, error) {
	orderID := strings.TrimSpace(req.OrderID)
	if orderID == "" {
		return CancelResponse{}, errOrderIDRequired
	}
	attr := NormalizeAttribution(req.Attribution)
	resp, err := e.client.CancelOrder(ctx, &clobtypes.CancelOrderRequest{OrderID: orderID})
	if err != nil {
		return CancelResponse{}, err
	}
	status := "ok"
	if _, ok := resp.NotCanceled[orderID]; ok {
		status = "failed"
	}
	return CancelResponse{Status: status, Attribution: attr}, nil
}

// Query fetches current order state by order id.
func (e *CLOBEngine) Query(ctx context.Context, req QueryRequest) (QueryResponse, error) {
	orderID := strings.TrimSpace(req.OrderID)
	if orderID == "" {
		return QueryResponse{}, errOrderIDRequired
	}
	attr := NormalizeAttribution(req.Attribution)
	order, err := e.client.Order(ctx, orderID)
	if err != nil {
		return QueryResponse{}, err
	}
	return QueryResponse{Order: order, Attribution: attr}, nil
}

// Replay fetches paginated order history that strategy and audit layers can replay.
func (e *CLOBEngine) Replay(ctx context.Context, req ReplayRequest) (ReplayResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultReplayLimit
	}
	attr := NormalizeAttribution(req.Attribution)

	resp, err := e.client.Orders(ctx, &clobtypes.OrdersRequest{
		Market: strings.TrimSpace(req.Market),
		Cursor: strings.TrimSpace(req.Cursor),
		Limit:  limit,
	})
	if err != nil {
		return ReplayResponse{}, err
	}
	return ReplayResponse{
		Orders:      resp.Data,
		NextCursor:  resp.NextCursor,
		Count:       resp.Count,
		Limit:       limit,
		Attribution: attr,
	}, nil
}
