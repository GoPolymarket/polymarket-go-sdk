// Package relayer provides a client for the Polymarket Relayer API.
// The Relayer API enables gasless transactions by having a relayer submit
// transactions on behalf of users. All endpoints support L2 HMAC authentication
// via the transport client's SetAuth mechanism.
package relayer

import "context"

// Client defines the interface for the Polymarket Relayer API.
type Client interface {
	// Submit submits a gasless transaction to the relayer.
	Submit(ctx context.Context, req *SubmitRequest) (*SubmitResponse, error)

	// GetTransaction retrieves a transaction by its ID.
	GetTransaction(ctx context.Context, id string) (*Transaction, error)

	// GetTransactions retrieves recent transactions for the authenticated user.
	GetTransactions(ctx context.Context, req *GetTransactionsRequest) ([]Transaction, error)

	// GetNonce retrieves the current Proxy/Safe nonce for a signer address.
	GetNonce(ctx context.Context, signer string) (*NonceResponse, error)

	// GetRelayPayload retrieves the relayer address and nonce for a signer address.
	GetRelayPayload(ctx context.Context, signer string) (*RelayPayloadResponse, error)

	// GetDeployed checks whether a wallet is deployed at the given address.
	GetDeployed(ctx context.Context, address string) (*DeployedResponse, error)

	// GetAPIKeys lists relayer API keys for the authenticated user.
	GetAPIKeys(ctx context.Context) ([]APIKeyResponse, error)
}
