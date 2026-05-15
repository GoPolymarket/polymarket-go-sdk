package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
)

func main() {
	pk := os.Getenv("POLYMARKET_PK")
	apiKey := os.Getenv("POLYMARKET_API_KEY")
	apiSecret := os.Getenv("POLYMARKET_API_SECRET")
	apiPassphrase := os.Getenv("POLYMARKET_API_PASSPHRASE")
	if pk == "" || apiKey == "" || apiSecret == "" || apiPassphrase == "" {
		log.Println("Missing credentials. Set POLYMARKET_PK, POLYMARKET_API_KEY, POLYMARKET_API_SECRET, POLYMARKET_API_PASSPHRASE to run this example.")
		return
	}

	signer, err := auth.NewPrivateKeySigner(pk, auth.PolygonChainID)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	creds := &auth.APIKey{
		Key:        apiKey,
		Secret:     apiSecret,
		Passphrase: apiPassphrase,
	}

	client := polymarket.NewClient()
	authClient := client.CLOB.WithAuth(signer, creds)

	sigType := auth.SignatureEOA
	if rawSigType := os.Getenv("POLYMARKET_SIGNATURE_TYPE"); rawSigType != "" {
		val, err := strconv.Atoi(rawSigType)
		if err != nil {
			log.Printf("Invalid POLYMARKET_SIGNATURE_TYPE=%q (expected 0,1,2). Using default EOA.", rawSigType)
		} else {
			sigType = auth.SignatureType(val)
		}
	}
	authClient = authClient.WithSignatureType(sigType)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("1. Balance/Allowance (legacy asset=USDC)")
	balanceResp, err := authClient.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{
		Asset: "USDC",
	})
	if err != nil {
		log.Printf("BalanceAllowance failed: %v", err)
	} else {
		fmt.Printf("Balance: %s\n", balanceResp.Balance)
		fmt.Printf("Allowances: %+v\n", balanceResp.Allowances)
	}

	if tokenID := os.Getenv("POLYMARKET_TOKEN_ID"); tokenID != "" {
		fmt.Println("\n2. Balance/Allowance (asset_type=CONDITIONAL + token_id)")
		conditionalResp, err := authClient.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{
			AssetType: clobtypes.AssetTypeConditional,
			TokenID:   tokenID,
		})
		if err != nil {
			log.Printf("Conditional BalanceAllowance failed: %v", err)
		} else {
			fmt.Printf("Balance: %s\n", conditionalResp.Balance)
			fmt.Printf("Allowances: %+v\n", conditionalResp.Allowances)
		}
	} else {
		fmt.Println("\n2. Skipping conditional balance/allowance (set POLYMARKET_TOKEN_ID to enable)")
	}

	fmt.Println("\n3. User Rewards By Market")
	rewardsDate := time.Now().UTC().Format("2006-01-02")
	rewardsResp, err := authClient.UserRewardsByMarket(ctx, &clobtypes.UserRewardsByMarketRequest{
		Date: rewardsDate,
	})
	if err != nil {
		log.Printf("UserRewardsByMarket failed: %v", err)
		return
	}

	fmt.Printf("Rewards entries: %d\n", len(rewardsResp))
	if len(rewardsResp) > 0 {
		fmt.Printf("First entry: condition_id=%s earning_pct=%s\n", rewardsResp[0].ConditionID, rewardsResp[0].EarningPercentage)
	}
}
