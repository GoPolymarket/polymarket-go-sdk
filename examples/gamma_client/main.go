package main

import (
	"context"
	"fmt"
	"log"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/gamma"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

func main() {
	// 1. Initialize Gamma Client
	client := gamma.NewClient(transport.NewClient(nil, "https://gamma-api.polymarket.com"))

	// 2. Fetch Markets
	fmt.Println("Fetching Markets...")
	limit := 5
	active := true
	markets, err := client.GetMarkets(context.Background(), &gamma.MarketsRequest{
		Limit:  &limit,
		Active: &active,
		Order:  "volume",
	})
	if err != nil {
		log.Fatalf("Failed to fetch markets: %v", err)
	}

	for _, m := range markets {
		fmt.Printf("- [%s] %s (Vol: %s)\n", m.ID, m.Question, m.Volume.String())
	}

	if len(markets) > 0 {
		// 3. Fetch Specific Market
		fmt.Printf("\nFetching details for market %s...\n", markets[0].ID)
		market, err := client.GetMarket(context.Background(), markets[0].ID)
		if err != nil {
			log.Fatalf("Failed to fetch market details: %v", err)
		}
		fmt.Printf("Title: %s\n", market.Question)
		fmt.Printf("Liquidity: %s\n", market.Liquidity.String())
	}
}
