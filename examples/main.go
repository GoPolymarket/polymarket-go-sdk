package main

import (
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"context"
	"fmt"
	"log"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	
)

func main() {
	client := polymarket.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("1. Fetching Server Time...")
	timeResp, err := client.CLOB.Time(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get time: %v (Network might be unreachable)", err)
	} else {
		fmt.Printf("Server Time: %s (ts=%d)\n", timeResp.ServerTime, timeResp.Timestamp)
	}

	fmt.Println("\n2. Fetching Active Markets (Limit 2)...")
	limit := 2
	active := true
	marketsResp, err := client.CLOB.Markets(ctx, &clobtypes.MarketsRequest{
		Limit:  limit,
		Active: &active,
	})
	if err != nil {
		log.Printf("Warning: Failed to get markets: %v (Network might be unreachable)", err)
	} else {
		fmt.Printf("Fetched %d markets (Total: %d):\n", len(marketsResp.Data), marketsResp.Count)
		for _, m := range marketsResp.Data {
			fmt.Printf("- [ConditionID: %s] %s\n", m.ConditionID, m.Question)
		}
	}
}
