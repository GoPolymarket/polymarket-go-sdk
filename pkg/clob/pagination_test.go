package clob

import (
	"context"
	"net/url"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

func TestOrdersAllPagination(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			buildKey("/data/orders", url.Values{"limit": {"1"}, "next_cursor": {clobtypes.InitialCursor}}): `{"data":[{"id":"1"}],"next_cursor":"NEXT"}`,
			buildKey("/data/orders", url.Values{"limit": {"1"}, "next_cursor": {"NEXT"}}):        `{"data":[{"id":"2"}],"next_cursor":"LTE="}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	results, err := client.OrdersAll(context.Background(), &clobtypes.OrdersRequest{Limit: 1})
	if err != nil {
		t.Fatalf("OrdersAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(results))
	}
}

func TestTradesAllPagination(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			buildKey("/data/trades", url.Values{"limit": {"1"}, "next_cursor": {clobtypes.InitialCursor}}): `{"data":[{"id":"1"}],"next_cursor":"NEXT"}`,
			buildKey("/data/trades", url.Values{"limit": {"1"}, "next_cursor": {"NEXT"}}):        `{"data":[{"id":"2"}],"next_cursor":"LTE="}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	results, err := client.TradesAll(context.Background(), &clobtypes.TradesRequest{Limit: 1})
	if err != nil {
		t.Fatalf("TradesAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(results))
	}
}

func TestBuilderTradesAllPagination(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			buildKey("/builder/trades", url.Values{"limit": {"1"}, "next_cursor": {clobtypes.InitialCursor}}): `{"data":[{"id":"1"}],"next_cursor":"NEXT"}`,
			buildKey("/builder/trades", url.Values{"limit": {"1"}, "next_cursor": {"NEXT"}}):        `{"data":[{"id":"2"}],"next_cursor":"LTE="}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	results, err := client.BuilderTradesAll(context.Background(), &clobtypes.BuilderTradesRequest{Limit: 1})
	if err != nil {
		t.Fatalf("BuilderTradesAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 builder trades, got %d", len(results))
	}
}

func TestMarketsAllPagination(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			buildKey("/markets", url.Values{"limit": {"1"}, "cursor": {clobtypes.InitialCursor}}): `{"data":[{"id":"1"}],"next_cursor":"NEXT"}`,
			buildKey("/markets", url.Values{"limit": {"1"}, "cursor": {"NEXT"}}):                  `{"data":[{"id":"2"}],"next_cursor":"LTE="}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	results, err := client.MarketsAll(context.Background(), &clobtypes.MarketsRequest{Limit: 1})
	if err != nil {
		t.Fatalf("MarketsAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 markets, got %d", len(results))
	}
}
