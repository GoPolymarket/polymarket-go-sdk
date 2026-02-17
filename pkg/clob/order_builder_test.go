package clob

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func mustSigner(t *testing.T) auth.Signer {
	t.Helper()
	signer, err := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	return signer
}

func TestBuildMarketPriceValidation(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(10).
		OrderType(clobtypes.OrderTypeFAK).
		Price(0.123).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "decimal places") {
		t.Fatalf("expected decimal place validation error, got %v", err)
	}
}

func TestBuildMarketAmountSharesValidation(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("SELL").
		AmountShares(1.234).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "amount has too many decimal places") {
		t.Fatalf("expected amount decimal validation error, got %v", err)
	}
}

func TestBuildMarketAmountUSDCValidation(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(0.0000001).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "amount has too many decimal places") {
		t.Fatalf("expected amount decimal validation error, got %v", err)
	}
}

func TestBuildMarketUsesOrderBookDepth(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0
	stub.book = clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.6", Size: "100"},
			{Price: "0.55", Size: "100"},
			{Price: "0.5", Size: "100"},
		},
	}

	signable, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(50).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err != nil {
		t.Fatalf("BuildMarket failed: %v", err)
	}

	expectedMaker := decimal.NewFromInt(50_000_000)
	expectedTaker := decimal.NewFromInt(100_000_000)

	if !signable.Order.MakerAmount.Equal(expectedMaker) {
		t.Fatalf("maker amount mismatch: got %s want %s", signable.Order.MakerAmount.String(), expectedMaker.String())
	}
	if !signable.Order.TakerAmount.Equal(expectedTaker) {
		t.Fatalf("taker amount mismatch: got %s want %s", signable.Order.TakerAmount.String(), expectedTaker.String())
	}
}

func TestBuildMarketFOKInsufficientLiquidity(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0
	stub.book = clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.6", Size: "1"},
		},
	}

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(100).
		OrderType(clobtypes.OrderTypeFOK).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "insufficient liquidity") {
		t.Fatalf("expected insufficient liquidity error, got %v", err)
	}
}

func TestBuildMarketFAKUsesTopPriceWhenInsufficient(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0
	stub.book = clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.6", Size: "1"},
			{Price: "0.55", Size: "1"},
		},
	}

	signable, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(100).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err != nil {
		t.Fatalf("BuildMarket failed: %v", err)
	}

	price := decimal.RequireFromString("0.6")
	tickScale := decimalPlaces(decimal.RequireFromString("0.01"))
	rawAmount := decimal.NewFromInt(100)
	takerAmount := rawAmount.Div(price).Truncate(tickScale + lotSizeScale)
	expectedTaker := toFixedDecimal(takerAmount)

	if !signable.Order.MakerAmount.Equal(decimal.NewFromInt(100_000_000)) {
		t.Fatalf("maker amount mismatch: got %s", signable.Order.MakerAmount.String())
	}
	if !signable.Order.TakerAmount.Equal(expectedTaker) {
		t.Fatalf("taker amount mismatch: got %s want %s", signable.Order.TakerAmount.String(), expectedTaker.String())
	}
}

func TestBuildLimitOrder(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 10

	ctx := context.Background()
	signer := mustSigner(t)

	t.Run("BasicLimit", func(t *testing.T) {
		order, err := NewOrderBuilder(stub, signer).
			TokenID("123").
			Side("BUY").
			Price(0.5).
			Size(100).
			BuildWithContext(ctx)
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if order.Side != "BUY" {
			t.Errorf("wrong side")
		}
	})

	t.Run("SignableLimit", func(t *testing.T) {
		postOnly := true
		signable, err := NewOrderBuilder(stub, signer).
			TokenID("123").
			Side("SELL").
			Price(0.6).
			Size(50).
			OrderType(clobtypes.OrderTypeGTD).
			ExpirationUnix(time.Now().Unix() + 3600).
			PostOnly(postOnly).
			BuildSignableWithContext(ctx)
		if err != nil {
			t.Fatalf("BuildSignable failed: %v", err)
		}
		if signable.OrderType != clobtypes.OrderTypeGTD {
			t.Errorf("wrong order type")
		}
		if signable.PostOnly == nil || !*signable.PostOnly {
			t.Errorf("postOnly mismatch")
		}
	})

	t.Run("WalletDerivation", func(t *testing.T) {
		builder := NewOrderBuilder(stub, signer).
			TokenID("123").
			Side("BUY").
			Price(0.5).
			Size(10).
			AmountUSDC(5)

		// Test Proxy
		signable, err := builder.UseProxy().BuildMarketWithContext(ctx)
		if err != nil {
			t.Fatalf("Proxy derivation failed: %v", err)
		}
		if signable.Order.SignatureType == nil || *signable.Order.SignatureType != 1 {
			t.Errorf("proxy type mismatch")
		}

		// Test Safe
		signable, err = builder.UseSafe().BuildMarketWithContext(ctx)
		if err != nil {
			t.Fatalf("Safe derivation failed: %v", err)
		}
		if signable.Order.SignatureType == nil || *signable.Order.SignatureType != 2 {
			t.Errorf("safe type mismatch")
		}
	})
}

func TestOrderBuilderDefaultsFromClient(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0

	signer := mustSigner(t)
	funder := common.HexToAddress("0x1111111111111111111111111111111111111111")
	stub.clientImpl.signatureType = auth.SignatureProxy
	stub.clientImpl.funder = &funder
	stub.clientImpl.saltGenerator = func() (*big.Int, error) {
		return big.NewInt(42), nil
	}

	signable, err := NewOrderBuilder(stub, signer).
		TokenID("123").
		Side("BUY").
		Price(0.5).
		Size(10).
		BuildSignableWithContext(context.Background())
	if err != nil {
		t.Fatalf("BuildSignable failed: %v", err)
	}
	if signable.Order.SignatureType == nil || *signable.Order.SignatureType != 1 {
		t.Fatalf("signature type mismatch: %+v", signable.Order.SignatureType)
	}
	if signable.Order.Maker != funder {
		t.Fatalf("maker mismatch: got %s want %s", signable.Order.Maker.Hex(), funder.Hex())
	}
	if signable.Order.Salt.Int == nil || signable.Order.Salt.Int.Int64() != 42 {
		t.Fatalf("salt mismatch: got %v", signable.Order.Salt.Int)
	}
}

func TestOrderBuilderFunderRequiresSignature(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0

	signer := mustSigner(t)
	funder := common.HexToAddress("0x2222222222222222222222222222222222222222")
	stub.clientImpl.signatureType = auth.SignatureEOA
	stub.clientImpl.funder = &funder

	_, err := NewOrderBuilder(stub, signer).
		TokenID("123").
		Side("BUY").
		Price(0.5).
		Size(1).
		BuildSignableWithContext(context.Background())
	if err == nil || !strings.Contains(err.Error(), "funder requires non-EOA") {
		t.Fatalf("expected funder signature error, got %v", err)
	}
}
