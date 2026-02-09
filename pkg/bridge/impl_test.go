package bridge

import (
	"bytes"
	"context"
	"io"
	"math/big"
	"net/http"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// mockBackend implements bind.ContractBackend for testing.
type mockBackend struct{}

func (m *mockBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return nil, nil
}
func (m *mockBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return nil, nil
}
func (m *mockBackend) HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error) {
	return nil, nil
}
func (m *mockBackend) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	return nil, nil
}
func (m *mockBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return 0, nil
}
func (m *mockBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (m *mockBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (m *mockBackend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	return 0, nil
}
func (m *mockBackend) SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error {
	return nil
}
func (m *mockBackend) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]ethtypes.Log, error) {
	return nil, nil
}
func (m *mockBackend) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- ethtypes.Log) (ethereum.Subscription, error) {
	return nil, nil
}

type mockDoer struct {
	responses map[string]string
}

func (m *mockDoer) reset() { m.responses = make(map[string]string) }
func (m *mockDoer) addResponse(path string, body string) {
	if m.responses == nil {
		m.responses = make(map[string]string)
	}
	m.responses[path] = body
}
func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	body := m.responses[req.URL.Path]
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}, nil
}

func TestBridgeMethods(t *testing.T) {
	mock := &mockDoer{}
	client := NewClient(transport.NewClient(mock, BaseURL))
	ctx := context.Background()

	t.Run("SupportedAssetsInfo", func(t *testing.T) {
		mock.reset()
		mock.addResponse("/supported-assets", `{"supported_assets":[]}`)
		_, _ = client.SupportedAssetsInfo(ctx)
	})

	t.Run("DepositAddress", func(t *testing.T) {
		mock.reset()
		mock.addResponse("/deposit", `{"address":{"evm":"0x123"}}`)
		_, _ = client.DepositAddress(ctx, &DepositRequest{Address: "0x123"})
	})

	t.Run("Status", func(t *testing.T) {
		mock.reset()
		mock.addResponse("/status/0x123", `{"transactions":[]}`)
		_, _ = client.Status(ctx, &StatusRequest{Address: "0x123"})
	})

	t.Run("SupportedAssets", func(t *testing.T) {
		mock.reset()
		mock.addResponse("/supported-assets", `{"supported_assets":[{"token":{"address":"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}}]}`)
		assets, err := client.SupportedAssets(ctx)
		if err != nil || len(assets) == 0 {
			t.Errorf("SupportedAssets failed: %v", err)
		}
	})

	t.Run("WithdrawTo", func(t *testing.T) {
		_, err := client.WithdrawTo(ctx, &WithdrawRequest{
			To:     common.HexToAddress("0x123"),
			Amount: big.NewInt(100),
			Asset:  common.HexToAddress("0xabc"),
		})
		if err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("DepositError", func(t *testing.T) {
		_, err := client.Deposit(ctx, big.NewInt(100), common.HexToAddress("0xabc"))
		if err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("WithdrawError", func(t *testing.T) {
		_, err := client.Withdraw(ctx, big.NewInt(100), common.HexToAddress("0xabc"))
		if err == nil {
			t.Errorf("expected error")
		}
	})
}

// --------------- NewClient ---------------

func TestNewClient_NilTransport(t *testing.T) {
	c := NewClient(nil)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithTransport(t *testing.T) {
	mock := &mockDoer{}
	c := NewClient(transport.NewClient(mock, BaseURL))
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// --------------- NewClientWithBackend ---------------

func TestNewClientWithBackend_NilBackend(t *testing.T) {
	_, err := NewClientWithBackend(nil, nil, &bind.TransactOpts{})
	if err != ErrMissingBackend {
		t.Fatalf("expected ErrMissingBackend, got %v", err)
	}
}

func TestNewClientWithBackend_NilTxOpts(t *testing.T) {
	_, err := NewClientWithBackend(nil, &mockBackend{}, nil)
	if err != ErrMissingTransactor {
		t.Fatalf("expected ErrMissingTransactor, got %v", err)
	}
}

func TestNewClientWithBackend_NilHTTPClient(t *testing.T) {
	c, err := NewClientWithBackend(nil, &mockBackend{}, &bind.TransactOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientWithBackend_AllProvided(t *testing.T) {
	mock := &mockDoer{}
	c, err := NewClientWithBackend(transport.NewClient(mock, BaseURL), &mockBackend{}, &bind.TransactOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// --------------- DepositAddress ---------------

func TestDepositAddress_NilRequest(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.DepositAddress(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestDepositAddress_EmptyAddress(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.DepositAddress(context.Background(), &DepositRequest{Address: ""})
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestDepositAddress_Success(t *testing.T) {
	mock := &mockDoer{}
	mock.addResponse("/deposit", `{"address":{"evm":"0xabc","svm":"sol123","btc":"bc1q"}}`)
	c := NewClient(transport.NewClient(mock, "http://example"))
	resp, err := c.DepositAddress(context.Background(), &DepositRequest{Address: "0x123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Address.EVM != "0xabc" {
		t.Fatalf("expected 0xabc, got %s", resp.Address.EVM)
	}
}

// --------------- Status ---------------

func TestStatus_NilRequest(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.Status(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestStatus_EmptyAddress(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.Status(context.Background(), &StatusRequest{Address: ""})
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestStatus_Success(t *testing.T) {
	mock := &mockDoer{}
	mock.addResponse("/status/0xabc", `{"transactions":[{"status":"completed","tx_hash":"0x1"}]}`)
	c := NewClient(transport.NewClient(mock, "http://example"))
	resp, err := c.Status(context.Background(), &StatusRequest{Address: "0xabc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(resp.Transactions))
	}
	if resp.Transactions[0].Status != "completed" {
		t.Fatalf("expected completed, got %s", resp.Transactions[0].Status)
	}
}

// --------------- SupportedAssets filtering ---------------

func TestSupportedAssets_EmptyList(t *testing.T) {
	mock := &mockDoer{}
	mock.addResponse("/supported-assets", `{"supported_assets":[]}`)
	c := NewClient(transport.NewClient(mock, "http://example"))
	assets, err := c.SupportedAssets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("expected 0 assets, got %d", len(assets))
	}
}

func TestSupportedAssets_FiltersInvalidAddresses(t *testing.T) {
	mock := &mockDoer{}
	mock.addResponse("/supported-assets", `{"supported_assets":[
		{"token":{"address":"not-hex"}},
		{"token":{"address":""}},
		{"token":{"address":"0x0000000000000000000000000000000000000000"}},
		{"token":{"address":"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}}
	]}`)
	c := NewClient(transport.NewClient(mock, "http://example"))
	assets, err := c.SupportedAssets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 valid asset, got %d", len(assets))
	}
}

func TestSupportedAssets_DeduplicatesAddresses(t *testing.T) {
	mock := &mockDoer{}
	mock.addResponse("/supported-assets", `{"supported_assets":[
		{"token":{"address":"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}},
		{"token":{"address":"0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"}}
	]}`)
	c := NewClient(transport.NewClient(mock, "http://example"))
	assets, err := c.SupportedAssets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 deduplicated asset, got %d", len(assets))
	}
}

// --------------- Withdraw ---------------

func TestWithdraw_AlwaysReturnsError(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.Withdraw(context.Background(), big.NewInt(100), common.HexToAddress("0x1"))
	if err != ErrWithdrawUnsupported {
		t.Fatalf("expected ErrWithdrawUnsupported, got %v", err)
	}
}

// --------------- WithdrawTo validation ---------------

func TestWithdrawTo_NilRequest(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.WithdrawTo(context.Background(), nil)
	if err != ErrMissingWithdrawRequest {
		t.Fatalf("expected ErrMissingWithdrawRequest, got %v", err)
	}
}

func TestWithdrawTo_NilAmount(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.WithdrawTo(context.Background(), &WithdrawRequest{
		Amount: nil,
		Asset:  common.HexToAddress("0x1"),
		To:     common.HexToAddress("0x2"),
	})
	if err == nil {
		t.Fatal("expected error for nil amount")
	}
}

func TestWithdrawTo_ZeroAmount(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.WithdrawTo(context.Background(), &WithdrawRequest{
		Amount: big.NewInt(0),
		Asset:  common.HexToAddress("0x1"),
		To:     common.HexToAddress("0x2"),
	})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestWithdrawTo_NegativeAmount(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.WithdrawTo(context.Background(), &WithdrawRequest{
		Amount: big.NewInt(-1),
		Asset:  common.HexToAddress("0x1"),
		To:     common.HexToAddress("0x2"),
	})
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestWithdrawTo_ZeroAsset(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.WithdrawTo(context.Background(), &WithdrawRequest{
		Amount: big.NewInt(100),
		Asset:  common.Address{},
		To:     common.HexToAddress("0x2"),
	})
	if err == nil {
		t.Fatal("expected error for zero asset")
	}
}

func TestWithdrawTo_ZeroTo(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.WithdrawTo(context.Background(), &WithdrawRequest{
		Amount: big.NewInt(100),
		Asset:  common.HexToAddress("0x1"),
		To:     common.Address{},
	})
	if err != ErrMissingWithdrawAddress {
		t.Fatalf("expected ErrMissingWithdrawAddress, got %v", err)
	}
}

func TestWithdrawTo_NoBackend(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.WithdrawTo(context.Background(), &WithdrawRequest{
		Amount: big.NewInt(100),
		Asset:  common.HexToAddress("0x1"),
		To:     common.HexToAddress("0x2"),
	})
	if err != ErrMissingBackend {
		t.Fatalf("expected ErrMissingBackend, got %v", err)
	}
}

// --------------- Deposit validation ---------------

func TestDeposit_NilAmount(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.Deposit(context.Background(), nil, common.HexToAddress("0x1"))
	if err == nil {
		t.Fatal("expected error for nil amount")
	}
}

func TestDeposit_ZeroAmount(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.Deposit(context.Background(), big.NewInt(0), common.HexToAddress("0x1"))
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestDeposit_ZeroAsset(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.Deposit(context.Background(), big.NewInt(100), common.Address{})
	if err == nil {
		t.Fatal("expected error for zero asset")
	}
}

func TestDeposit_NoBackend(t *testing.T) {
	c := NewClient(transport.NewClient(&mockDoer{}, BaseURL))
	_, err := c.Deposit(context.Background(), big.NewInt(100), common.HexToAddress("0x1"))
	if err != ErrMissingBackend {
		t.Fatalf("expected ErrMissingBackend, got %v", err)
	}
}

// --------------- parseERC20ABI ---------------

func TestParseERC20ABI(t *testing.T) {
	parsed, err := parseERC20ABI()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := parsed.Methods["transfer"]; !ok {
		t.Fatal("expected transfer method in ABI")
	}
}

// --------------- Error sentinels ---------------

func TestErrorSentinels(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrMissingBackend", ErrMissingBackend},
		{"ErrMissingTransactor", ErrMissingTransactor},
		{"ErrMissingFromAddress", ErrMissingFromAddress},
		{"ErrMissingDepositAddress", ErrMissingDepositAddress},
		{"ErrWithdrawUnsupported", ErrWithdrawUnsupported},
		{"ErrMissingWithdrawRequest", ErrMissingWithdrawRequest},
		{"ErrMissingWithdrawAddress", ErrMissingWithdrawAddress},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Fatal("error message should not be empty")
			}
		})
	}
}
