package main

import (
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob"
	
"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/rfq"
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

	client := polymarket.NewClient(polymarket.WithUseServerTime(true))

	apiKey := &auth.APIKey{
		Key:        os.Getenv("POLYMARKET_API_KEY"),
		Secret:     os.Getenv("POLYMARKET_API_SECRET"),
		Passphrase: os.Getenv("POLYMARKET_API_PASSPHRASE"),
	}
	if apiKey.Key == "" || apiKey.Secret == "" || apiKey.Passphrase == "" {
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
	rfqClient := authClient.RFQ()

	assetIn := os.Getenv("RFQ_ASSET_IN")
	assetOut := os.Getenv("RFQ_ASSET_OUT")
	amountIn := os.Getenv("RFQ_AMOUNT_IN")
	amountOut := os.Getenv("RFQ_AMOUNT_OUT")
	userType := strings.TrimSpace(os.Getenv("RFQ_USER_TYPE"))
	if userType == "" {
		userType = "EOA"
	}

	if assetIn == "" || assetOut == "" || amountIn == "" || amountOut == "" {
		log.Fatal("RFQ_ASSET_IN/RFQ_ASSET_OUT/RFQ_AMOUNT_IN/RFQ_AMOUNT_OUT are required")
	}

	reqResp, err := rfqClient.CreateRFQRequest(ctx, &rfq.RFQRequest{
		AssetIn:   assetIn,
		AssetOut:  assetOut,
		AmountIn:  amountIn,
		AmountOut: amountOut,
		UserType:  userType,
	})
	if err != nil {
		log.Fatalf("CreateRFQRequest failed: %v", err)
	}

	requestID := reqResp.RequestID
	if requestID == "" {
		requestID = reqResp.ID
	}
	fmt.Printf("Created RFQ request: %s (expiry=%d)\n", requestID, reqResp.Expiry)

	quoteResp, err := rfqClient.CreateRFQQuote(ctx, &rfq.RFQQuote{
		RequestIDV2: requestID,
		AssetIn:     assetIn,
		AssetOut:    assetOut,
		AmountIn:    amountIn,
		AmountOut:   amountOut,
		UserType:    userType,
	})
	if err != nil {
		log.Printf("CreateRFQQuote failed: %v", err)
	} else {
		quoteID := quoteResp.QuoteID
		if quoteID == "" {
			quoteID = quoteResp.ID
		}
		fmt.Printf("Created RFQ quote: %s\n", quoteID)
	}

	reqs, err := rfqClient.RFQRequests(ctx, &rfq.RFQRequestsQuery{
		Limit:   5,
		State:   rfq.RFQStateActive,
		SortBy:  rfq.RFQSortByCreated,
		SortDir: rfq.RFQSortDirDesc,
	})
	if err != nil {
		log.Printf("RFQRequests failed: %v", err)
	} else if len(reqs) > 0 {
		detail, err := reqs[0].ToDetail()
		if err == nil {
			fmt.Printf("Sample RFQ request detail: %+v\n", detail)
		}
	}

	log.Println("Accept/Approve require signed order fields; provide env to enable:")
	log.Println("RFQ_ACCEPT_REQUEST_ID/RFQ_ACCEPT_QUOTE_ID/RFQ_ACCEPT_MAKER_AMOUNT/RFQ_ACCEPT_TAKER_AMOUNT/RFQ_ACCEPT_TOKEN_ID/")
	log.Println("RFQ_ACCEPT_MAKER/RFQ_ACCEPT_SIGNER/RFQ_ACCEPT_TAKER/RFQ_ACCEPT_NONCE/RFQ_ACCEPT_EXPIRATION/")
	log.Println("RFQ_ACCEPT_SIDE/RFQ_ACCEPT_FEE_RATE_BPS/RFQ_ACCEPT_SIGNATURE/RFQ_ACCEPT_SALT/RFQ_ACCEPT_OWNER")

	if signed, err := buildSignedOrderFromEnv(authClient, signer, apiKey); err != nil {
		log.Printf("RFQ signed order build failed: %v", err)
	} else if signed != nil {
		if requestID := os.Getenv("RFQ_ACCEPT_REQUEST_ID"); requestID != "" {
			if quoteID := os.Getenv("RFQ_ACCEPT_QUOTE_ID"); quoteID != "" {
				req, err := rfq.BuildRFQAcceptRequestFromSignedOrder(requestID, quoteID, signed)
				if err != nil {
					log.Printf("BuildRFQAcceptRequestFromSignedOrder failed: %v", err)
				} else if _, err := rfqClient.RFQRequestAccept(ctx, req); err != nil {
					log.Printf("RFQRequestAccept failed: %v", err)
				} else {
					log.Printf("RFQRequestAccept submitted (from signed order)")
				}
			}
		}
		if requestID := os.Getenv("RFQ_APPROVE_REQUEST_ID"); requestID != "" {
			if quoteID := os.Getenv("RFQ_APPROVE_QUOTE_ID"); quoteID != "" {
				req, err := rfq.BuildRFQApproveQuoteFromSignedOrder(requestID, quoteID, signed)
				if err != nil {
					log.Printf("BuildRFQApproveQuoteFromSignedOrder failed: %v", err)
				} else if _, err := rfqClient.RFQQuoteApprove(ctx, req); err != nil {
					log.Printf("RFQQuoteApprove failed: %v", err)
				} else {
					log.Printf("RFQQuoteApprove submitted (from signed order)")
				}
			}
		}
	}

	if acceptReq := loadAcceptRequestFromEnv(); acceptReq != nil {
		if _, err := rfqClient.RFQRequestAccept(ctx, acceptReq); err != nil {
			log.Printf("RFQRequestAccept failed: %v", err)
		} else {
			log.Printf("RFQRequestAccept submitted")
		}
	}

	if approveReq := loadApproveRequestFromEnv(); approveReq != nil {
		if _, err := rfqClient.RFQQuoteApprove(ctx, approveReq); err != nil {
			log.Printf("RFQQuoteApprove failed: %v", err)
		} else {
			log.Printf("RFQQuoteApprove submitted")
		}
	}
}

func buildSignedOrderFromEnv(client clob.Client, signer auth.Signer, apiKey *auth.APIKey) (*clobtypes.SignedOrder, error) {
	tokenID := os.Getenv("RFQ_SIGN_TOKEN_ID")
	if tokenID == "" {
		return nil, nil
	}
	side := os.Getenv("RFQ_SIGN_SIDE")
	priceStr := os.Getenv("RFQ_SIGN_PRICE")
	sizeStr := os.Getenv("RFQ_SIGN_SIZE")
	if side == "" || priceStr == "" || sizeStr == "" {
		return nil, fmt.Errorf("RFQ_SIGN_SIDE/RFQ_SIGN_PRICE/RFQ_SIGN_SIZE are required when RFQ_SIGN_TOKEN_ID is set")
	}

	price, err := parseFloat(priceStr)
	if err != nil {
		return nil, err
	}
	size, err := parseFloat(sizeStr)
	if err != nil {
		return nil, err
	}

	builder := clob.NewOrderBuilder(client, signer).
		TokenID(tokenID).
		Side(side).
		Price(price).
		Size(size)

	if orderType := strings.TrimSpace(os.Getenv("RFQ_SIGN_ORDER_TYPE")); orderType != "" {
		builder.OrderType(clobtypes.OrderType(strings.ToUpper(orderType)))
	}
	if postOnlyRaw := strings.TrimSpace(os.Getenv("RFQ_SIGN_POST_ONLY")); postOnlyRaw != "" {
		postOnly := strings.EqualFold(postOnlyRaw, "true") || postOnlyRaw == "1"
		builder.PostOnly(postOnly)
	}

	signable, err := builder.BuildSignable()
	if err != nil {
		return nil, err
	}

	signed, err := clob.SignOrder(signer, apiKey, signable.Order)
	if err != nil {
		return nil, err
	}
	signed.OrderType = signable.OrderType
	signed.PostOnly = signable.PostOnly
	return signed, nil
}

func parseFloat(raw string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %q: %w", raw, err)
	}
	return value, nil
}

