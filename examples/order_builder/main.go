package main

import (
	"fmt"
	"log"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
)

func main() {
	// 1. Setup Dummy Signer
	// This is a random private key for testing purposes only. DO NOT USE IN PRODUCTION.
	privateKeyHex := "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	signer, err := auth.NewPrivateKeySigner(privateKeyHex, 137)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	fmt.Printf("Signer Address: %s\n", signer.Address().String())

	// 2. Derive Proxy Address (Just to verify)
	proxy, err := auth.DeriveProxyWallet(signer.Address())
	if err != nil {
		log.Fatalf("Failed to derive proxy: %v", err)
	}
	fmt.Printf("Derived Proxy Address: %s\n", proxy.String())

	// 3. Initialize Client (Mock/Nil transport as we are just building order)
	// We pass nil for transport because NewOrderBuilder only needs the interface.
	// We'll provide a manual tick size below to avoid network calls in this example.
	// However, NewOrderBuilder takes `clob.Client`.
	// Let's just create a basic client.
	client := clob.NewClient(nil)

	// 4. Use OrderBuilder
	fmt.Println("\nBuilding Order...")
	builder := clob.NewOrderBuilder(client, signer)

	order, err := builder.
		TokenID("1234567890").
		Side("BUY").
		Price(0.5).
		Size(100.0).
		FeeRateBps(0).
		TickSize(0.01).
		UseProxy(). // Important: Use Proxy Wallet
		Build()

	if err != nil {
		log.Fatalf("Failed to build order: %v", err)
	}

	// 5. Verify Fields
	fmt.Println("Order Built Successfully!")
	fmt.Printf("Salt: %s\n", order.Salt.Int.String())
	fmt.Printf("Signer: %s\n", order.Signer.String())
	fmt.Printf("Maker: %s\n", order.Maker.String())
	fmt.Printf("Taker: %s\n", order.Taker.String())
	fmt.Printf("TokenID: %s\n", order.TokenID.Int.String())
	fmt.Printf("MakerAmount: %s\n", order.MakerAmount.String())
	fmt.Printf("TakerAmount: %s\n", order.TakerAmount.String())
	fmt.Printf("Expiration: %s\n", order.Expiration.Int.String())
	fmt.Printf("Side: %s\n", order.Side)
	fmt.Printf("SignatureType: %d\n", *order.SignatureType)

	// Verify Maker matches Derived Proxy
	if order.Maker != proxy {
		log.Fatalf("Error: Maker address %s does not match derived proxy %s", order.Maker.String(), proxy.String())
	} else {
		fmt.Println("SUCCESS: Maker address matches derived proxy.")
	}

	// Verify Salt is non-zero
	if order.Salt.Int.Sign() == 0 {
		log.Fatalf("Error: Salt is zero")
	} else {
		fmt.Println("SUCCESS: Salt is generated.")
	}

	// Verify Expiration is set (0 for GTC)
	if order.Expiration.Int == nil {
		log.Fatalf("Error: Expiration is nil")
	} else {
		fmt.Println("SUCCESS: Expiration is set.")
	}
}
