package main

import (
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
	"context"
	"fmt"
	"log"
	"os"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	
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

	ctx := context.Background()

	trades, err := authClient.TradesAll(ctx, &clobtypes.TradesRequest{
		Limit: 50,
	})
	if err != nil {
		log.Fatalf("TradesAll failed: %v", err)
	}
	fmt.Printf("Fetched %d trades\n", len(trades))
}
