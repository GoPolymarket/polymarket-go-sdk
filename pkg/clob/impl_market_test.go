package clob

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

type assertBodyDoer struct {
	t         *testing.T
	expected  map[string]string
	responses map[string]string
}

func (d *assertBodyDoer) Do(req *http.Request) (*http.Response, error) {
	key := req.URL.Path
	if req.URL.RawQuery != "" {
		key += "?" + req.URL.RawQuery
	}
	if expected, ok := d.expected[key]; ok {
		body, _ := io.ReadAll(req.Body)
		if expected == "" {
			if len(bytes.TrimSpace(body)) > 0 {
				d.t.Fatalf("expected empty body for %s, got %s", key, string(body))
			}
		} else {
			var gotPayload interface{}
			var wantPayload interface{}
			if err := json.Unmarshal(body, &gotPayload); err != nil {
				d.t.Fatalf("failed to decode request body for %s: %v", key, err)
			}
			if err := json.Unmarshal([]byte(expected), &wantPayload); err != nil {
				d.t.Fatalf("failed to decode expected body for %s: %v", key, err)
			}
			if !reflect.DeepEqual(gotPayload, wantPayload) {
				d.t.Fatalf("body mismatch for %s: got %v want %v", key, gotPayload, wantPayload)
			}
		}
	}

	payload := "{}"
	if d.responses != nil {
		if resp, ok := d.responses[key]; ok {
			payload = resp
		}
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(payload)),
		Header:     make(http.Header),
	}
	return resp, nil
}

