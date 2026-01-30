package clob

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

func TestOrderManagementMethods(t *testing.T) {
	signer, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	apiKey := &auth.APIKey{Key: "k1", Secret: "s1", Passphrase: "p1"}
	ctx := context.Background()

	t.Run("PostOrder", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/order": `{"id":"o1","status":"OK"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
			signer:     signer,
			apiKey:     apiKey,
		}
		order := &clobtypes.SignedOrder{
			Order:     clobtypes.Order{Side: "BUY"},
			Signature: "0x123",
			Owner:     "0xabc",
		}
		resp, err := client.PostOrder(ctx, order)
		if err != nil || resp.ID != "o1" {
			t.Errorf("PostOrder failed: %v", err)
		}
	})

	t.Run("CancelAll", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/cancel-all": `{"status":"OK","count":10}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelAll(ctx)
		if err != nil || resp.Status != "OK" {
			t.Errorf("CancelAll failed: %v", err)
		}
	})

	t.Run("CancelOrder", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/order": `{"status":"OK"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelOrder(ctx, &clobtypes.CancelOrderRequest{OrderID: "o1"})
		if err != nil || resp.Status != "OK" {
			t.Errorf("CancelOrder failed: %v", err)
		}
	})

	t.Run("CancelOrders", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/orders": `{"status":"OK"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelOrders(ctx, &clobtypes.CancelOrdersRequest{OrderIDs: []string{"o1"}})
		if err != nil || resp.Status != "OK" {
			t.Errorf("CancelOrders failed: %v", err)
		}
	})

	t.Run("CancelMarketOrders", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/cancel-market-orders": `{"status":"OK"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelMarketOrders(ctx, &clobtypes.CancelMarketOrdersRequest{Market: "m1"})
		if err != nil || resp.Status != "OK" {
			t.Errorf("CancelMarketOrders failed: %v", err)
		}
	})

	t.Run("BuilderTrades", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/builder/trades": `{"data":[]}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.BuilderTrades(ctx, nil)
		if err != nil {
			t.Errorf("BuilderTrades failed: %v", err)
		}
		if resp.Data == nil {
			t.Errorf("expected empty slice")
		}
	})

	t.Run("OrderLookup", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/data/order/o1": `{"id":"o1","status":"OK"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.Order(ctx, "o1")
		if err != nil || resp.ID != "o1" {
			t.Errorf("Order lookup failed: %v", err)
		}
	})

	t.Run("OrdersList", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/data/orders": `{"data":[{"id":"o1"}],"next_cursor":"LTE="}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.Orders(ctx, nil)
		if err != nil || len(resp.Data) == 0 {
			t.Errorf("Orders list failed: %v", err)
		}
	})

	t.Run("OrderScoring", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/order-scoring?order_id=o1": `{"scoring":true}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.OrderScoring(ctx, &clobtypes.OrderScoringRequest{ID: "o1"})
		if err != nil || !resp.Scoring {
			t.Errorf("OrderScoring failed: %v", err)
		}
	})

	t.Run("OrdersScoring", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/orders-scoring": `{"o1":true,"o2":false}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.OrdersScoring(ctx, &clobtypes.OrdersScoringRequest{IDs: []string{"o1", "o2"}})
		if err != nil || resp["o1"] != true || resp["o2"] != false {
			t.Errorf("OrdersScoring failed: %v", err)
		}
	})
}

func TestSignOrderDefaults(t *testing.T) {
	signer, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	apiKey := &auth.APIKey{Key: "k1", Secret: "s1", Passphrase: "p1"}

	funder := common.HexToAddress("0x3333333333333333333333333333333333333333")
	client := &clientImpl{
		signer:        signer,
		apiKey:        apiKey,
		signatureType: auth.SignatureProxy,
		funder:        &funder,
		saltGenerator: func() (*big.Int, error) { return big.NewInt(7), nil },
	}

	order := &clobtypes.Order{
		Side:        "BUY",
		TokenID:     types.U256{Int: big.NewInt(1)},
		MakerAmount: decimal.NewFromInt(10),
		TakerAmount: decimal.NewFromInt(5),
		FeeRateBps:  decimal.NewFromInt(0),
		Nonce:       types.U256{Int: big.NewInt(1)},
		Expiration:  types.U256{Int: big.NewInt(0)},
		Taker:       common.Address{},
		Signer:      signer.Address(),
	}

	signed, err := client.signOrder(order)
	if err != nil {
		t.Fatalf("signOrder failed: %v", err)
	}
	if signed.Order.SignatureType == nil || *signed.Order.SignatureType != 1 {
		t.Fatalf("signature type mismatch: %+v", signed.Order.SignatureType)
	}
	if signed.Order.Maker != funder {
		t.Fatalf("maker mismatch: got %s want %s", signed.Order.Maker.Hex(), funder.Hex())
	}
	if signed.Order.Salt.Int == nil || signed.Order.Salt.Int.Int64() != 7 {
		t.Fatalf("salt mismatch: got %v", signed.Order.Salt.Int)
	}
}
