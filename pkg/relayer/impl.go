package relayer

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
)

const (
	// BaseURL is the default production endpoint for the Polymarket Relayer API.
	BaseURL = "https://relayer.polymarket.com"
)

type clientImpl struct {
	httpClient *transport.Client
}

// NewClient creates a new Relayer API client.
// If httpClient is nil, a default transport client targeting the production
// Relayer API base URL is created.
func NewClient(httpClient *transport.Client) Client {
	if httpClient == nil {
		httpClient = transport.NewClient(nil, BaseURL)
	}
	return &clientImpl{
		httpClient: httpClient,
	}
}

func (c *clientImpl) Submit(ctx context.Context, req *SubmitRequest) (*SubmitResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("submit request is required")
	}
	var resp SubmitResponse
	err := c.httpClient.Post(ctx, "/submit", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImpl) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	if id == "" {
		return nil, fmt.Errorf("transaction id is required")
	}
	q := url.Values{}
	q.Set("id", id)
	var resp Transaction
	err := c.httpClient.Get(ctx, "/transaction", q, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImpl) GetTransactions(ctx context.Context, req *GetTransactionsRequest) ([]Transaction, error) {
	q := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Offset > 0 {
			q.Set("offset", strconv.Itoa(req.Offset))
		}
	}
	var resp []Transaction
	err := c.httpClient.Get(ctx, "/transactions", q, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clientImpl) GetNonce(ctx context.Context, signer string) (*NonceResponse, error) {
	if signer == "" {
		return nil, fmt.Errorf("signer is required")
	}
	q := url.Values{}
	q.Set("signer", signer)
	var resp NonceResponse
	err := c.httpClient.Get(ctx, "/nonce", q, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImpl) GetRelayPayload(ctx context.Context, signer string) (*RelayPayloadResponse, error) {
	if signer == "" {
		return nil, fmt.Errorf("signer is required")
	}
	q := url.Values{}
	q.Set("signer", signer)
	var resp RelayPayloadResponse
	err := c.httpClient.Get(ctx, "/relay-payload", q, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImpl) GetDeployed(ctx context.Context, address string) (*DeployedResponse, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	q := url.Values{}
	q.Set("address", address)
	var resp DeployedResponse
	err := c.httpClient.Get(ctx, "/deployed", q, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImpl) GetAPIKeys(ctx context.Context) ([]APIKeyResponse, error) {
	var resp []APIKeyResponse
	err := c.httpClient.Get(ctx, "/relayer/api/keys", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
