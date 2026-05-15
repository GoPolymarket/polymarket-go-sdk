package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/rfq"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
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

	apiKey := &auth.APIKey{
		Key:        os.Getenv("POLYMARKET_API_KEY"),
		Secret:     os.Getenv("POLYMARKET_API_SECRET"),
		Passphrase: os.Getenv("POLYMARKET_API_PASSPHRASE"),
	}
	if apiKey.Key == "" || apiKey.Secret == "" || apiKey.Passphrase == "" {
		log.Fatal("POLYMARKET_API_KEY/POLYMARKET_API_SECRET/POLYMARKET_API_PASSPHRASE are required")
	}

	client := transport.NewClient(nil, "https://clob.polymarket.com")
	client.SetAuth(signer, apiKey)

	rfqClient := rfq.NewClient(client)

	query := &rfq.RFQQuotesQuery{
		Limit: 5,
	}

	if ids := splitCSV(os.Getenv("RFQ_REQUEST_IDS")); len(ids) > 0 {
		query.RequestIDs = ids
	}
	if ids := splitCSV(os.Getenv("RFQ_QUOTE_IDS")); len(ids) > 0 {
		query.QuoteIDs = ids
	}
	if markets := splitCSV(os.Getenv("RFQ_MARKETS")); len(markets) > 0 {
		query.Markets = markets
	}
	if state := strings.TrimSpace(os.Getenv("RFQ_STATE")); state != "" {
		query.State = rfq.RFQState(strings.ToLower(state))
	}

	quotes, err := rfqClient.RFQQuotes(ctx, query)
	if err != nil {
		log.Fatalf("RFQQuotes failed: %v", err)
	}

	fmt.Printf("Found %d RFQ quotes\n", len(quotes))
	for _, q := range quotes {
		id := q.QuoteID
		if id == "" {
			id = q.ID
		}
		fmt.Printf("- quote=%s request=%s side=%s price=%s\n", id, q.RequestID, q.Side, q.Price)
	}
}

func splitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
