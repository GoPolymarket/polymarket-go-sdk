package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
	clobtypes "github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func main() {
	fmt.Println("=== Polymarket Trading Bot Example ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pk := os.Getenv("POLYMARKET_PK")
	if pk == "" {
		fmt.Println("SKIPPED: POLYMARKET_PK not set")
		return
	}

	signer, err := auth.NewPrivateKeySigner(pk, 137)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}
	fmt.Printf("Signer: %s\n", signer.Address().Hex())

	client := polymarket.NewClient()
	clobClient := client.CLOB.WithAuth(signer, nil)

	runBot(ctx, clobClient)
}

func runBot(ctx context.Context, client clob.Client) {
	fmt.Println("\n--- Bot Scanning Markets ---")

	markets, err := client.Markets(ctx, &clobtypes.MarketsRequest{
		Active: ptrBool(true),
		Limit:  10,
	})
	if err != nil {
		log.Printf("Failed to get markets: %v", err)
		return
	}

	fmt.Printf("Found %d active markets\n", len(markets.Data))

	for i, market := range markets.Data {
		if len(market.Tokens) == 0 {
			continue
		}

		tokenID := market.Tokens[0].TokenID
		book, err := client.OrderBook(ctx, &clobtypes.BookRequest{
			TokenID: tokenID,
		})
		if err != nil {
			continue
		}

		if len(book.Asks) == 0 || len(book.Bids) == 0 {
			continue
		}

		bestAsk := book.Asks[0].Price
		bestBid := book.Bids[0].Price
		fmt.Printf("%d. %s\n   Ask: %s, Bid: %s\n", i+1, market.Question, bestAsk, bestBid)
	}

	fmt.Println("\n--- Bot Checking Account ---")

	bal, err := client.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{
		AssetType: "conditional",
	})
	if err != nil {
		log.Printf("Failed to get balance: %v", err)
		return
	}
	fmt.Printf("Balance: %s\n", bal.Balance)

	orders, err := client.Orders(ctx, &clobtypes.OrdersRequest{Limit: 5})
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
		return
	}
	fmt.Printf("Open orders: %d\n", len(orders.Data))

	fmt.Println("\n--- Bot Complete ---")
}

func ptrBool(b bool) *bool {
	return &b
}
