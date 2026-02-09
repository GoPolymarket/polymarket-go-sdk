package rfq

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

type staticDoer struct {
	responses map[string]string
}

func (d *staticDoer) Do(req *http.Request) (*http.Response, error) {
	payload := d.responses[req.URL.Path]
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(payload)),
		Header:     make(http.Header),
	}, nil
}

func TestRFQMethods(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/rfq/request":      `{"id":"r1"}`,
			"/rfq/data/requests": `[]`,
			"/rfq/quote":        `{"id":"q1"}`,
			"/rfq/data/quotes":   `[]`,
			"/rfq/data/best-quote": `{"id":"q1"}`,
			"/rfq/config":       `{"status":"OK"}`,
		},
	}
	client := NewClient(transport.NewClient(doer, "http://example"))
	ctx := context.Background()

	t.Run("CreateRFQRequest", func(t *testing.T) {
		_, err := client.CreateRFQRequest(ctx, &RFQRequest{})
		if err != nil {
			t.Errorf("CreateRFQRequest failed: %v", err)
		}
	})

	t.Run("RFQRequests", func(t *testing.T) {
		_, err := client.RFQRequests(ctx, nil)
		if err != nil {
			t.Errorf("RFQRequests failed: %v", err)
		}
	})

	t.Run("CreateRFQQuote", func(t *testing.T) {
		_, err := client.CreateRFQQuote(ctx, &RFQQuote{})
		if err != nil {
			t.Errorf("CreateRFQQuote failed: %v", err)
		}
	})

	t.Run("RFQQuotes", func(t *testing.T) {
		_, err := client.RFQQuotes(ctx, nil)
		if err != nil {
			t.Errorf("RFQQuotes failed: %v", err)
		}
	})

	t.Run("RFQConfig", func(t *testing.T) {
		_, err := client.RFQConfig(ctx)
		if err != nil {
			t.Errorf("RFQConfig failed: %v", err)
		}
	})
}

func TestRFQMethods_Cancel(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/rfq/request": `{"status":"OK"}`,
			"/rfq/quote":   `{"status":"OK"}`,
		},
	}
	client := NewClient(transport.NewClient(doer, "http://example"))
	ctx := context.Background()

	t.Run("CancelRFQRequest", func(t *testing.T) {
		resp, err := client.CancelRFQRequest(ctx, &RFQCancelRequest{ID: "r1"})
		if err != nil {
			t.Errorf("CancelRFQRequest failed: %v", err)
		}
		if resp.Status != "OK" {
			t.Errorf("expected OK, got %s", resp.Status)
		}
	})

	t.Run("CancelRFQQuote", func(t *testing.T) {
		resp, err := client.CancelRFQQuote(ctx, &RFQCancelQuote{ID: "q1"})
		if err != nil {
			t.Errorf("CancelRFQQuote failed: %v", err)
		}
		if resp.Status != "OK" {
			t.Errorf("expected OK, got %s", resp.Status)
		}
	})
}

func TestRFQMethods_BestQuote(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/rfq/data/best-quote": `{"id":"q1","quoteId":"q1","requestId":"r1","price":"0.5"}`,
		},
	}
	client := NewClient(transport.NewClient(doer, "http://example"))
	ctx := context.Background()

	t.Run("BestQuote_WithRequestID", func(t *testing.T) {
		resp, err := client.RFQBestQuote(ctx, &RFQBestQuoteQuery{RequestID: "r1"})
		if err != nil {
			t.Errorf("RFQBestQuote failed: %v", err)
		}
		if resp.ID != "q1" {
			t.Errorf("expected q1, got %s", resp.ID)
		}
	})

	t.Run("BestQuote_WithRequestIDs", func(t *testing.T) {
		resp, err := client.RFQBestQuote(ctx, &RFQBestQuoteQuery{RequestIDs: []string{"r1", "r2"}})
		if err != nil {
			t.Errorf("RFQBestQuote failed: %v", err)
		}
		if resp.ID != "q1" {
			t.Errorf("expected q1, got %s", resp.ID)
		}
	})

	t.Run("BestQuote_Nil", func(t *testing.T) {
		_, err := client.RFQBestQuote(ctx, nil)
		if err != nil {
			t.Errorf("RFQBestQuote nil failed: %v", err)
		}
	})
}

func TestRFQMethods_AcceptAndApprove(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/rfq/request/accept": `{"status":"OK","tradeIds":["t1"]}`,
			"/rfq/quote/approve":  `{"status":"OK","tradeIds":["t2"]}`,
		},
	}
	client := NewClient(transport.NewClient(doer, "http://example"))
	ctx := context.Background()

	t.Run("RFQRequestAccept", func(t *testing.T) {
		resp, err := client.RFQRequestAccept(ctx, &RFQAcceptRequest{RequestID: "r1", QuoteID: "q1"})
		if err != nil {
			t.Errorf("RFQRequestAccept failed: %v", err)
		}
		if resp.Status != "OK" {
			t.Errorf("expected OK, got %s", resp.Status)
		}
		if len(resp.TradeIDs) != 1 || resp.TradeIDs[0] != "t1" {
			t.Errorf("unexpected tradeIDs: %v", resp.TradeIDs)
		}
	})

	t.Run("RFQQuoteApprove", func(t *testing.T) {
		resp, err := client.RFQQuoteApprove(ctx, &RFQApproveQuote{RequestID: "r1", QuoteID: "q1"})
		if err != nil {
			t.Errorf("RFQQuoteApprove failed: %v", err)
		}
		if resp.Status != "OK" {
			t.Errorf("expected OK, got %s", resp.Status)
		}
		if len(resp.TradeIDs) != 1 || resp.TradeIDs[0] != "t2" {
			t.Errorf("unexpected tradeIDs: %v", resp.TradeIDs)
		}
	})
}

func TestRFQRequests_WithQuery(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/rfq/data/requests": `[]`,
		},
	}
	client := NewClient(transport.NewClient(doer, "http://example"))
	ctx := context.Background()

	_, err := client.RFQRequests(ctx, &RFQRequestsQuery{
		Limit:      10,
		Cursor:     "abc",
		State:      RFQStateActive,
		RequestIDs: []string{"r1", "r2"},
		Markets:    []string{"m1"},
		SizeMin:    "1",
		SizeMax:    "100",
		PriceMin:   "0.1",
		PriceMax:   "0.9",
		SortBy:     RFQSortByPrice,
		SortDir:    RFQSortDirAsc,
	})
	if err != nil {
		t.Errorf("RFQRequests with query failed: %v", err)
	}
}

func TestRFQQuotes_WithQuery(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"/rfq/data/quotes": `[]`,
		},
	}
	client := NewClient(transport.NewClient(doer, "http://example"))
	ctx := context.Background()

	_, err := client.RFQQuotes(ctx, &RFQQuotesQuery{
		Limit:       5,
		Offset:      "10",
		State:       RFQStateInactive,
		RequestIDs:  []string{"r1"},
		QuoteIDs:    []string{"q1"},
		Markets:     []string{"m1"},
		SizeUsdcMin: "100",
		SizeUsdcMax: "1000",
		SortBy:      RFQSortByExpiry,
		SortDir:     RFQSortDirDesc,
	})
	if err != nil {
		t.Errorf("RFQQuotes with query failed: %v", err)
	}
}
