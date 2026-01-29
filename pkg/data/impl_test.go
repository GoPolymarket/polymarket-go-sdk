package data

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

type staticDoer struct {
	responses map[string]string
}

func (d *staticDoer) Do(req *http.Request) (*http.Response, error) {
	key := req.URL.Path
	if req.URL.RawQuery != "" {
		key += "?" + req.URL.RawQuery
	}
	payload, ok := d.responses[key]
	if !ok {
		return nil, fmt.Errorf("unexpected request %q", key)
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(payload)),
		Header:     make(http.Header),
	}
	return resp, nil
}

func TestDataMethods(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			"":                  `{"status":"UP"}`,
			"/positions":        `[]`,
			"/trades":           `[]`,
			"/activity":         `[]`,
			"/holders":          `[]`,
			"/value":            `{"value":"100"}`,
			"/closed-positions": `[]`,
			"/traded":           `{"traded":true}`,
			"/oi":               `{"oi":"1000"}`,
			"/live-volume":      `[]`,
			"/v1/leaderboard":   `[]`,
			"/v1/builders/leaderboard": `[]`,
			"/v1/builders/volume":      `[]`,
		},
	}
	client := NewClient(transport.NewClient(doer, BaseURL))
	ctx := context.Background()

	t.Run("Health", func(t *testing.T) {
		_, _ = client.Health(ctx)
	})

	t.Run("Positions", func(t *testing.T) {
		_, _ = client.Positions(ctx, nil)
	})

	t.Run("Value", func(t *testing.T) {
		_, _ = client.Value(ctx, nil)
	})

	t.Run("Trades", func(t *testing.T) {
		_, _ = client.Trades(ctx, nil)
	})

	t.Run("Activity", func(t *testing.T) {
		_, _ = client.Activity(ctx, nil)
	})

	t.Run("Holders", func(t *testing.T) {
		_, _ = client.Holders(ctx, nil)
	})

	t.Run("ClosedPositions", func(t *testing.T) {
		_, _ = client.ClosedPositions(ctx, nil)
	})

	t.Run("Traded", func(t *testing.T) {
		_, _ = client.Traded(ctx, nil)
	})

	t.Run("OpenInterest", func(t *testing.T) {
		_, _ = client.OpenInterest(ctx, nil)
	})

	t.Run("LiveVolume", func(t *testing.T) {
		_, _ = client.LiveVolume(ctx, nil)
	})

	t.Run("Builders", func(t *testing.T) {
		_, _ = client.BuildersLeaderboard(ctx, nil)
		_, _ = client.BuildersVolume(ctx, nil)
	})

	t.Run("Leaderboard", func(t *testing.T) {
		_, _ = client.Leaderboard(ctx, nil)
	})
}
