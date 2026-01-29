package clob

import (
	"context"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

func TestAccountMethods(t *testing.T) {
	signer, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	ctx := context.Background()

	t.Run("BalanceAllowance", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/balance-allowance?asset=USDC": `{"balance":"100","allowance":"100"}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{Asset: "USDC"})
		if err != nil || resp.Balance != "100" {
			t.Errorf("BalanceAllowance failed: %v", err)
		}
	})

	t.Run("Notifications", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/notifications": `[{"id":"n1"}]`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.Notifications(ctx, nil)
		if err != nil || len(resp) == 0 {
			t.Errorf("Notifications failed: %v", err)
		}
	})

	t.Run("UserEarnings", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/rewards/user": `{"earnings":"10"}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.UserEarnings(ctx, nil)
		if err != nil || resp.Earnings != "10" {
			t.Errorf("UserEarnings failed: %v", err)
		}
	})

	t.Run("ListAPIKeys", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/auth/api-keys": `{"apiKeys":[{"apiKey":"k1"}]}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.ListAPIKeys(ctx)
		if err != nil || len(resp.APIKeys) == 0 {
			t.Errorf("ListAPIKeys failed: %v", err)
		}
	})

	t.Run("CreateAPIKey", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/auth/api-key": `{"apiKey":"k2"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
			signer:     signer,
		}
		resp, err := client.CreateAPIKey(ctx)
		if err != nil || resp.APIKey != "k2" {
			t.Errorf("CreateAPIKey failed: %v", err)
		}
	})

	t.Run("DeriveAPIKey", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/auth/derive-api-key": `{"apiKey":"k3"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
			signer:     signer,
		}
		resp, err := client.DeriveAPIKey(ctx)
		if err != nil || resp.APIKey != "k3" {
			t.Errorf("DeriveAPIKey failed: %v", err)
		}
	})

	t.Run("DeleteAPIKey", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/auth/api-key?api_key=k1": `{"apiKey":"k1"}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		_, err := client.DeleteAPIKey(ctx, "k1")
		if err != nil {
			t.Errorf("DeleteAPIKey failed: %v", err)
		}
	})

	t.Run("ClosedOnlyStatus", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/auth/ban-status/closed-only": `{"closed_only":false}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.ClosedOnlyStatus(ctx)
		if err != nil || resp.ClosedOnly != false {
			t.Errorf("ClosedOnlyStatus failed: %v", err)
		}
	})
}

