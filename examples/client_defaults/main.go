package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

func main() {
	pk := os.Getenv("POLYMARKET_PK")
	if pk == "" {
		log.Println("POLYMARKET_PK is required to run this example")
		return
	}

	signer, err := auth.NewPrivateKeySigner(pk, auth.PolygonChainID)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	client := polymarket.NewClient()
	clobClient := client.CLOB.
		WithAuth(signer, nil).
		WithAuthNonce(1).
		WithSignatureType(auth.SignatureProxy).
		WithSaltGenerator(func() (*big.Int, error) {
			return big.NewInt(42), nil
		})

	if funderHex := os.Getenv("POLYMARKET_FUNDER"); funderHex != "" {
		clobClient = clobClient.WithFunder(common.HexToAddress(funderHex))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("1. Build a signable order using client defaults")
	builder := clob.NewOrderBuilder(clobClient, signer)
	signable, err := builder.
		TokenID("123456").
		Side("BUY").
		Price(0.5).
		Size(10).
		TickSize("0.01").
		FeeRateBps(0).
		BuildSignableWithContext(ctx)
	if err != nil {
		log.Fatalf("BuildSignable failed: %v", err)
	}
	fmt.Printf("Maker: %s\n", signable.Order.Maker.String())
	fmt.Printf("SignatureType: %d\n", *signable.Order.SignatureType)
	fmt.Printf("Salt: %s\n", signable.Order.Salt.Int.String())

	if os.Getenv("POLYMARKET_CREATE_API_KEY") == "1" {
		fmt.Println("\n2. Create API key using default auth nonce")
		resp, err := clobClient.CreateAPIKey(ctx)
		if err != nil {
			log.Fatalf("CreateAPIKey failed: %v", err)
		}
		fmt.Println("API key created successfully.")
	}

	fmt.Println("\n3. Example order payload (no submit)")
	order := &clobtypes.Order{
		Maker:         signable.Order.Maker,
		Signer:        signer.Address(),
		Taker:         common.Address{},
		TokenID:       types.U256{Int: big.NewInt(123456)},
		MakerAmount:   signable.Order.MakerAmount,
		TakerAmount:   signable.Order.TakerAmount,
		Expiration:    signable.Order.Expiration,
		Nonce:         signable.Order.Nonce,
		FeeRateBps:    signable.Order.FeeRateBps,
		Side:          signable.Order.Side,
		SignatureType: signable.Order.SignatureType,
		Salt:          signable.Order.Salt,
	}
	fmt.Printf("Order side: %s, maker=%s\n", order.Side, order.Maker.String())
}
