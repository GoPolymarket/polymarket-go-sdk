package heartbeat

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

func TestHeartbeat(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{"/v1/heartbeats": `{"status":"OK"}`},
	}
	client := NewClient(transport.NewClient(doer, "http://example"))
	_, err := client.Heartbeat(context.Background(), nil)
	if err != nil {
		t.Errorf("Heartbeat failed: %v", err)
	}
}