func loadAcceptRequestFromEnv() *rfq.RFQAcceptRequest {
	requestID := os.Getenv("RFQ_ACCEPT_REQUEST_ID")
	quoteID := os.Getenv("RFQ_ACCEPT_QUOTE_ID")
	makerAmount := os.Getenv("RFQ_ACCEPT_MAKER_AMOUNT")
	takerAmount := os.Getenv("RFQ_ACCEPT_TAKER_AMOUNT")
	tokenID := os.Getenv("RFQ_ACCEPT_TOKEN_ID")
	maker := os.Getenv("RFQ_ACCEPT_MAKER")
	signer := os.Getenv("RFQ_ACCEPT_SIGNER")
	expiration := os.Getenv("RFQ_ACCEPT_EXPIRATION")
	side := os.Getenv("RFQ_ACCEPT_SIDE")
	signature := os.Getenv("RFQ_ACCEPT_SIGNATURE")
	salt := os.Getenv("RFQ_ACCEPT_SALT")
	owner := os.Getenv("RFQ_ACCEPT_OWNER")

	if requestID == "" || quoteID == "" || makerAmount == "" || takerAmount == "" || tokenID == "" || maker == "" || signer == "" || expiration == "" || side == "" || signature == "" || salt == "" || owner == "" {
		return nil
	}

	return &rfq.RFQAcceptRequest{
		RequestID:   requestID,
		QuoteIDV2:   quoteID,
		MakerAmount: makerAmount,
		TakerAmount: takerAmount,
		TokenID:     tokenID,
		Maker:       maker,
		Signer:      signer,
		Expiration:  expiration,
		Side:        side,
		Signature:   signature,
		Salt:        salt,
		Owner:       owner,
	}
}

