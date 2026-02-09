package clob

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func (c *clientImpl) BalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceRequest) (clobtypes.BalanceAllowanceResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Asset != "" {
			q.Set("asset", req.Asset)
		}
		if req.AssetType != "" {
			q.Set("asset_type", string(req.AssetType))
		}
		if req.TokenID != "" {
			q.Set("token_id", req.TokenID)
		}
		sigType := req.SignatureType
		if sigType == nil {
			val := int(c.signatureType)
			sigType = &val
		}
		if sigType != nil {
			q.Set("signature_type", strconv.Itoa(*sigType))
		}
	}
	var resp clobtypes.BalanceAllowanceResponse
	err := c.httpClient.Get(ctx, "/balance-allowance", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) UpdateBalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceUpdateRequest) (clobtypes.BalanceAllowanceResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Asset != "" {
			q.Set("asset", req.Asset)
		}
		if req.AssetType != "" {
			q.Set("asset_type", string(req.AssetType))
		}
		if req.TokenID != "" {
			q.Set("token_id", req.TokenID)
		}
		sigType := req.SignatureType
		if sigType == nil {
			val := int(c.signatureType)
			sigType = &val
		}
		if sigType != nil {
			q.Set("signature_type", strconv.Itoa(*sigType))
		}
		if req.Amount != "" {
			q.Set("amount", req.Amount)
		}
	}
	var resp clobtypes.BalanceAllowanceResponse
	err := c.httpClient.Call(ctx, "GET", "/balance-allowance/update", q, nil, &resp, nil)
	return resp, mapError(err)
}

func (c *clientImpl) Notifications(ctx context.Context, req *clobtypes.NotificationsRequest) (clobtypes.NotificationsResponse, error) {
	q := url.Values{}
	if req != nil && req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	var resp clobtypes.NotificationsResponse
	err := c.httpClient.Get(ctx, "/notifications", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) DropNotifications(ctx context.Context, req *clobtypes.DropNotificationsRequest) (clobtypes.DropNotificationsResponse, error) {
	q := url.Values{}
	if req != nil {
		ids := req.IDs
		if len(ids) > 0 {
			q.Set("id", strings.Join(ids, ","))
		}
	}
	var resp clobtypes.DropNotificationsResponse
	var err error
	if len(q) > 0 {
		err = c.httpClient.Call(ctx, "DELETE", "/notifications", q, nil, &resp, nil)
	} else {
		err = c.httpClient.Delete(ctx, "/notifications", nil, &resp)
	}
	return resp, mapError(err)
}

func (c *clientImpl) UserEarnings(ctx context.Context, req *clobtypes.UserEarningsRequest) (clobtypes.UserEarningsResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Date != "" {
			q.Set("date", req.Date)
		}
		sigType := req.SignatureType
		if sigType == nil {
			val := int(c.signatureType)
			sigType = &val
		}
		if sigType != nil {
			q.Set("signature_type", strconv.Itoa(*sigType))
		}
		if req.NextCursor != "" {
			q.Set("next_cursor", req.NextCursor)
		}
		if req.Asset != "" {
			q.Set("asset", req.Asset)
		}
	}
	var resp clobtypes.UserEarningsResponse
	err := c.httpClient.Get(ctx, "/rewards/user", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) UserTotalEarnings(ctx context.Context, req *clobtypes.UserTotalEarningsRequest) (clobtypes.UserTotalEarningsResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Date != "" {
			q.Set("date", req.Date)
		}
		sigType := req.SignatureType
		if sigType == nil {
			val := int(c.signatureType)
			sigType = &val
		}
		if sigType != nil {
			q.Set("signature_type", strconv.Itoa(*sigType))
		}
		if req.Asset != "" {
			q.Set("asset", req.Asset)
		}
	}
	var resp clobtypes.UserTotalEarningsResponse
	err := c.httpClient.Get(ctx, "/rewards/user/total", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) UserRewardPercentages(ctx context.Context, req *clobtypes.UserRewardPercentagesRequest) (clobtypes.UserRewardPercentagesResponse, error) {
	var resp clobtypes.UserRewardPercentagesResponse
	err := c.httpClient.Get(ctx, "/rewards/user/percentages", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) RewardsMarketsCurrent(ctx context.Context, req *clobtypes.RewardsMarketsRequest) (clobtypes.RewardsMarketsResponse, error) {
	q := url.Values{}
	if req != nil && req.NextCursor != "" {
		q.Set("next_cursor", req.NextCursor)
	}
	var resp clobtypes.RewardsMarketsResponse
	err := c.httpClient.Get(ctx, "/rewards/markets/current", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) RewardsMarkets(ctx context.Context, req *clobtypes.RewardsMarketRequest) (clobtypes.RewardsMarketResponse, error) {
	path := ""
	q := url.Values{}
	if req != nil {
		path = req.MarketID
		if req.NextCursor != "" {
			q.Set("next_cursor", req.NextCursor)
		}
	}
	if path == "" {
		return clobtypes.RewardsMarketResponse{}, fmt.Errorf("market_id is required")
	}
	var resp clobtypes.RewardsMarketResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/rewards/markets/%s", path), q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) UserRewardsByMarket(ctx context.Context, req *clobtypes.UserRewardsByMarketRequest) (clobtypes.UserRewardsByMarketResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Date != "" {
			q.Set("date", req.Date)
		}
		if req.OrderBy != "" {
			q.Set("order_by", req.OrderBy)
		}
		if req.Position != "" {
			q.Set("position", req.Position)
		}
		q.Set("no_competition", strconv.FormatBool(req.NoCompetition))
		sigType := req.SignatureType
		if sigType == nil {
			val := int(c.signatureType)
			sigType = &val
		}
		if sigType != nil {
			q.Set("signature_type", strconv.Itoa(*sigType))
		}
		if req.NextCursor != "" {
			q.Set("next_cursor", req.NextCursor)
		}
	}
	var resp clobtypes.UserRewardsByMarketResponse
	err := c.httpClient.Get(ctx, "/rewards/user/by-market", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) CreateAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	nonce := int64(0)
	if c.authNonce != nil {
		nonce = *c.authNonce
	}
	return c.CreateAPIKeyWithNonce(ctx, nonce)
}

func (c *clientImpl) CreateAPIKeyWithNonce(ctx context.Context, nonce int64) (clobtypes.APIKeyResponse, error) {
	if c.signer == nil {
		return clobtypes.APIKeyResponse{}, auth.ErrMissingSigner
	}

	headersRaw, err := auth.BuildL1Headers(c.signer, 0, nonce)
	if err != nil {
		return clobtypes.APIKeyResponse{}, err
	}

	headers := map[string]string{
		auth.HeaderPolyAddress:   headersRaw.Get(auth.HeaderPolyAddress),
		auth.HeaderPolyTimestamp: headersRaw.Get(auth.HeaderPolyTimestamp),
		auth.HeaderPolyNonce:     headersRaw.Get(auth.HeaderPolyNonce),
		auth.HeaderPolySignature: headersRaw.Get(auth.HeaderPolySignature),
	}

	var resp clobtypes.APIKeyResponse
	// Note: We use CallWithHeaders to inject L1 headers.
	// clobtypes.CreateAPIKey uses POST /auth/api-key
	err = c.httpClient.CallWithHeaders(ctx, "POST", "/auth/api-key", nil, nil, &resp, headers)
	return resp, mapError(err)
}

func (c *clientImpl) ListAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error) {
	var resp clobtypes.APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/api-keys", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) DeleteAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	q := url.Values{}
	if id != "" {
		q.Set("api_key", id)
	}
	if len(q) > 0 {
		err := c.httpClient.Call(ctx, "DELETE", "/auth/api-key", q, nil, &resp, nil)
		return resp, mapError(err)
	}
	err := c.httpClient.Delete(ctx, "/auth/api-key", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) DeriveAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	nonce := int64(0)
	if c.authNonce != nil {
		nonce = *c.authNonce
	}
	return c.DeriveAPIKeyWithNonce(ctx, nonce)
}

