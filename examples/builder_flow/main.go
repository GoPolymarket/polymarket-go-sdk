package main

import (
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob"
	
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
		log.Fatalf("failed to create signer: %v", err)
	}

	client := polymarket.NewClient(polymarket.WithUseServerTime(true))
	gammaClient := gamma.NewClient(nil)

	apiKey := loadAPIKeyFromEnv()
	if apiKey == nil || apiKey.Key == "" || apiKey.Secret == "" || apiKey.Passphrase == "" {
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

	authClient := client.CLOB.WithAuth(signer, apiKey)

	builderConfig := loadBuilderConfigFromEnv()
	if builderConfig == nil || !builderConfig.IsValid() {
		log.Println("No builder config provided, creating builder API key...")
		resp, err := authClient.CreateBuilderAPIKey(ctx)
		if err != nil {
			log.Fatalf("CreateBuilderAPIKey failed: %v", err)
		}
		builderConfig = &auth.BuilderConfig{
			Local: &auth.BuilderCredentials{
				Key:        resp.APIKey,
				Secret:     resp.Secret,
				Passphrase: resp.Passphrase,
			},
		}
	} else if builderConfig.Remote != nil {
		log.Printf("Using remote builder signer: %s", builderConfig.Remote.Host)
	}

	builderClient := authClient.WithBuilderConfig(builderConfig)

	tokenID, question, err := pickTokenID(ctx, gammaClient, client.CLOB)
	if err != nil {
		log.Fatalf("Failed to find token ID: %v", err)
	}
	log.Printf("Using token %s (market: %s)", tokenID, question)

	signable, err := clob.NewOrderBuilder(builderClient, signer).
		TokenID(tokenID).
		Side("BUY").
		Price(0.5).
		Size(1).
		OrderType(clobtypes.OrderTypeGTC).
		BuildSignable()
	if err != nil {
		log.Fatalf("BuildSignable failed: %v", err)
	}

	resp, err := builderClient.CreateOrderFromSignable(ctx, signable)
	if err != nil {
		log.Printf("Order creation returned error (expected in demo): %v", err)
	} else {
		log.Printf("Order Created! ID: %s", resp.ID)
	}

	trades, err := builderClient.BuilderTrades(ctx, &clobtypes.BuilderTradesRequest{
		Maker: signer.Address().Hex(),
		Limit: 5,
	})
	if err != nil {
		log.Printf("BuilderTrades failed: %v", err)
	} else {
		log.Printf("Builder trades returned %d records", len(trades.Data))
	}
}

func loadAPIKeyFromEnv() *auth.APIKey {
	key := os.Getenv("POLYMARKET_API_KEY")
	secret := os.Getenv("POLYMARKET_API_SECRET")
	passphrase := os.Getenv("POLYMARKET_API_PASSPHRASE")
	if key == "" && secret == "" && passphrase == "" {
		return nil
	}
	return &auth.APIKey{
		Key:        key,
		Secret:     secret,
		Passphrase: passphrase,
	}
}

func loadBuilderConfigFromEnv() *auth.BuilderConfig {
	remoteHost := strings.TrimSpace(os.Getenv("POLYMARKET_BUILDER_REMOTE_HOST"))
	if remoteHost != "" {
		return &auth.BuilderConfig{
			Remote: &auth.BuilderRemoteConfig{
				Host:  remoteHost,
				Token: strings.TrimSpace(os.Getenv("POLYMARKET_BUILDER_REMOTE_TOKEN")),
			},
		}
	}

	key := os.Getenv("POLYMARKET_BUILDER_API_KEY")
	secret := os.Getenv("POLYMARKET_BUILDER_API_SECRET")
	passphrase := os.Getenv("POLYMARKET_BUILDER_API_PASSPHRASE")
	if key == "" && secret == "" && passphrase == "" {
		return nil
	}
	return &auth.BuilderConfig{
		Local: &auth.BuilderCredentials{
			Key:        key,
			Secret:     secret,
			Passphrase: passphrase,
		},
	}
}

func pickTokenID(ctx context.Context, gammaClient gamma.Client, clobClient clob.Client) (string, string, error) {
	tokenID, question, err := pickTokenIDFromGamma(ctx, gammaClient)
	if err == nil && tokenID != "" {
		return tokenID, question, nil
	}
	if err != nil {
		log.Printf("Gamma lookup failed: %v", err)
	}
	return pickTokenIDFromCLOB(ctx, clobClient)
}

func pickTokenIDFromGamma(ctx context.Context, gammaClient gamma.Client) (string, string, error) {
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
		if !market.Active || market.Closed {
			continue
		}
		if tokenID := firstGammaTokenID(market); tokenID != "" {
			return tokenID, market.Question, nil
		}
	}
	return "", "", fmt.Errorf("no token IDs found in gamma markets")
}

func pickTokenIDFromCLOB(ctx context.Context, clobClient clob.Client) (string, string, error) {
	limit := 50
	active := true
	resp, err := clobClient.Markets(ctx, &clobtypes.MarketsRequest{
		Limit:  limit,
		Active: &active,
	})
	if err != nil {
		return "", "", err
	}

	for _, market := range resp.Data {
		if market.Closed {
			continue
		}
		for _, token := range market.Tokens {
			if token.TokenID != "" {
				return token.TokenID, market.Question, nil
			}
		}
	}
	return "", "", fmt.Errorf("no token IDs found in clob markets")
}

func firstGammaTokenID(market gamma.Market) string {
	for _, token := range market.Tokens {
		if token.TokenID != "" {
			return token.TokenID
		}
	}
	ids := parseClobTokenIDs(market.ClobTokenIds)
	if len(ids) > 0 {
		return ids[0]
	}
	return ""
}

func parseClobTokenIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err == nil {
		return ids
	}

	var anyIDs []interface{}
	if err := json.Unmarshal([]byte(raw), &anyIDs); err != nil {
		return nil
	}

	out := make([]string, 0, len(anyIDs))
	for _, entry := range anyIDs {
		switch v := entry.(type) {
		case string:
			if v != "" {
				out = append(out, v)
			}
		case float64:
			out = append(out, strconv.FormatInt(int64(v), 10))
		}
	}
	return out
}