func loadApproveRequestFromEnv() *rfq.RFQApproveQuote {
	requestID := os.Getenv("RFQ_APPROVE_REQUEST_ID")
	quoteID := os.Getenv("RFQ_APPROVE_QUOTE_ID")
	makerAmount := os.Getenv("RFQ_APPROVE_MAKER_AMOUNT")
	takerAmount := os.Getenv("RFQ_APPROVE_TAKER_AMOUNT")
	tokenID := os.Getenv("RFQ_APPROVE_TOKEN_ID")
	maker := os.Getenv("RFQ_APPROVE_MAKER")
	signer := os.Getenv("RFQ_APPROVE_SIGNER")
	expiration := os.Getenv("RFQ_APPROVE_EXPIRATION")
	side := os.Getenv("RFQ_APPROVE_SIDE")
	signature := os.Getenv("RFQ_APPROVE_SIGNATURE")
	salt := os.Getenv("RFQ_APPROVE_SALT")
	owner := os.Getenv("RFQ_APPROVE_OWNER")

	if requestID == "" || quoteID == "" || makerAmount == "" || takerAmount == "" || tokenID == "" || maker == "" || signer == "" || expiration == "" || side == "" || signature == "" || salt == "" || owner == "" {
		return nil
	}

	return &rfq.RFQApproveQuote{
		RequestID:   requestID,
		QuoteIDV2:   quoteID,
		MakerAmount: makerAmount,
		TakerAmount: takerAmount,
		TokenID:     tokenID,
		Maker:       maker,
		Signer:      signer,
		Expiration:  expiration,
		Side:        side,
		Signature:   signature,
		Salt:        salt,
		Owner:       owner,
	}
}
