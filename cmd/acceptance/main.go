package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk/v2"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/rtds"
)

type checkResult struct {
	name     string
	err      error
	optional bool
}

func main() {
	var (
		timeout    = flag.Duration("timeout", 12*time.Second, "per-check timeout")
		publicOnly = flag.Bool("public-only", false, "skip private checks")
		skipWS     = flag.Bool("skip-ws", false, "skip CLOB websocket checks")
		skipRTDS   = flag.Bool("skip-rtds", false, "skip RTDS checks")
		strict     = flag.Bool("strict", false, "fail on optional checks")
		tokenID    = flag.String("token", "", "token id for market data checks")
		marketID   = flag.String("market", "", "market/condition id for price history checks")
	)
	flag.Parse()

	client := polymarket.NewClient()
	results := make([]checkResult, 0, 32)

	ctx := context.Background()

	results = append(results, runCheck(ctx, *timeout, "clob.time", false, func(ctx context.Context) error {
		_, err := client.CLOB.Time(ctx)
		return err
	}))

	results = append(results, runCheck(ctx, *timeout, "clob.health", true, func(ctx context.Context) error {
		_, err := client.CLOB.Health(ctx)
		return err
	}))

	marketCondition, marketAlt, token := *marketID, "", *tokenID
	mCond, mAlt, tkn, err := pickMarketAndToken(ctx, client, *timeout)
	results = append(results, checkResult{name: "clob.markets", err: err, optional: false})
	if marketCondition == "" {
		marketCondition = mCond
	}
	if marketAlt == "" {
		marketAlt = mAlt
	}
	if token == "" {
		token = tkn
	}

	results = append(results, runCheck(ctx, *timeout, "clob.simplified_markets", true, func(ctx context.Context) error {
		_, err := client.CLOB.SimplifiedMarkets(ctx, &clobtypes.MarketsRequest{Limit: 1})
		return err
	}))

	results = append(results, runCheck(ctx, *timeout, "clob.sampling_markets", true, func(ctx context.Context) error {
		_, err := client.CLOB.SamplingMarkets(ctx, nil)
		return err
	}))

	results = append(results, runCheck(ctx, *timeout, "clob.sampling_simplified_markets", true, func(ctx context.Context) error {
		_, err := client.CLOB.SamplingSimplifiedMarkets(ctx, nil)
		return err
	}))

	if marketCondition != "" || marketAlt != "" {
		results = append(results, runCheck(ctx, *timeout, "clob.market", false, func(ctx context.Context) error {
			id := firstNonEmpty(marketCondition, marketAlt)
			_, err := client.CLOB.Market(ctx, id)
			if err != nil && marketAlt != "" && marketAlt != id {
				_, err = client.CLOB.Market(ctx, marketAlt)
			}
			return err
		}))
	} else {
		results = append(results, checkResult{name: "clob.market", err: fmt.Errorf("missing market id"), optional: true})
	}

	if token != "" {
		results = append(results, runCheck(ctx, *timeout, "clob.order_book", false, func(ctx context.Context) error {
			_, err := client.CLOB.OrderBook(ctx, &clobtypes.BookRequest{TokenID: token})
			return err
		}))
		results = append(results, runCheck(ctx, *timeout, "clob.midpoint", true, func(ctx context.Context) error {
			_, err := client.CLOB.Midpoint(ctx, &clobtypes.MidpointRequest{TokenID: token})
			return err
		}))
		results = append(results, runCheck(ctx, *timeout, "clob.price", true, func(ctx context.Context) error {
			_, err := client.CLOB.Price(ctx, &clobtypes.PriceRequest{TokenID: token, Side: "BUY"})
			return err
		}))
		results = append(results, runCheck(ctx, *timeout, "clob.spread", true, func(ctx context.Context) error {
			_, err := client.CLOB.Spread(ctx, &clobtypes.SpreadRequest{TokenID: token, Side: "BUY"})
			return err
		}))
		results = append(results, runCheck(ctx, *timeout, "clob.last_trade_price", true, func(ctx context.Context) error {
			_, err := client.CLOB.LastTradePrice(ctx, &clobtypes.LastTradePriceRequest{TokenID: token})
			return err
		}))
		results = append(results, runCheck(ctx, *timeout, "clob.tick_size", true, func(ctx context.Context) error {
			_, err := client.CLOB.TickSize(ctx, &clobtypes.TickSizeRequest{TokenID: token})
			return err
		}))
		results = append(results, runCheck(ctx, *timeout, "clob.neg_risk", true, func(ctx context.Context) error {
			_, err := client.CLOB.NegRisk(ctx, &clobtypes.NegRiskRequest{TokenID: token})
			return err
		}))
		results = append(results, runCheck(ctx, *timeout, "clob.fee_rate", true, func(ctx context.Context) error {
			_, err := client.CLOB.FeeRate(ctx, &clobtypes.FeeRateRequest{TokenID: token})
			return err
		}))
	} else {
		results = append(results, checkResult{name: "clob.token_dependent", err: fmt.Errorf("missing token id"), optional: true})
	}

	if marketCondition != "" || token != "" {
		results = append(results, runCheck(ctx, *timeout, "clob.prices_history", true, func(ctx context.Context) error {
			req := &clobtypes.PricesHistoryRequest{Interval: clobtypes.PriceHistoryInterval1d}
			if marketCondition != "" {
				req.Market = marketCondition
			} else {
				req.TokenID = token
			}
			_, err := client.CLOB.PricesHistory(ctx, req)
			return err
		}))
	}

	results = append(results, runCheck(ctx, *timeout, "clob.geoblock", true, func(ctx context.Context) error {
		_, err := client.CLOB.Geoblock(ctx)
		return err
	}))

	if !*publicOnly {
		pk := os.Getenv("POLYMARKET_PK")
		apiKey := os.Getenv("POLYMARKET_API_KEY")
		apiSecret := os.Getenv("POLYMARKET_API_SECRET")
		apiPassphrase := os.Getenv("POLYMARKET_API_PASSPHRASE")
		if pk == "" || apiKey == "" || apiSecret == "" || apiPassphrase == "" {
			results = append(results, checkResult{name: "private.auth", err: fmt.Errorf("missing credentials"), optional: true})
		} else {
			chainID := auth.PolygonChainID
			if raw := os.Getenv("POLYMARKET_CHAIN_ID"); raw != "" {
				if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
					chainID = parsed
				}
			}
			signer, err := auth.NewPrivateKeySigner(pk, chainID)
			results = append(results, checkResult{name: "private.signer", err: err, optional: false})
			if err == nil {
				creds := &auth.APIKey{Key: apiKey, Secret: apiSecret, Passphrase: apiPassphrase}
				authClient := client.CLOB.WithAuth(signer, creds)

				if raw := os.Getenv("POLYMARKET_SIGNATURE_TYPE"); raw != "" {
					if sig, err := strconv.Atoi(raw); err == nil {
						authClient = authClient.WithSignatureType(auth.SignatureType(sig))
					}
				}
				if raw := os.Getenv("POLYMARKET_AUTH_NONCE"); raw != "" {
					if nonce, err := strconv.ParseInt(raw, 10, 64); err == nil {
						authClient = authClient.WithAuthNonce(nonce)
					}
				}
				if raw := os.Getenv("POLYMARKET_FUNDER"); raw != "" {
					authClient = authClient.WithFunder(common.HexToAddress(raw))
				}

				results = append(results, runCheck(ctx, *timeout, "private.balance_allowance", false, func(ctx context.Context) error {
					_, err := authClient.BalanceAllowance(ctx, &clobtypes.BalanceAllowanceRequest{Asset: "USDC"})
					return err
				}))

				results = append(results, runCheck(ctx, *timeout, "private.orders", false, func(ctx context.Context) error {
					_, err := authClient.Orders(ctx, &clobtypes.OrdersRequest{Limit: 1})
					return err
				}))

				results = append(results, runCheck(ctx, *timeout, "private.trades", true, func(ctx context.Context) error {
					_, err := authClient.Trades(ctx, &clobtypes.TradesRequest{Limit: 1})
					return err
				}))

				results = append(results, runCheck(ctx, *timeout, "private.notifications", true, func(ctx context.Context) error {
					_, err := authClient.Notifications(ctx, nil)
					return err
				}))

				today := time.Now().UTC().Format("2006-01-02")
				results = append(results, runCheck(ctx, *timeout, "private.user_earnings", true, func(ctx context.Context) error {
					_, err := authClient.UserEarnings(ctx, &clobtypes.UserEarningsRequest{Date: today})
					return err
				}))

				results = append(results, runCheck(ctx, *timeout, "private.user_rewards_by_market", true, func(ctx context.Context) error {
					_, err := authClient.UserRewardsByMarket(ctx, &clobtypes.UserRewardsByMarketRequest{Date: today})
					return err
				}))

				results = append(results, runCheck(ctx, *timeout, "private.closed_only", true, func(ctx context.Context) error {
					_, err := authClient.ClosedOnlyStatus(ctx)
					return err
				}))
			}
		}
	}

	if !*skipWS {
		if token == "" {
			results = append(results, checkResult{name: "clob.ws", err: fmt.Errorf("missing token id"), optional: true})
		} else {
			wsClient := client.CLOBWS
			createdWS := false
			if wsClient == nil {
				url := os.Getenv("POLYMARKET_CLOB_WS_URL")
				if url == "" {
					url = ws.ProdBaseURL
				}
				var err error
				wsClient, err = ws.NewClient(url, nil, nil)
				if err != nil {
					results = append(results, checkResult{name: "clob.ws", err: err, optional: true})
					wsClient = nil
				} else {
					createdWS = true
				}
			}
			if wsClient == nil {
				results = append(results, checkResult{name: "clob.ws", err: fmt.Errorf("missing ws client"), optional: true})
			} else {
				closeWS := func() error { return wsClient.Close() }
				if !createdWS {
					closeWS = client.CLOBWS.Close
				}
				results = append(results, runCheck(ctx, *timeout, "clob.ws", true, func(ctx context.Context) error {
					_, err := wsClient.SubscribePrices(ctx, []string{token})
					if err != nil {
						return err
					}
					_ = wsClient.UnsubscribeMarketAssets(context.Background(), []string{token})
					return closeWS()
				}))
			}
		}
	}

	if !*skipRTDS {
		rtdsClient := client.RTDS
		createdRTDS := false
		if rtdsClient == nil {
			url := os.Getenv("POLYMARKET_RTDS_URL")
			var err error
			rtdsClient, err = rtds.NewClient(url)
			if err != nil {
				results = append(results, checkResult{name: "rtds", err: err, optional: true})
				rtdsClient = nil
			} else {
				createdRTDS = true
			}
		}
		if rtdsClient == nil {
			results = append(results, checkResult{name: "rtds", err: fmt.Errorf("missing rtds client"), optional: true})
		} else {
			closeRTDS := func() error { return rtdsClient.Close() }
			if !createdRTDS {
				closeRTDS = client.RTDS.Close
			}
			results = append(results, runCheck(ctx, *timeout, "rtds.crypto_prices", true, func(ctx context.Context) error {
				stream, err := rtdsClient.SubscribeCryptoPricesStream(ctx, []string{"btcusdt"})
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
				case <-stream.C:
				}
				_ = rtdsClient.UnsubscribeCryptoPrices(context.Background())
				return closeRTDS()
			}))
		}
	}

	printSummary(results)

	failed := 0
	for _, res := range results {
		if res.err == nil {
			continue
		}
		if res.optional && !*strict {
			continue
		}
		failed++
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func runCheck(ctx context.Context, timeout time.Duration, name string, optional bool, fn func(ctx context.Context) error) checkResult {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Printf("checking %s...", name)
	err := fn(ctx)
	if err != nil {
		log.Printf("check %s failed: %v", name, err)
	} else {
		log.Printf("check %s ok", name)
	}
	return checkResult{name: name, err: err, optional: optional}
}

func pickMarketAndToken(ctx context.Context, client *polymarket.Client, timeout time.Duration) (string, string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	active := true
	resp, err := client.CLOB.Markets(ctx, &clobtypes.MarketsRequest{Limit: 25, Active: &active})
	if err != nil {
		return "", "", "", err
	}
	if len(resp.Data) == 0 {
		return "", "", "", fmt.Errorf("no markets returned")
	}

	var conditionID string
	var altID string
	maxTokens := 50
	tokens := make([]string, 0, maxTokens)
	for _, market := range resp.Data {
		if conditionID == "" {
			if market.ConditionID != "" {
				conditionID = market.ConditionID
			} else if market.ID != "" {
				conditionID = market.ID
			}
		}
		if altID == "" && market.ID != "" {
			altID = market.ID
		}
		for _, token := range market.Tokens {
			if token.TokenID != "" {
				tokens = append(tokens, token.TokenID)
				if len(tokens) >= maxTokens {
					break
				}
			}
		}
		if len(tokens) >= maxTokens {
			break
		}
	}
	tokenID := findTokenWithOrderbook(ctx, client, timeout, tokens)
	return conditionID, altID, tokenID, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func printSummary(results []checkResult) {
	ok := 0
	failed := 0
	warn := 0
	for _, res := range results {
		if res.err == nil {
			ok++
			continue
		}
		if res.optional {
			warn++
		} else {
			failed++
		}
	}
	log.Printf("summary: ok=%d failed=%d warnings=%d", ok, failed, warn)
	for _, res := range results {
		if res.err == nil {
			continue
		}
		level := "ERROR"
		if res.optional {
			level = "WARN"
		}
		log.Printf("%s: %s -> %v", level, res.name, res.err)
	}
}

func findTokenWithOrderbook(ctx context.Context, client *polymarket.Client, timeout time.Duration, tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	maxChecks := 25
	if len(tokens) < maxChecks {
		maxChecks = len(tokens)
	}
	probeTimeout := minDuration(timeout, 2*time.Second)
	for i := 0; i < maxChecks; i++ {
		if ctx.Err() != nil {
			return ""
		}
		token := strings.TrimSpace(tokens[i])
		if token == "" {
			continue
		}
		probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
		_, err := client.CLOB.OrderBook(probeCtx, &clobtypes.BookRequest{TokenID: token})
		cancel()
		if err == nil {
			return token
		}
	}
	return ""
}

func minDuration(a, b time.Duration) time.Duration {
	if a <= 0 {
		return b
	}
	if a < b {
		return a
	}
	return b
}
