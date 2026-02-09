package clob

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

type headerCaptureDoer struct {
	response   string
	lastHeader http.Header
}

func (d *headerCaptureDoer) Do(req *http.Request) (*http.Response, error) {
	d.lastHeader = req.Header.Clone()
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(d.response)),
		Header:     make(http.Header),
	}
	return resp, nil
}

func TestAccountMethods(t *testing.T) {
	signer, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	ctx := context.Background()

	t.Run("BalanceAllowance", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/balance-allowance?asset=USDC&signature_type=0": `{"balance":"100","allowances":{"0xabc":"100"}}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{Asset: "USDC"})
		if err != nil || resp.Balance != "100" {
			t.Errorf("BalanceAllowance failed: %v", err)
		}
	})

	t.Run("BalanceAllowanceParams", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/balance-allowance?asset_type=CONDITIONAL&signature_type=1&token_id=123": `{"balance":"50","allowances":{"0xdef":"50"}}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		sigType := 1
		resp, err := client.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{
			AssetType:     clobtypes.AssetTypeConditional,
			TokenID:       "123",
			SignatureType: &sigType,
		})
		if err != nil || resp.Balance != "50" {
			t.Errorf("BalanceAllowance params failed: %v", err)
		}
	})

	t.Run("BalanceAllowanceDefaultSignature", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/balance-allowance?asset_type=CONDITIONAL&signature_type=1&token_id=999": `{"balance":"75","allowances":{"0xaaa":"75"}}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		client.signatureType = auth.SignatureProxy
		resp, err := client.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{
			AssetType: clobtypes.AssetTypeConditional,
			TokenID:   "999",
		})
		if err != nil || resp.Balance != "75" {
			t.Errorf("BalanceAllowance default signature failed: %v", err)
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
			responses: map[string]string{
				"/rewards/user?date=2025-01-01&signature_type=0": `{"data":[{"date":"2025-01-01","condition_id":"c1","asset_address":"a1","maker_address":"m1","earnings":"10","asset_rate":"1"}],"next_cursor":"LTE=","limit":1,"count":1}`,
			},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.UserEarnings(ctx, &clobtypes.UserEarningsRequest{Date: "2025-01-01"})
		if err != nil || len(resp.Data) != 1 || resp.Data[0].Earnings != "10" {
			t.Errorf("UserEarnings failed: %v", err)
		}
	})

	t.Run("RewardsMarketsCurrent", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{
				"/rewards/markets/current?next_cursor=NEXT": `{"data":[{"condition_id":"c1","rewards_max_spread":"0.1","rewards_min_size":"1"}],"next_cursor":"LTE=","limit":1,"count":1}`,
			},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.RewardsMarketsCurrent(ctx, &clobtypes.RewardsMarketsRequest{NextCursor: "NEXT"})
		if err != nil || len(resp.Data) != 1 {
			t.Errorf("RewardsMarketsCurrent failed: %v", err)
		}
	})

	t.Run("UserRewardsByMarket", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{
				"/rewards/user/by-market?date=2025-01-01&no_competition=false&signature_type=0": `[{"condition_id":"c1","question":"q","market_slug":"m","event_slug":"e","image":"i","rewards_max_spread":"0.1","rewards_min_size":"1","market_competitiveness":"0.5","maker_address":"m1","earning_percentage":"0.2"}]`,
			},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		resp, err := client.UserRewardsByMarket(ctx, &clobtypes.UserRewardsByMarketRequest{
			Date:          "2025-01-01",
			NoCompetition: false,
		})
		if err != nil || len(resp) != 1 || resp[0].ConditionID != "c1" {
			t.Errorf("UserRewardsByMarket failed: %v", err)
		}
	})

	t.Run("UpdateBalanceAllowanceEmptyBody", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/balance-allowance/update?asset=USDC&signature_type=0": `{"balance":"0","allowances":{}}`},
		}
		client := &clientImpl{httpClient: transport.NewClient(doer, "http://example")}
		_, err := client.UpdateBalanceAllowance(ctx, &clobtypes.BalanceAllowanceUpdateRequest{Asset: "USDC"})
		if err != nil {
			t.Errorf("UpdateBalanceAllowance empty body failed: %v", err)
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

	t.Run("CreateAPIKeyDefaultNonce", func(t *testing.T) {
		doer := &headerCaptureDoer{response: `{"apiKey":"k2"}`}
		nonce := int64(42)
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
			signer:     signer,
			authNonce:  &nonce,
		}
		_, err := client.CreateAPIKey(ctx)
		if err != nil {
			t.Errorf("CreateAPIKey default nonce failed: %v", err)
		}
		if got := doer.lastHeader.Get(auth.HeaderPolyNonce); got != "42" {
			t.Errorf("expected nonce header 42, got %q", got)
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
