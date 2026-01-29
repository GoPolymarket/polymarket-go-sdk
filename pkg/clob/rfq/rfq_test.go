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
