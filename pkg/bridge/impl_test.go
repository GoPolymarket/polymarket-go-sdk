package bridge

import (
	"bytes"
	"context"
	"io"
	"math/big"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

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
