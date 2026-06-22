package bridge

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Client defines the Bridge API interface.
type Client interface {
	Deposit(ctx context.Context, amount *big.Int, asset common.Address) (*types.Transaction, error)
	Withdraw(ctx context.Context, amount *big.Int, asset common.Address) (*types.Transaction, error)
	WithdrawTo(ctx context.Context, req *WithdrawRequest) (*types.Transaction, error)
	SupportedAssets(ctx context.Context) ([]common.Address, error)

	DepositAddress(ctx context.Context, req *DepositRequest) (DepositResponse, error)
	SupportedAssetsInfo(ctx context.Context) (SupportedAssetsResponse, error)
	Status(ctx context.Context, req *StatusRequest) (StatusResponse, error)

	// Quote returns a bridge quote for a cross-chain transfer.
	Quote(ctx context.Context, req *BridgeQuoteRequest) (BridgeQuoteResponse, error)
	// WithdrawAddress creates bridge withdrawal addresses to send funds from a Polymarket wallet.
	WithdrawAddress(ctx context.Context, req *WithdrawAddressRequest) (WithdrawAddressResponse, error)
}
