package clob

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// CreateOrder builds and signs an order, then posts it to the CLOB.
// This is a higher-level helper that combines signing and posting.
func (c *clientImpl) CreateOrder(ctx context.Context, order *clobtypes.Order) (clobtypes.OrderResponse, error) {
	return c.CreateOrderWithOptions(ctx, order, nil)
}

func (c *clientImpl) CreateOrderWithOptions(ctx context.Context, order *clobtypes.Order, opts *clobtypes.OrderOptions) (clobtypes.OrderResponse, error) {
	signed, err := c.signOrder(order)
	if err != nil {
		return clobtypes.OrderResponse{}, err
	}
	if opts != nil {
		signed.OrderType = opts.OrderType
		signed.PostOnly = opts.PostOnly
		signed.DeferExec = opts.DeferExec
	}
	return c.PostOrder(ctx, signed)
}

func (c *clientImpl) CreateOrderFromSignable(ctx context.Context, order *clobtypes.SignableOrder) (clobtypes.OrderResponse, error) {
	if order == nil || order.Order == nil {
		return clobtypes.OrderResponse{}, fmt.Errorf("order is required")
	}
	opts := &clobtypes.OrderOptions{
		OrderType: order.OrderType,
		PostOnly:  order.PostOnly,
	}
	return c.CreateOrderWithOptions(ctx, order.Order, opts)
}

func (c *clientImpl) signOrder(order *clobtypes.Order) (*clobtypes.SignedOrder, error) {
	return signOrderWithCreds(c.signer, c.apiKey, order, &c.signatureType, c.funder, c.saltGenerator)
}

// SignOrder builds an EIP-712 signature for the given order without posting it.
func SignOrder(signer auth.Signer, apiKey *auth.APIKey, order *clobtypes.Order) (*clobtypes.SignedOrder, error) {
	return signOrderWithCreds(signer, apiKey, order, nil, nil, nil)
}

func signOrderWithCreds(signer auth.Signer, apiKey *auth.APIKey, order *clobtypes.Order, sigType *auth.SignatureType, funder *types.Address, saltGen SaltGenerator) (*clobtypes.SignedOrder, error) {
	if signer == nil {
		return nil, auth.ErrMissingSigner
	}
	if apiKey == nil {
		return nil, auth.ErrMissingCreds
	}
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}

	sigTypeVal := int(auth.SignatureEOA)
	if order.SignatureType != nil {
		sigTypeVal = *order.SignatureType
	} else if sigType != nil {
		sigTypeVal = int(*sigType)
		val := sigTypeVal
		order.SignatureType = &val
	}

	if order.Maker == (types.Address{}) {
		if funder != nil {
			if sigTypeVal == int(auth.SignatureEOA) {
				return nil, fmt.Errorf("funder requires non-EOA signature type")
			}
			if *funder == (types.Address{}) {
				return nil, fmt.Errorf("funder cannot be zero address")
			}
			order.Maker = *funder
		} else {
			maker, err := deriveMakerFromSignature(signer, sigTypeVal)
			if err != nil {
				return nil, err
			}
			order.Maker = maker
		}
	}

	domain := &apitypes.TypedDataDomain{
		Name:              "Polymarket CTF Exchange",
		Version:           "1",
		ChainId:           (*math.HexOrDecimal256)(signer.ChainID()),
		VerifyingContract: "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E", // Exchange Contract Address (Mainnet)
	}

	typesDef := apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"Order": {
			{Name: "salt", Type: "uint256"},
			{Name: "maker", Type: "address"},
			{Name: "signer", Type: "address"},
			{Name: "taker", Type: "address"},
			{Name: "tokenId", Type: "uint256"},
			{Name: "makerAmount", Type: "uint256"},
			{Name: "takerAmount", Type: "uint256"},
			{Name: "expiration", Type: "uint256"},
			{Name: "nonce", Type: "uint256"},
			{Name: "feeRateBps", Type: "uint256"},
			{Name: "side", Type: "uint8"},
			{Name: "signatureType", Type: "uint8"},
		},
	}

	sideInt := 0
	if strings.ToUpper(order.Side) == "SELL" {
		sideInt = 1
	}

	if order.Salt.Int == nil || order.Salt.Int.Sign() == 0 {
		var salt *big.Int
		var err error
		if saltGen != nil {
			salt, err = saltGen()
		} else {
			salt, err = generateSalt()
		}
		if err != nil {
			return nil, err
		}
		order.Salt = types.U256{Int: salt}
	}

	expiration := big.NewInt(0)
	if order.Expiration.Int != nil {
		expiration = order.Expiration.Int
	}

	message := apitypes.TypedDataMessage{
		"salt":          (*math.HexOrDecimal256)(order.Salt.Int),
		"maker":         order.Maker.String(),
		"signer":        signer.Address().String(),
		"taker":         order.Taker.String(),
		"tokenId":       (*math.HexOrDecimal256)(order.TokenID.Int),
		"makerAmount":   (*math.HexOrDecimal256)(order.MakerAmount.BigInt()),
		"takerAmount":   (*math.HexOrDecimal256)(order.TakerAmount.BigInt()),
		"expiration":    (*math.HexOrDecimal256)(expiration),
		"nonce":         (*math.HexOrDecimal256)(order.Nonce.Int),
		"feeRateBps":    (*math.HexOrDecimal256)(order.FeeRateBps.BigInt()),
		"side":          (*math.HexOrDecimal256)(big.NewInt(int64(sideInt))),
		"signatureType": (*math.HexOrDecimal256)(big.NewInt(int64(sigTypeVal))),
	}

	sig, err := signer.SignTypedData(domain, typesDef, message, "Order")
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	owner := apiKey.Key
	if owner == "" {
		owner = signer.Address().String()
	}

	return &clobtypes.SignedOrder{
		Order:     *order,
		Signature: hexutil.Encode(sig),
		Owner:     owner,
	}, nil
}

