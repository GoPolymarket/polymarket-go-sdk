package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/gamma"
)

func main() {
	ctx := context.Background()

	pkHex := os.Getenv("POLYMARKET_PK")
	if pkHex == "" {
		log.Fatal("POLYMARKET_PK is required")
	}

	signer, err := auth.NewPrivateKeySigner(pkHex, 137)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	client := polymarket.NewClient(polymarket.WithUseServerTime(true))

	apiKey := &auth.APIKey{
		Key:        os.Getenv("POLYMARKET_API_KEY"),
		Secret:     os.Getenv("POLYMARKET_API_SECRET"),
		Passphrase: os.Getenv("POLYMARKET_API_PASSPHRASE"),
	}
	if apiKey.Key == "" || apiKey.Secret == "" || apiKey.Passphrase == "" {
		log.Println("No L2 API key provided, deriving via L1 signature...")
		l1Client := client.CLOB.WithAuth(signer, nil)
		resp, err := l1Client.DeriveAPIKey(ctx)
		if err != nil {
			log.Fatalf("DeriveAPIKey failed: %v", err)
		}
		apiKey = &auth.APIKey{
			Key:        resp.APIKey,
			Secret:     resp.Secret,
			Passphrase: resp.Passphrase,
		}
	}

	marketID, question, err := pickMarketID(ctx)
	if err != nil {
		log.Fatalf("Failed to find market ID: %v", err)
	}
	log.Printf("Using market %s (%s)", marketID, question)

	wsClient, err := ws.NewClient("", signer, apiKey)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	orderCh, err := wsClient.SubscribeUserOrders(ctx, []string{marketID})
	if err != nil {
		log.Fatalf("SubscribeUserOrders failed: %v", err)
	}
	tradeCh, err := wsClient.SubscribeUserTrades(ctx, []string{marketID})
	if err != nil {
		log.Fatalf("SubscribeUserTrades failed: %v", err)
	}

	go func() {
		for event := range orderCh {
			fmt.Printf("Order Event: %+v\n", event)
		}
	}()
	go func() {
		for event := range tradeCh {
			fmt.Printf("Trade Event: %+v\n", event)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("Shutting down...")
}

func pickMarketID(ctx context.Context) (string, string, error) {
	gammaClient := gamma.NewClient(nil)
	limit := 50
	active := true
	closed := false
	markets, err := gammaClient.GetMarkets(ctx, &gamma.MarketsRequest{
		Limit:  &limit,
		Active: &active,
		Closed: &closed,
		Order:  "volume",
	})
	if err != nil {
		return "", "", err
	}

	for _, market := range markets {
		if market.ConditionID != "" && market.Active && !market.Closed {
			return market.ConditionID, market.Question, nil
		}
	}

	return "", "", fmt.Errorf("no market IDs found")
}
