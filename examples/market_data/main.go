package main

import (
	"context"
	"fmt"
	"log"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/heartbeat"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/rfq"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
)

func main() {
	// 1. Initialize Client
	client := clob.NewClient(transport.NewClient(nil, "https://clob.polymarket.com"))

	// 2. Fetch Rewards Markets
	fmt.Println("Fetching Rewards Markets...")
	markets, err := client.RewardsMarketsCurrent(context.Background(), nil)
	if err != nil {
		log.Printf("Failed to fetch rewards markets: %v (This might require auth or be restricted)", err)
	} else {
		fmt.Printf("Found %d rewards markets\n", len(markets.Data))
		for _, m := range markets.Data {
			fmt.Printf("- %s\n", m.ConditionID)
		}
	}

	// 3. Fetch RFQ Config (Public endpoint usually)
	fmt.Println("\nFetching RFQ Config...")
	rfqClient := rfq.NewClient(transport.NewClient(nil, "https://clob.polymarket.com"))
	rfqConfig, err := rfqClient.RFQConfig(context.Background())
	if err != nil {
		log.Printf("Failed to fetch RFQ config: %v", err)
	} else {
		fmt.Printf("RFQ Config: %+v\n", rfqConfig)
	}

	// 4. Check Heartbeat
	fmt.Println("\nChecking System Heartbeat...")
	heartbeatClient := heartbeat.NewClient(transport.NewClient(nil, "https://clob.polymarket.com"))
	heartbeatResp, err := heartbeatClient.Heartbeat(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to check heartbeat: %v", err)
	}
	fmt.Printf("System Status: %v\n", heartbeatResp)
}
