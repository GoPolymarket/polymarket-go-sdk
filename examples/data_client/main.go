package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/data"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	client := polymarket.NewClient()
	ctx := context.Background()

	health, err := client.Data.Health(ctx)
	if err != nil {
		log.Printf("Data API health error: %v", err)
	} else {
		fmt.Printf("Data API health: %s\n", health)
	}

	userAddr := strings.TrimSpace(os.Getenv("DATA_USER_ADDRESS"))
	if userAddr == "" {
		fmt.Println("Set DATA_USER_ADDRESS to run user-specific queries.")
		return
	}
	if !common.IsHexAddress(userAddr) {
		log.Fatalf("Invalid DATA_USER_ADDRESS: %s", userAddr)
	}
	user := common.HexToAddress(userAddr)

	limit := 5
	positions, err := client.Data.Positions(ctx, &data.PositionsRequest{
		User:  user,
		Limit: &limit,
	})
	if err != nil {
		log.Printf("Positions error: %v", err)
	} else {
		fmt.Printf("Positions: %d\n", len(positions))
	}

	trades, err := client.Data.Trades(ctx, &data.TradesRequest{
		User:  &user,
		Limit: &limit,
	})
	if err != nil {
		log.Printf("Trades error: %v", err)
	} else {
		fmt.Printf("Trades: %d\n", len(trades))
	}

	traded, err := client.Data.Traded(ctx, &data.TradedRequest{User: user})
	if err != nil {
		log.Printf("Traded error: %v", err)
	} else {
		fmt.Printf("Unique markets traded: %d\n", traded.Traded)
	}
}
