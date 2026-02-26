package clob

import (
	"context"
	"net/http"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

func TestClientInitializationAndOptions(t *testing.T) {
	httpClient := transport.NewClient(http.DefaultClient, "http://example")
	client := NewClient(httpClient)
	ctx := context.Background()

	t.Run("Health", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/": `"OK"`},
		}
		client := NewClient(transport.NewClient(doer, "http://example"))
		status, err := client.Health(ctx)
		if err != nil || status != "OK" {
			t.Errorf("Health failed: %v", err)
		}
	})

	t.Run("Time", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/time": `123456789`},
		}
		client := NewClient(transport.NewClient(doer, "http://example"))
		resp, err := client.Time(ctx)
		if err != nil || resp.Timestamp != 123456789 {
			t.Errorf("Time failed: %v", err)
		}
	})

	t.Run("Geoblock", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/api/geoblock": `{"blocked":false}`},
		}
		client := NewClient(transport.NewClient(doer, "http://example"))
		resp, err := client.Geoblock(ctx)
		if err != nil || resp.Blocked != false {
			t.Errorf("Geoblock failed: %v", err)
		}
	})

	t.Run("WithAuth", func(t *testing.T) {
		signer, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
		apiKey := &auth.APIKey{Key: "k"}
		orig := client.(*clientImpl)
		newClient := client.WithAuth(signer, apiKey)
		next := newClient.(*clientImpl)
		if newClient == nil {
			t.Errorf("WithAuth failed")
		}
		if orig == next {
			t.Errorf("WithAuth should return a new client instance")
		}
		if orig.httpClient == next.httpClient {
			t.Errorf("WithAuth should not reuse mutable transport client")
		}
		if orig.signer != nil || orig.apiKey != nil {
			t.Errorf("WithAuth should not mutate original client auth state")
		}
	})

	t.Run("WithBuilderConfig", func(t *testing.T) {
		orig := client.(*clientImpl)
		newClient := client.WithBuilderConfig(&auth.BuilderConfig{})
		next := newClient.(*clientImpl)
		if newClient == nil {
			t.Errorf("WithBuilderConfig failed")
		}
		if orig.httpClient == next.httpClient {
			t.Errorf("WithBuilderConfig should not reuse mutable transport client")
		}
	})

	t.Run("WithUseServerTime", func(t *testing.T) {
		orig := client.(*clientImpl)
		newClient := client.WithUseServerTime(true)
		next := newClient.(*clientImpl)
		if newClient == nil {
			t.Errorf("WithUseServerTime failed")
		}
		if orig == next {
			t.Errorf("WithUseServerTime should return a new client instance")
		}
		if orig.httpClient == next.httpClient {
			t.Errorf("WithUseServerTime should not reuse mutable transport client")
		}
	})

	t.Run("WithGeoblockHost", func(t *testing.T) {
		newClient := client.WithGeoblockHost("http://geo")
		if newClient == nil {
			t.Errorf("WithGeoblockHost failed")
		}
	})

	t.Run("WithWS", func(t *testing.T) {
		newClient := client.WithWS(nil)
		if newClient == nil {
			t.Errorf("WithWS failed")
		}
	})

	t.Run("SubClients", func(t *testing.T) {
		if client.RFQ() == nil {
			t.Errorf("RFQ() nil")
		}
		if client.Heartbeat() == nil {
			t.Errorf("Heartbeat() nil")
		}
		// WS might be nil if not set
	})

	t.Run("Caches", func(t *testing.T) {
		client.SetNegRisk("t1", true)
		client.SetFeeRateBps("t1", 10)
		client.InvalidateCaches()
	})
}
