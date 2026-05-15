package main

import (
	"context"
	"fmt"
	"log"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
)

func main() {
	client := polymarket.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream := clob.StreamData(ctx, func(ctx context.Context, cursor string) ([]clobtypes.Market, string, error) {
		resp, err := client.CLOB.Markets(ctx, &clobtypes.MarketsRequest{
			Limit:  3,
			Cursor: cursor,
		})
		if err != nil {
			return nil, "", err
		}
		return resp.Data, resp.NextCursor, nil
	})

	count := 0
	for res := range stream {
		if res.Err != nil {
			log.Fatalf("StreamData failed: %v", res.Err)
		}
		fmt.Printf("%d. %s\n", count+1, res.Item.Question)
		count++
		if count >= 5 {
			break
		}
	}
}
