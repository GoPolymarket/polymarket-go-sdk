package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/gamma"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

func main() {
	// 1. Fetch a valid Token ID using Gamma API (Sorted by Volume)
	fmt.Println("Fetching top active market by volume...")
	gammaClient := gamma.NewClient(transport.NewClient(http.DefaultClient, "https://gamma-api.polymarket.com"))

	limit := 50
	closed := false
	ascending := false
	volumeMin := "1000"
	markets, err := gammaClient.GetMarkets(context.Background(), &gamma.MarketsRequest{
		Limit:     &limit,
		Closed:    &closed,
		Order:     "volume",
		Ascending: &ascending,
		VolumeMin: &volumeMin,
	})
	if err != nil {
		log.Fatalf("Failed to fetch markets: %v", err)
	}

	if len(markets) == 0 {
		log.Fatal("No active markets found")
	}

	fmt.Printf("Found %d markets\n", len(markets))
	for i, m := range markets {
		fmt.Printf("[%d] Question: %s, Volume: %s, Closed: %v\n", i, m.Question, m.Volume.String(), m.Closed)
	}

	// Extract a token ID (Asset ID)
	var market gamma.Market
	var tokenID string

	for _, m := range markets {
		// Skip low volume markets to ensure we get some WS events
		if m.Volume.IsZero() {
			continue
		}

		// Try Tokens first
		if len(m.Tokens) > 0 && !m.Closed {
			market = m
			tokenID = m.Tokens[0].TokenID
			break
		}
		// Try ClobTokenIds
		if m.ClobTokenIds != "" && !m.Closed {
			var ids []string
			if err := json.Unmarshal([]byte(m.ClobTokenIds), &ids); err == nil && len(ids) > 0 {
				market = m
				tokenID = ids[0]
				break
			}
		}
	}

	if tokenID == "" {
		log.Fatal("No tokens found in active high-volume markets")
	}

	fmt.Printf("Using Asset ID: %s (Market: %s)\n", tokenID, market.Question)

	// 2. Connect WS
	fmt.Println("Connecting to WebSocket...")
	wsClient, err := ws.NewClient("", nil, nil)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	// 3. Subscribe to Orderbook
	fmt.Printf("Subscribing to Orderbook for Asset ID: %s\n", tokenID)
	ch, err := wsClient.SubscribeOrderbook(context.Background(), []string{tokenID})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// 4. Read Loop
	go func() {
		for event := range ch {
			fmt.Printf("Received Orderbook Event: %+v\n", event)
		}
	}()

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("Shutting down...")
}

func boolPtr(b bool) *bool {
	return &b
}
