package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/rtds"
)

func main() {
	// 1. Connect to RTDS (Real-Time Data Service)
	fmt.Println("Connecting to RTDS WebSocket...")
	client, err := rtds.NewClient("") // Use default ProdURL
	if err != nil {
		log.Fatalf("Failed to connect to RTDS: %v", err)
	}
	defer client.Close()

	// 2. Subscribe to Crypto Prices
	// Symbols follow Binance pairs like "btcusdt", "ethusdt"
	symbols := []string{"btcusdt", "ethusdt"}
	fmt.Printf("Subscribing to Crypto Prices for: %v\n", symbols)

	priceCh, err := client.SubscribeCryptoPrices(context.Background(), symbols)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// 3. Read Loop
	go func() {
		for event := range priceCh {
			fmt.Printf("[RTDS] %s Price: %s (ts=%d)\n", event.Symbol, event.Value.String(), event.Timestamp)
		}
	}()

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("Shutting down...")
}
