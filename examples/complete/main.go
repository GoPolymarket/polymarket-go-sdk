package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	clobtypes "github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func main() {
	fmt.Println("=== Polymarket Go SDK Examples ===")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client := polymarket.NewClient()

	if err := runMarketDataExamples(ctx, client); err != nil {
		log.Printf("Market data examples error: %v", err)
	}

	if err := runOrderBookExamples(ctx, client); err != nil {
		log.Printf("Order book examples error: %v", err)
	}

	if err := runAccountExamples(ctx, client); err != nil {
		log.Printf("Account examples error: %v", err)
	}

	if err := runAuthExamples(ctx); err != nil {
		log.Printf("Auth examples error: %v", err)
	}

	fmt.Println("\n=== All Examples Completed ===")
}

func runMarketDataExamples(ctx context.Context, client *polymarket.Client) error {
	fmt.Println("\n--- Market Data Examples ---")

	health, err := client.CLOB.Health(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	fmt.Printf("1. Health: %s\n", health)

	timeResp, err := client.CLOB.Time(ctx)
	if err != nil {
		return fmt.Errorf("get time failed: %w", err)
	}
	fmt.Printf("2. Server Time: %d\n", timeResp.Timestamp)

	marketsResp, err := client.CLOB.Markets(ctx, &clobtypes.MarketsRequest{
		Active: ptrBool(true),
		Limit:  5,
	})
	if err != nil {
		return fmt.Errorf("get markets failed: %w", err)
	}
	fmt.Printf("3. Markets found: %d\n", len(marketsResp.Data))
	if len(marketsResp.Data) > 0 {
		fmt.Printf("   First market: %s\n", marketsResp.Data[0].Question)
	}

	allMarkets, err := client.CLOB.MarketsAll(ctx, &clobtypes.MarketsRequest{
		Active: ptrBool(true),
	})
	if err != nil {
		return fmt.Errorf("get all markets failed: %w", err)
	}
	fmt.Printf("4. Total active markets: %d\n", len(allMarkets))

	simplifiedResp, err := client.CLOB.SimplifiedMarkets(ctx, &clobtypes.MarketsRequest{
		Active: ptrBool(true),
		Limit:  3,
	})
	if err != nil {
		return fmt.Errorf("get simplified markets failed: %w", err)
	}
	fmt.Printf("5. Simplified markets: %d\n", len(simplifiedResp.Data))

	allPrices, err := client.CLOB.AllPrices(ctx)
	if err != nil {
		return fmt.Errorf("get all prices failed: %w", err)
	}
	fmt.Printf("6. Total prices: %d\n", len(allPrices))

	return nil
}

func runOrderBookExamples(ctx context.Context, client *polymarket.Client) error {
	fmt.Println("\n--- Order Book Examples ---")

	marketsResp, err := client.CLOB.Markets(ctx, &clobtypes.MarketsRequest{
		Active: ptrBool(true),
		Limit:  1,
	})
	if err != nil || len(marketsResp.Data) == 0 {
		return fmt.Errorf("no markets available: %w", err)
	}

	tokenID := marketsResp.Data[0].Tokens[0].TokenID
	fmt.Printf("Using token: %s\n", tokenID)

	book, err := client.CLOB.OrderBook(ctx, &clobtypes.BookRequest{
		TokenID: tokenID,
	})
	if err != nil {
		return fmt.Errorf("get orderbook failed: %w", err)
	}
	fmt.Printf("1. OrderBook - Asks: %d, Bids: %d\n", len(book.Asks), len(book.Bids))

	books, err := client.CLOB.OrderBooks(ctx, &clobtypes.BooksRequest{
		TokenIDs: []string{tokenID},
	})
	if err != nil {
		return fmt.Errorf("get orderbooks failed: %w", err)
	}
	fmt.Printf("2. OrderBooks response count: %d\n", len(books))

	midpoint, err := client.CLOB.Midpoint(ctx, &clobtypes.MidpointRequest{
		TokenID: tokenID,
	})
	if err != nil {
		return fmt.Errorf("get midpoint failed: %w", err)
	}
	fmt.Printf("3. Midpoint price: %s\n", midpoint.Midpoint)

	priceResp, err := client.CLOB.Price(ctx, &clobtypes.PriceRequest{
		TokenID: tokenID,
		Side:    "BUY",
	})
	if err != nil {
		return fmt.Errorf("get price failed: %w", err)
	}
	fmt.Printf("4. Best BUY price: %s\n", priceResp.Price)

	spread, err := client.CLOB.Spread(ctx, &clobtypes.SpreadRequest{
		TokenID: tokenID,
	})
	if err != nil {
		return fmt.Errorf("get spread failed: %w", err)
	}
	fmt.Printf("5. Spread: %s\n", spread.Spread)

	lastTrade, err := client.CLOB.LastTradePrice(ctx, &clobtypes.LastTradePriceRequest{
		TokenID: tokenID,
	})
	if err != nil {
		return fmt.Errorf("get last trade price failed: %w", err)
	}
	fmt.Printf("6. Last trade price: %s\n", lastTrade.Price)

	tickSize, err := client.CLOB.TickSize(ctx, &clobtypes.TickSizeRequest{
		TokenID: tokenID,
	})
	if err != nil {
		return fmt.Errorf("get tick size failed: %w", err)
	}
	fmt.Printf("7. Tick size: %f\n", tickSize.TickSize)

	feeRate, err := client.CLOB.FeeRate(ctx, &clobtypes.FeeRateRequest{
		TokenID: tokenID,
	})
	if err != nil {
		return fmt.Errorf("get fee rate failed: %w", err)
	}
	fmt.Printf("8. Fee rate: %s\n", feeRate.FeeRate)

	return nil
}

func runAccountExamples(ctx context.Context, client *polymarket.Client) error {
	fmt.Println("\n--- Account Examples (require auth) ---")

	pk := os.Getenv("POLYMARKET_PK")
	if pk == "" {
		fmt.Println("SKIPPED: POLYMARKET_PK not set")
		return nil
	}

	signer, err := auth.NewPrivateKeySigner(pk, 137)
	if err != nil {
		return fmt.Errorf("create signer failed: %w", err)
	}

	authClient := client.CLOB.WithAuth(signer, nil)

	bal, err := authClient.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{
		AssetType: "conditional",
	})
	if err != nil {
		return fmt.Errorf("get balance failed: %w", err)
	}
	fmt.Printf("1. Balance: %s\n", bal.Balance)

	orders, err := authClient.Orders(ctx, &clobtypes.OrdersRequest{
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("get orders failed: %w", err)
	}
	fmt.Printf("2. Open orders: %d\n", len(orders.Data))

	trades, err := authClient.Trades(ctx, &clobtypes.TradesRequest{
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("get trades failed: %w", err)
	}
	fmt.Printf("3. Recent trades: %d\n", len(trades.Data))

	notifs, err := authClient.Notifications(ctx, &clobtypes.NotificationsRequest{
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("get notifications failed: %w", err)
	}
	fmt.Printf("4. Notifications: %d\n", len(notifs))

	earnings, err := authClient.UserEarnings(ctx, &clobtypes.UserEarningsRequest{})
	if err != nil {
		return fmt.Errorf("get earnings failed: %w", err)
	}
	fmt.Printf("5. User earnings count: %d\n", len(earnings.Data))

	return nil
}

func runAuthExamples(ctx context.Context) error {
	fmt.Println("\n--- Authentication Examples ---")

	pk := os.Getenv("POLYMARKET_PK")
	if pk == "" {
		fmt.Println("SKIPPED: POLYMARKET_PK not set")
		return nil
	}

	signer, err := auth.NewPrivateKeySigner(pk, 137)
	if err != nil {
		return fmt.Errorf("create signer failed: %w", err)
	}
	fmt.Printf("1. Signer address: %s\n", signer.Address().Hex())

	proxyAddr, err := auth.DeriveProxyWallet(signer.Address())
	if err != nil {
		fmt.Printf("2. Proxy wallet: (error: %v)\n", err)
	} else {
		fmt.Printf("2. Proxy wallet: %s\n", proxyAddr.Hex())
	}

	safeAddr, err := auth.DeriveSafeWallet(signer.Address())
	if err != nil {
		fmt.Printf("3. Safe wallet: (error: %v)\n", err)
	} else {
		fmt.Printf("3. Safe wallet: %s\n", safeAddr.Hex())
	}

	secret := "dGVzdF9zZWNyZXQ"
	message := "POST\n/v1/orders\n{}"
	hmacSig, err := auth.SignHMAC(secret, message)
	if err != nil {
		return fmt.Errorf("HMAC sign failed: %w", err)
	}
	fmt.Printf("4. HMAC signature: %s...\n", hmacSig[:20])

	headers, err := auth.BuildL1Headers(signer, time.Now().Unix(), 1)
	if err != nil {
		return fmt.Errorf("build L1 headers failed: %w", err)
	}
	fmt.Printf("5. L1 headers: POLY_ADDRESS=%s\n", headers.Get("POLY_ADDRESS"))

	return nil
}

func ptrBool(b bool) *bool {
	return &b
}
