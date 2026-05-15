package ctf

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	conditionalTokensABI = `[{"inputs":[{"internalType":"address","name":"oracle","type":"address"},{"internalType":"bytes32","name":"questionId","type":"bytes32"},{"internalType":"uint256","name":"outcomeSlotCount","type":"uint256"}],"name":"prepareCondition","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"collateralToken","type":"address"},{"internalType":"bytes32","name":"parentCollectionId","type":"bytes32"},{"internalType":"bytes32","name":"conditionId","type":"bytes32"},{"internalType":"uint256[]","name":"partition","type":"uint256[]"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"splitPosition","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"collateralToken","type":"address"},{"internalType":"bytes32","name":"parentCollectionId","type":"bytes32"},{"internalType":"bytes32","name":"conditionId","type":"bytes32"},{"internalType":"uint256[]","name":"partition","type":"uint256[]"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"mergePositions","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"collateralToken","type":"address"},{"internalType":"bytes32","name":"parentCollectionId","type":"bytes32"},{"internalType":"bytes32","name":"conditionId","type":"bytes32"},{"internalType":"uint256[]","name":"indexSets","type":"uint256[]"}],"name":"redeemPositions","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	negRiskAdapterABI    = `[{"inputs":[{"internalType":"bytes32","name":"conditionId","type":"bytes32"},{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"name":"redeemPositions","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
)

// Use unified error definitions from pkg/errors
var (
	ErrMissingRequest    = sdkerrors.ErrMissingRequest
	ErrMissingU256Value  = sdkerrors.ErrMissingU256Value
	ErrMissingBackend    = sdkerrors.ErrMissingBackend
	ErrMissingTransactor = sdkerrors.ErrMissingTransactor
	ErrNegRiskAdapter    = sdkerrors.ErrNegRiskAdapter
	ErrConfigNotFound    = sdkerrors.ErrConfigNotFound
)

type clientImpl struct {
	backend           Backend
	txOpts            *bind.TransactOpts
	conditionalTokens *bind.BoundContract
	negRiskAdapter    *bind.BoundContract
}

// NewClient creates a lightweight CTF client for ID calculations.
// Transaction methods require a backend and transactor.
func NewClient() Client {
	return &clientImpl{}
}

// NewClientWithBackend creates a CTF client with a chain backend for transactions.
func NewClientWithBackend(backend Backend, txOpts *bind.TransactOpts, chainID int64) (Client, error) {
	return newClientWithConfig(backend, txOpts, chainID, false)
}

// NewClientWithNegRisk creates a CTF client with NegRisk adapter support.
func NewClientWithNegRisk(backend Backend, txOpts *bind.TransactOpts, chainID int64) (Client, error) {
	return newClientWithConfig(backend, txOpts, chainID, true)
}

func newClientWithConfig(backend Backend, txOpts *bind.TransactOpts, chainID int64, negRisk bool) (Client, error) {
	if backend == nil {
		return nil, ErrMissingBackend
	}
	cfg, ok := resolveConfig(chainID, negRisk)
	if !ok {
		return nil, ErrConfigNotFound
	}
	contractABI, err := abi.JSON(strings.NewReader(conditionalTokensABI))
	if err != nil {
		return nil, fmt.Errorf("parse conditional tokens ABI: %w", err)
	}
	contract := bind.NewBoundContract(cfg.ConditionalTokens, contractABI, backend, backend, backend)

	var neg *bind.BoundContract
	if cfg.NegRiskAdapter != nil {
		negABI, err := abi.JSON(strings.NewReader(negRiskAdapterABI))
		if err != nil {
			return nil, fmt.Errorf("parse neg risk ABI: %w", err)
		}
		neg = bind.NewBoundContract(*cfg.NegRiskAdapter, negABI, backend, backend, backend)
	}

	return &clientImpl{
		backend:           backend,
		txOpts:            txOpts,
		conditionalTokens: contract,
		negRiskAdapter:    neg,
	}, nil
}

func (c *clientImpl) PrepareCondition(ctx context.Context, req *PrepareConditionRequest) (PrepareConditionResponse, error) {
	if req == nil {
		return PrepareConditionResponse{}, ErrMissingRequest
	}
	if req.OutcomeSlotCount == nil {
		return PrepareConditionResponse{}, ErrMissingU256Value
	}
	tx, err := c.transact(ctx, c.conditionalTokens, "prepareCondition", req.Oracle, req.QuestionID, req.OutcomeSlotCount)
	if err != nil {
		return PrepareConditionResponse{}, err
	}
	return PrepareConditionResponse{TransactionHash: tx.Hash, BlockNumber: tx.BlockNumber}, nil
}

func (c *clientImpl) ConditionID(ctx context.Context, req *ConditionIDRequest) (ConditionIDResponse, error) {
	if req == nil {
		return ConditionIDResponse{}, ErrMissingRequest
	}
	if req.OutcomeSlotCount == nil {
		return ConditionIDResponse{}, ErrMissingU256Value
	}
	buf := make([]byte, 0, 20+32+32)
	buf = append(buf, req.Oracle.Bytes()...)
	buf = append(buf, req.QuestionID.Bytes()...)
	buf = append(buf, leftPad32(req.OutcomeSlotCount)...)
	hash := crypto.Keccak256Hash(buf)
	return ConditionIDResponse{ConditionID: hash}, nil
}

func (c *clientImpl) CollectionID(ctx context.Context, req *CollectionIDRequest) (CollectionIDResponse, error) {
	if req == nil {
		return CollectionIDResponse{}, ErrMissingRequest
	}
	if req.IndexSet == nil {
		return CollectionIDResponse{}, ErrMissingU256Value
	}
	buf := make([]byte, 0, 32+32+32)
	buf = append(buf, req.ParentCollectionID.Bytes()...)
	buf = append(buf, req.ConditionID.Bytes()...)
	buf = append(buf, leftPad32(req.IndexSet)...)
	hash := crypto.Keccak256Hash(buf)
	return CollectionIDResponse{CollectionID: hash}, nil
}

func (c *clientImpl) PositionID(ctx context.Context, req *PositionIDRequest) (PositionIDResponse, error) {
	if req == nil {
		return PositionIDResponse{}, ErrMissingRequest
	}
	buf := make([]byte, 0, 20+32)
	buf = append(buf, req.CollateralToken.Bytes()...)
	buf = append(buf, req.CollectionID.Bytes()...)
	hash := crypto.Keccak256Hash(buf)
	return PositionIDResponse{PositionID: new(big.Int).SetBytes(hash.Bytes())}, nil
}

func (c *clientImpl) SplitPosition(ctx context.Context, req *SplitPositionRequest) (SplitPositionResponse, error) {
	if req == nil {
		return SplitPositionResponse{}, ErrMissingRequest
	}
	if req.Amount == nil {
		return SplitPositionResponse{}, ErrMissingU256Value
	}
	if len(req.Partition) == 0 {
		return SplitPositionResponse{}, fmt.Errorf("partition is required")
	}
	tx, err := c.transact(ctx, c.conditionalTokens, "splitPosition",
		req.CollateralToken, req.ParentCollectionID, req.ConditionID, req.Partition, req.Amount)
	if err != nil {
		return SplitPositionResponse{}, err
	}
	return SplitPositionResponse{TransactionHash: tx.Hash, BlockNumber: tx.BlockNumber}, nil
}

func (c *clientImpl) MergePositions(ctx context.Context, req *MergePositionsRequest) (MergePositionsResponse, error) {
	if req == nil {
		return MergePositionsResponse{}, ErrMissingRequest
	}
	if req.Amount == nil {
		return MergePositionsResponse{}, ErrMissingU256Value
	}
	if len(req.Partition) == 0 {
		return MergePositionsResponse{}, fmt.Errorf("partition is required")
	}
	tx, err := c.transact(ctx, c.conditionalTokens, "mergePositions",
		req.CollateralToken, req.ParentCollectionID, req.ConditionID, req.Partition, req.Amount)
	if err != nil {
		return MergePositionsResponse{}, err
	}
	return MergePositionsResponse{TransactionHash: tx.Hash, BlockNumber: tx.BlockNumber}, nil
}

func (c *clientImpl) RedeemPositions(ctx context.Context, req *RedeemPositionsRequest) (RedeemPositionsResponse, error) {
	if req == nil {
		return RedeemPositionsResponse{}, ErrMissingRequest
	}
	if len(req.IndexSets) == 0 {
		return RedeemPositionsResponse{}, fmt.Errorf("index_sets is required")
	}
	tx, err := c.transact(ctx, c.conditionalTokens, "redeemPositions",
		req.CollateralToken, req.ParentCollectionID, req.ConditionID, req.IndexSets)
	if err != nil {
		return RedeemPositionsResponse{}, err
	}
	return RedeemPositionsResponse{TransactionHash: tx.Hash, BlockNumber: tx.BlockNumber}, nil
}

func (c *clientImpl) RedeemNegRisk(ctx context.Context, req *RedeemNegRiskRequest) (RedeemNegRiskResponse, error) {
	if req == nil {
		return RedeemNegRiskResponse{}, ErrMissingRequest
	}
	if len(req.Amounts) == 0 {
		return RedeemNegRiskResponse{}, fmt.Errorf("amounts is required")
	}
	if c.negRiskAdapter == nil {
		return RedeemNegRiskResponse{}, ErrNegRiskAdapter
	}
	tx, err := c.transact(ctx, c.negRiskAdapter, "redeemPositions", req.ConditionID, req.Amounts)
	if err != nil {
		return RedeemNegRiskResponse{}, err
	}
	return RedeemNegRiskResponse{TransactionHash: tx.Hash, BlockNumber: tx.BlockNumber}, nil
}

type txResult struct {
	Hash        common.Hash
	BlockNumber uint64
}

func (c *clientImpl) transact(ctx context.Context, contract *bind.BoundContract, method string, args ...interface{}) (txResult, error) {
	if c.backend == nil || contract == nil {
		return txResult{}, ErrMissingBackend
	}
	if c.txOpts == nil {
		return txResult{}, ErrMissingTransactor
	}
	opts := *c.txOpts
	opts.Context = ctx

	tx, err := contract.Transact(&opts, method, args...)
	if err != nil {
		return txResult{}, fmt.Errorf("send %s: %w", method, err)
	}
	receipt, err := bind.WaitMined(ctx, c.backend, tx)
	if err != nil {
		return txResult{}, fmt.Errorf("wait %s receipt: %w", method, err)
	}
	if receipt == nil || receipt.BlockNumber == nil {
		return txResult{}, errors.New("receipt missing block number")
	}
	return txResult{Hash: tx.Hash(), BlockNumber: receipt.BlockNumber.Uint64()}, nil
}

func leftPad32(value *big.Int) []byte {
	if value == nil {
		return make([]byte, 32)
	}
	raw := value.Bytes()
	if len(raw) >= 32 {
		return raw[len(raw)-32:]
	}
	padded := make([]byte, 32)
	copy(padded[32-len(raw):], raw)
	return padded
}