func (c *clientImpl) DeriveAPIKeyWithNonce(ctx context.Context, nonce int64) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	headersRaw, err := auth.BuildL1Headers(c.signer, 0, nonce)
	if err != nil {
		return clobtypes.APIKeyResponse{}, err
	}
	headers := map[string]string{
		auth.HeaderPolyAddress:   headersRaw.Get(auth.HeaderPolyAddress),
		auth.HeaderPolyTimestamp: headersRaw.Get(auth.HeaderPolyTimestamp),
		auth.HeaderPolyNonce:     headersRaw.Get(auth.HeaderPolyNonce),
		auth.HeaderPolySignature: headersRaw.Get(auth.HeaderPolySignature),
	}
	err = c.httpClient.CallWithHeaders(ctx, "GET", "/auth/derive-api-key", nil, nil, &resp, headers)
	return resp, mapError(err)
}

func (c *clientImpl) CreateOrDeriveAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	nonce := int64(0)
	if c.authNonce != nil {
		nonce = *c.authNonce
	}
	return c.CreateOrDeriveAPIKeyWithNonce(ctx, nonce)
}

func (c *clientImpl) CreateOrDeriveAPIKeyWithNonce(ctx context.Context, nonce int64) (clobtypes.APIKeyResponse, error) {
	resp, err := c.CreateAPIKeyWithNonce(ctx, nonce)
	if err == nil {
		return resp, nil
	}
	return c.DeriveAPIKeyWithNonce(ctx, nonce)
}

func (c *clientImpl) ClosedOnlyStatus(ctx context.Context) (clobtypes.ClosedOnlyResponse, error) {
	var resp clobtypes.ClosedOnlyResponse
	err := c.httpClient.Get(ctx, "/auth/ban-status/closed-only", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) CreateReadonlyAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	err := c.httpClient.Post(ctx, "/auth/readonly-api-key", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) ListReadonlyAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error) {
	var resp clobtypes.APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/readonly-api-keys", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) DeleteReadonlyAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	body := map[string]string{"key": id}
	err := c.httpClient.Delete(ctx, "/auth/readonly-api-key", body, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) ValidateReadonlyAPIKey(ctx context.Context, req *clobtypes.ValidateReadonlyAPIKeyRequest) (clobtypes.ValidateReadonlyAPIKeyResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Address != "" {
			q.Set("address", req.Address)
		}
		if req.APIKey != "" {
			q.Set("key", req.APIKey)
		}
	}
	var resp clobtypes.ValidateReadonlyAPIKeyResponse
	err := c.httpClient.Get(ctx, "/auth/validate-readonly-api-key", q, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) CreateBuilderAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	err := c.httpClient.Post(ctx, "/auth/builder-api-key", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) ListBuilderAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error) {
	var resp clobtypes.APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/builder-api-key", nil, &resp)
	return resp, mapError(err)
}

func (c *clientImpl) RevokeBuilderAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error) {
	// Endpoint returns empty body; ignore response.
	err := c.httpClient.Call(ctx, "DELETE", "/auth/builder-api-key", nil, nil, nil, nil)
	return clobtypes.APIKeyResponse{}, mapError(err)
}
