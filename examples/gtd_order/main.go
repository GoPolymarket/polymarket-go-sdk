package main

import (
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob"
	
)

func main() {
	pkHex := os.Getenv("POLYMARKET_PK")
	if pkHex == "" {
		log.Fatalf("POLYMARKET_PK is required")
	}
	apiKey := &auth.APIKey{
		Key:        os.Getenv("POLYMARKET_API_KEY"),
		Secret:     os.Getenv("POLYMARKET_API_SECRET"),
		Passphrase: os.Getenv("POLYMARKET_API_PASSPHRASE"),
	}

	signer, err := auth.NewPrivateKeySigner(pkHex, 137)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	client := polymarket.NewClient(polymarket.WithUseServerTime(true))
	authClient := client.CLOB.WithAuth(signer, apiKey)

	expiration := time.Now().Add(30 * time.Minute).Unix()
	signable, err := clob.NewOrderBuilder(authClient, signer).
		TokenID("1234567890").
		Side("SELL").
		Price(0.42).
		Size(10).
		OrderType(clobtypes.OrderTypeGTD).
		ExpirationUnix(expiration).
		PostOnly(false).
		BuildSignable()
	if err != nil {
		log.Fatalf("BuildSignable failed: %v", err)
	}

	resp, err := authClient.CreateOrderFromSignable(context.Background(), signable)
	if err != nil {
		log.Printf("Order creation returned error (expected in demo): %v", err)
		return
	}
	fmt.Printf("Order Created! ID: %s\n", resp.ID)
}