func (c *clientImpl) PostOrder(ctx context.Context, req *clobtypes.SignedOrder) (clobtypes.OrderResponse, error) {
	var resp clobtypes.OrderResponse
	payload, err := buildOrderPayload(req)
	if err != nil {
		return resp, err
	}
	err = c.httpClient.Post(ctx, "/order", payload, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) PostOrders(ctx context.Context, req *clobtypes.SignedOrders) (clobtypes.PostOrdersResponse, error) {
	var resp clobtypes.PostOrdersResponse
	payload, err := buildOrdersPayload(req)
	if err != nil {
		return resp, err
	}
	err = c.httpClient.Post(ctx, "/orders", payload, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) CancelOrder(ctx context.Context, req *clobtypes.CancelOrderRequest) (clobtypes.CancelResponse, error) {
	var resp clobtypes.CancelResponse
	var body interface{}
	if req != nil {
		if req.OrderID != "" {
			body = map[string]string{"orderId": req.OrderID}
		}
	}
	err := c.httpClient.Delete(ctx, "/order", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) CancelOrders(ctx context.Context, req *clobtypes.CancelOrdersRequest) (clobtypes.CancelResponse, error) {
	var resp clobtypes.CancelResponse
	var body interface{}
	if req != nil {
		ids := req.OrderIDs
		if len(ids) > 0 {
			body = ids
		}
	}
	err := c.httpClient.Delete(ctx, "/orders", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) CancelAll(ctx context.Context) (clobtypes.CancelAllResponse, error) {
	var resp clobtypes.CancelAllResponse
	err := c.httpClient.Delete(ctx, "/cancel-all", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) CancelMarketOrders(ctx context.Context, req *clobtypes.CancelMarketOrdersRequest) (clobtypes.CancelMarketOrdersResponse, error) {
	var resp clobtypes.CancelMarketOrdersResponse
	var body interface{}
	if req != nil {
		market := req.Market
		payload := map[string]string{}
		if market != "" {
			payload["market"] = market
		}
		if req.AssetID != "" {
			payload["asset_id"] = req.AssetID
		}
		if len(payload) > 0 {
			body = payload
		}
	}
	err := c.httpClient.Delete(ctx, "/cancel-market-orders", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Order(ctx context.Context, id string) (clobtypes.OrderResponse, error) {
	var resp clobtypes.OrderResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/data/order/%s", id), nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Orders(ctx context.Context, req *clobtypes.OrdersRequest) (clobtypes.OrdersResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.ID != "" {
			q.Set("id", req.ID)
		}
		if req.Market != "" {
			q.Set("market", req.Market)
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		nextCursor := req.NextCursor
		if nextCursor == "" {
			nextCursor = req.Cursor
		}
		if nextCursor != "" {
			q.Set("next_cursor", nextCursor)
		}
	}
	var resp clobtypes.OrdersResponse
	err := c.httpClient.Get(ctx, "/data/orders", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) Trades(ctx context.Context, req *clobtypes.TradesRequest) (clobtypes.TradesResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.ID != "" {
			q.Set("id", req.ID)
		}
		if req.Taker != "" {
			q.Set("taker", req.Taker)
		}
		if req.Maker != "" {
			q.Set("maker", req.Maker)
		}
		if req.Market != "" {
			q.Set("market", req.Market)
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
		if req.Before > 0 {
			q.Set("before", strconv.FormatInt(req.Before, 10))
		}
		if req.After > 0 {
			q.Set("after", strconv.FormatInt(req.After, 10))
		}
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		nextCursor := req.NextCursor
		if nextCursor == "" {
			nextCursor = req.Cursor
		}
		if nextCursor != "" {
			q.Set("next_cursor", nextCursor)
		}
	}
	var resp clobtypes.TradesResponse
	err := c.httpClient.Get(ctx, "/data/trades", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) OrdersAll(ctx context.Context, req *clobtypes.OrdersRequest) ([]clobtypes.OrderResponse, error) {
	var results []clobtypes.OrderResponse
	cursor := clobtypes.InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = clobtypes.InitialCursor
	}

	for cursor != clobtypes.EndCursor {
		nextReq := clobtypes.OrdersRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.NextCursor = cursor

		resp, err := c.Orders(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Data...)

		if resp.NextCursor == "" || resp.NextCursor == cursor {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}

func (c *clientImpl) TradesAll(ctx context.Context, req *clobtypes.TradesRequest) ([]clobtypes.Trade, error) {
	var results []clobtypes.Trade
	cursor := clobtypes.InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = clobtypes.InitialCursor
	}

	for cursor != clobtypes.EndCursor {
		nextReq := clobtypes.TradesRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.NextCursor = cursor

		resp, err := c.Trades(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Data...)

		if resp.NextCursor == "" || resp.NextCursor == cursor {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}

func (c *clientImpl) BuilderTradesAll(ctx context.Context, req *clobtypes.BuilderTradesRequest) ([]clobtypes.Trade, error) {
	var results []clobtypes.Trade
	cursor := clobtypes.InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = clobtypes.InitialCursor
	}

	for cursor != clobtypes.EndCursor {
		nextReq := clobtypes.BuilderTradesRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.NextCursor = cursor

		resp, err := c.BuilderTrades(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Data...)

		if resp.NextCursor == "" || resp.NextCursor == cursor {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}

func (c *clientImpl) OrderScoring(ctx context.Context, req *clobtypes.OrderScoringRequest) (clobtypes.OrderScoringResponse, error) {
	q := url.Values{}
	if req != nil && req.ID != "" {
		q.Set("order_id", req.ID)
	}
	var resp clobtypes.OrderScoringResponse
	err := c.httpClient.Get(ctx, "/order-scoring", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) OrdersScoring(ctx context.Context, req *clobtypes.OrdersScoringRequest) (clobtypes.OrdersScoringResponse, error) {
	var resp clobtypes.OrdersScoringResponse
	var body []string
	if req != nil {
		body = req.IDs
	}
	err := c.httpClient.Post(ctx, "/orders-scoring", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) BuilderTrades(ctx context.Context, req *clobtypes.BuilderTradesRequest) (clobtypes.BuilderTradesResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.ID != "" {
			q.Set("id", req.ID)
		}
		if req.Maker != "" {
			q.Set("maker", req.Maker)
		}
		if req.Market != "" {
			q.Set("market", req.Market)
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
		if req.Before > 0 {
			q.Set("before", strconv.FormatInt(req.Before, 10))
		}
		if req.After > 0 {
			q.Set("after", strconv.FormatInt(req.After, 10))
		}
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		nextCursor := req.NextCursor
		if nextCursor == "" {
			nextCursor = req.Cursor
		}
		if nextCursor != "" {
			q.Set("next_cursor", nextCursor)
		}
	}
	var resp clobtypes.BuilderTradesResponse
	err := c.httpClient.Get(ctx, "/builder/trades", q, &resp)
	return resp, mapError(err)
}
