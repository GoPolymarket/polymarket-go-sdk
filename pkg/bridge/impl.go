package bridge

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/errors"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/transport"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	BaseURL = "https://bridge.polymarket.com"
)

type clientImpl struct {
	httpClient *transport.Client
	backend    bind.ContractBackend
	txOpts     *bind.TransactOpts
}

func NewClient(httpClient *transport.Client) Client {
	if httpClient == nil {
		httpClient = transport.NewClient(nil, BaseURL)
	}
	return &clientImpl{
		httpClient: httpClient,
	}
}

// NewClientWithBackend creates a Bridge client with an on-chain backend for EVM transfers.
func NewClientWithBackend(httpClient *transport.Client, backend bind.ContractBackend, txOpts *bind.TransactOpts) (Client, error) {
	if backend == nil {
		return nil, ErrMissingBackend
	}
	if txOpts == nil {
		return nil, ErrMissingTransactor
	}
	if httpClient == nil {
		httpClient = transport.NewClient(nil, BaseURL)
	}
	return &clientImpl{
		httpClient: httpClient,
		backend:    backend,
		txOpts:     txOpts,
	}, nil
}

func (c *clientImpl) Deposit(ctx context.Context, amount *big.Int, asset common.Address) (*types.Transaction, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if asset == (common.Address{}) {
		return nil, fmt.Errorf("asset is required")
	}
	if c.backend == nil {
		return nil, ErrMissingBackend
	}
	if c.txOpts == nil {
		return nil, ErrMissingTransactor
	}
	if c.txOpts.From == (common.Address{}) {
		return nil, ErrMissingFromAddress
	}

	resp, err := c.DepositAddress(ctx, &DepositRequest{Address: c.txOpts.From.Hex()})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.Address.EVM) == "" {
		return nil, ErrMissingDepositAddress
	}
	if !common.IsHexAddress(resp.Address.EVM) {
		return nil, fmt.Errorf("invalid evm deposit address: %s", resp.Address.EVM)
	}
	to := common.HexToAddress(resp.Address.EVM)

	return c.transferERC20(ctx, asset, to, amount)
}

func (c *clientImpl) Withdraw(ctx context.Context, amount *big.Int, asset common.Address) (*types.Transaction, error) {
	return nil, ErrWithdrawUnsupported
}

func (c *clientImpl) WithdrawTo(ctx context.Context, req *WithdrawRequest) (*types.Transaction, error) {
	if req == nil {
		return nil, ErrMissingWithdrawRequest
	}
	if req.Amount == nil || req.Amount.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if req.Asset == (common.Address{}) {
		return nil, fmt.Errorf("asset is required")
	}
	if req.To == (common.Address{}) {
		return nil, ErrMissingWithdrawAddress
	}
	if c.backend == nil {
		return nil, ErrMissingBackend
	}
	if c.txOpts == nil {
		return nil, ErrMissingTransactor
	}
	if c.txOpts.From == (common.Address{}) {
		return nil, ErrMissingFromAddress
	}

	return c.transferERC20(ctx, req.Asset, req.To, req.Amount)
}

func (c *clientImpl) SupportedAssets(ctx context.Context) ([]common.Address, error) {
	info, err := c.SupportedAssetsInfo(ctx)
	if err != nil {
		return nil, err
	}
	seen := map[common.Address]struct{}{}
	out := make([]common.Address, 0, len(info.SupportedAssets))
	for _, asset := range info.SupportedAssets {
		addrStr := strings.TrimSpace(asset.Token.Address)
		if addrStr == "" || !common.IsHexAddress(addrStr) {
			continue
		}
		addr := common.HexToAddress(addrStr)
		if addr == (common.Address{}) {
			continue
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	return out, nil
}

func (c *clientImpl) DepositAddress(ctx context.Context, req *DepositRequest) (DepositResponse, error) {
	if req == nil || req.Address == "" {
		return DepositResponse{}, fmt.Errorf("address is required")
	}
	var resp DepositResponse
	err := c.httpClient.Post(ctx, "/deposit", req, &resp)
	return resp, err
}

func (c *clientImpl) SupportedAssetsInfo(ctx context.Context) (SupportedAssetsResponse, error) {
	var resp SupportedAssetsResponse
	err := c.httpClient.Get(ctx, "/supported-assets", nil, &resp)
	return resp, err
}

func (c *clientImpl) Status(ctx context.Context, req *StatusRequest) (StatusResponse, error) {
	if req == nil || req.Address == "" {
		return StatusResponse{}, fmt.Errorf("address is required")
	}
	var resp StatusResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/status/%s", req.Address), nil, &resp)
	return resp, err
}

// Use unified error definitions from pkg/errors
var (
	ErrMissingBackend         = sdkerrors.ErrMissingBackend
	ErrMissingTransactor      = sdkerrors.ErrMissingTransactor
	ErrMissingFromAddress     = sdkerrors.ErrMissingFromAddress
	ErrMissingDepositAddress  = sdkerrors.ErrMissingDepositAddress
	ErrWithdrawUnsupported    = sdkerrors.ErrWithdrawUnsupported
	ErrMissingWithdrawRequest = sdkerrors.ErrMissingWithdrawRequest
	ErrMissingWithdrawAddress = sdkerrors.ErrMissingWithdrawAddress
)

const erc20ABI = `[{
    "inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],
    "name":"transfer",
    "outputs":[{"internalType":"bool","name":"","type":"bool"}],
    "stateMutability":"nonpayable",
    "type":"function"
}]`

var (
	erc20ABIOnce  sync.Once
	erc20ABICache abi.ABI
	erc20ABIError error
)

func parseERC20ABI() (abi.ABI, error) {
	erc20ABIOnce.Do(func() {
		erc20ABICache, erc20ABIError = abi.JSON(strings.NewReader(erc20ABI))
	})
	return erc20ABICache, erc20ABIError
}

func (c *clientImpl) transferERC20(ctx context.Context, token common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	contractABI, err := parseERC20ABI()
	if err != nil {
		return nil, fmt.Errorf("parse erc20 abi: %w", err)
	}
	contract := bind.NewBoundContract(token, contractABI, c.backend, c.backend, c.backend)
	opts := *c.txOpts
	opts.Context = ctx
	tx, err := contract.Transact(&opts, "transfer", to, amount)
	if err != nil {
		return nil, fmt.Errorf("erc20 transfer: %w", err)
	}
	return tx, nil
}