func TestMarketMethods(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/markets":                     `{"data":[{"id":"m1"}],"next_cursor":"LTE="}`,
			"/markets/m1":                  `{"id":"m1","question":"test?"}`,
			"/simplified-markets":          `{"data":[{"id":"s1"}]}`,
			"/sampling-markets":            `{"data":[{"id":"sam1"}]}`,
			"/sampling-simplified-markets": `{"data":[{"id":"ss1"}]}`,
			"/book?token_id=t1":            `{"market_id":"m1","bids":[],"asks":[]}`,
			"/midpoint?token_id=t1":        `{"midpoint":"0.5"}`,
			"/price?token_id=t1":           `{"price":"0.51"}`,
			"/spread?token_id=t1":          `{"spread":"0.01"}`,
			"/tick-size?token_id=t1":       `{"minimum_tick_size":0.01}`,
			"/neg-risk?token_id=t1":        `{"neg_risk":true}`,
			"/fee-rate?token_id=t1":        `{"base_fee":10}`,
			"/prices-history?token_id=t1":  `{"history":[{"t":123,"p":0.5}]}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	ctx := context.Background()

	t.Run("Markets", func(t *testing.T) {
		resp, err := client.Markets(ctx, nil)
		if err != nil || len(resp.Data) == 0 {
			t.Errorf("Markets failed: %v", err)
		}
	})

	t.Run("Market", func(t *testing.T) {
		resp, err := client.Market(ctx, "m1")
		if err != nil || resp.ID != "m1" {
			t.Errorf("Market failed: %v", err)
		}
	})

	t.Run("SimplifiedMarkets", func(t *testing.T) {
		resp, err := client.SimplifiedMarkets(ctx, nil)
		if err != nil || len(resp.Data) == 0 {
			t.Errorf("SimplifiedMarkets failed: %v", err)
		}
	})

	t.Run("SamplingMarkets", func(t *testing.T) {
		resp, err := client.SamplingMarkets(ctx, nil)
		if err != nil || len(resp.Data) == 0 {
			t.Errorf("SamplingMarkets failed: %v", err)
		}
	})

	t.Run("SamplingSimplifiedMarkets", func(t *testing.T) {
		resp, err := client.SamplingSimplifiedMarkets(ctx, nil)
		if err != nil || len(resp.Data) == 0 {
			t.Errorf("SamplingSimplifiedMarkets failed: %v", err)
		}
	})

	t.Run("OrderBook", func(t *testing.T) {
		resp, err := client.OrderBook(ctx, &clobtypes.BookRequest{TokenID: "t1"})
		if err != nil || resp.MarketID != "m1" {
			t.Errorf("OrderBook failed: %v", err)
		}
	})

	t.Run("Midpoint", func(t *testing.T) {
		resp, err := client.Midpoint(ctx, &clobtypes.MidpointRequest{TokenID: "t1"})
		if err != nil || resp.Midpoint != "0.5" {
			t.Errorf("Midpoint failed: %v", err)
		}
	})

	t.Run("Price", func(t *testing.T) {
		resp, err := client.Price(ctx, &clobtypes.PriceRequest{TokenID: "t1"})
		if err != nil || resp.Price != "0.51" {
			t.Errorf("Price failed: %v", err)
		}
	})

	t.Run("TickSize", func(t *testing.T) {
		resp, err := client.TickSize(ctx, &clobtypes.TickSizeRequest{TokenID: "t1"})
		if err != nil || resp.MinimumTickSize != 0.01 {
			t.Errorf("TickSize failed: %v", err)
		}
		// Test cache
		client.SetTickSize("t1", 0.02)
		resp, _ = client.TickSize(ctx, &clobtypes.TickSizeRequest{TokenID: "t1"})
		if resp.MinimumTickSize != 0.02 {
			t.Errorf("cache failed")
		}
	})

	t.Run("PricesHistory", func(t *testing.T) {
		resp, err := client.PricesHistory(ctx, &clobtypes.PricesHistoryRequest{TokenID: "t1"})
		if err != nil || len(resp) == 0 {
			t.Errorf("PricesHistory failed: %v", err)
		}
	})
}

func TestBatchMethods(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/books":              `[{"market_id":"m1"}]`,
			"/midpoints":          `[{"midpoint":"0.5"}]`,
			"/prices":             `[{"price":"0.5"}]`,
			"/spreads":            `[{"spread":"0.01"}]`,
			"/last-trades-prices": `[{"price":"0.5"}]`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
	}
	ctx := context.Background()

	t.Run("OrderBooks", func(t *testing.T) {
		resp, err := client.OrderBooks(ctx, &clobtypes.BooksRequest{TokenIDs: []string{"t1"}})
		if err != nil || len(resp) == 0 {
			t.Errorf("OrderBooks failed: %v", err)
		}
	})

	t.Run("Prices", func(t *testing.T) {
		resp, err := client.Prices(ctx, &clobtypes.PricesRequest{TokenIDs: []string{"t1"}})
		if err != nil || len(resp) == 0 {
			t.Errorf("Prices failed: %v", err)
		}
	})

	t.Run("Midpoints", func(t *testing.T) {
		resp, err := client.Midpoints(ctx, &clobtypes.MidpointsRequest{TokenIDs: []string{"t1"}})
		if err != nil || len(resp) == 0 {
			t.Errorf("Midpoints failed: %v", err)
		}
	})

	t.Run("Spreads", func(t *testing.T) {
		resp, err := client.Spreads(ctx, &clobtypes.SpreadsRequest{TokenIDs: []string{"t1"}})
		if err != nil || len(resp) == 0 {
			t.Errorf("Spreads failed: %v", err)
		}
	})

	t.Run("OrderBooksRequestsBody", func(t *testing.T) {
		doer := &assertBodyDoer{
			t: t,
			expected: map[string]string{
				"/books": `[{"token_id":"t1","side":"BUY"},{"token_id":"t2"}]`,
			},
			responses: map[string]string{
				"/books": `[{"market_id":"m1"}]`,
			},
		}
		bodyClient := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		_, err := bodyClient.OrderBooks(ctx, &clobtypes.BooksRequest{
			Requests: []clobtypes.BookRequest{
				{TokenID: "t1", Side: "BUY"},
				{TokenID: "t2"},
			},
		})
		if err != nil {
			t.Errorf("OrderBooks requests body failed: %v", err)
		}
	})

	t.Run("PricesRequestsBody", func(t *testing.T) {
		doer := &assertBodyDoer{
			t: t,
			expected: map[string]string{
				"/prices": `[{"token_id":"t1","side":"BUY"},{"token_id":"t2","side":"SELL"}]`,
			},
			responses: map[string]string{
				"/prices": `[{"price":"0.5"}]`,
			},
		}
		bodyClient := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		_, err := bodyClient.Prices(ctx, &clobtypes.PricesRequest{
			Requests: []clobtypes.PriceRequest{
				{TokenID: "t1", Side: "BUY"},
				{TokenID: "t2", Side: "SELL"},
			},
		})
		if err != nil {
			t.Errorf("Prices requests body failed: %v", err)
		}
	})

	t.Run("SpreadsRequestsBody", func(t *testing.T) {
		doer := &assertBodyDoer{
			t: t,
			expected: map[string]string{
				"/spreads": `[{"token_id":"t1","side":"BUY"},{"token_id":"t2"}]`,
			},
			responses: map[string]string{
				"/spreads": `[{"spread":"0.01"}]`,
			},
		}
		bodyClient := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		_, err := bodyClient.Spreads(ctx, &clobtypes.SpreadsRequest{
			Requests: []clobtypes.SpreadRequest{
				{TokenID: "t1", Side: "BUY"},
				{TokenID: "t2"},
			},
		})
		if err != nil {
			t.Errorf("Spreads requests body failed: %v", err)
		}
	})

	t.Run("LastTradesPrices", func(t *testing.T) {
		resp, err := client.LastTradesPrices(ctx, &clobtypes.LastTradesPricesRequest{TokenIDs: []string{"t1"}})
		if err != nil || len(resp) == 0 {
			t.Errorf("LastTradesPrices failed: %v", err)
		}
	})
}
