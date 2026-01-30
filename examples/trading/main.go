package main

import (
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	
"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

func main() {
	// 1. Setup Auth (Signer & API Key)
	// For demo, we generate a random private key if PK is not set
	pkHex := os.Getenv("POLYMARKET_PK")
	if pkHex == "" {
		key, _ := crypto.GenerateKey()
		pkHex = fmt.Sprintf("%x", crypto.FromECDSA(key))
		fmt.Println("Using generated private key (random)")
	}

	// ChainID 137 for Polygon Mainnet
	signer, err := auth.NewPrivateKeySigner(pkHex, 137)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	apiKey := &auth.APIKey{
		Key:        os.Getenv("POLYMARKET_API_KEY"),
		Secret:     os.Getenv("POLYMARKET_API_SECRET"),
		Passphrase: os.Getenv("POLYMARKET_API_PASSPHRASE"),
	}

	// 2. Initialize Client
	client := polymarket.NewClient()

	// 3. Create Authenticated CLOB Client
	authClient := client.CLOB.WithAuth(signer, apiKey)

	ctx := context.Background()

	// 4. Create an Order (Example)
	// Note: This requires valid TokenID and sufficient balance/allowance to succeed on server
	order := &clobtypes.Order{
		Maker:       types.Address(signer.Address()),
		Taker:       common.HexToAddress("0x0000000000000000000000000000000000000000"), // Open order
		TokenID:     types.U256{Int: big.NewInt(123456789)},                            // Dummy Token ID
		MakerAmount: decimal.NewFromFloat(10.0),
		TakerAmount: decimal.NewFromFloat(5.0),
		Side:        "BUY",
		FeeRateBps:  decimal.NewFromInt(0),
		Nonce:       types.U256{Int: big.NewInt(time.Now().Unix())},
	}

	fmt.Println("Signing and posting order...")
	resp, err := authClient.CreateOrder(ctx, order)
	if err != nil {
		// It is expected to fail if API keys are invalid or funds missing, but we want to see if Signing worked
		log.Printf("Order creation returned error (expected): %v", err)
	} else {
		fmt.Printf("Order Created! ID: %s\n", resp.ID)
	}
}
