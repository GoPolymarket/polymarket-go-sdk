package data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type staticDoer struct {
	responses map[string]string
}

func (d *staticDoer) Do(req *http.Request) (*http.Response, error) {
	key := req.URL.Path
	if req.URL.RawQuery != "" {
		key += "?" + req.URL.RawQuery
	}
	payload, ok := d.responses[key]
	if !ok {
		// Try matching by path only for tests that don't care about exact query params.
		reqPath := req.URL.Path
		for k, v := range d.responses {
			if k == reqPath {
				payload = v
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil, fmt.Errorf("unexpected request %q", key)
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(payload)),
		Header:     make(http.Header),
	}
	return resp, nil
}

func intPtr(v int) *int       { return &v }
func int64Ptr(v int64) *int64 { return &v }
func strPtr(v string) *string { return &v }
func boolPtr(v bool) *bool    { return &v }

func TestNewClientNilTransport(t *testing.T) {
	// Should not panic with nil transport.
	client := NewClient(nil)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestHealthSuccess(t *testing.T) {
	doer := &staticDoer{responses: map[string]string{
		"/": `{"data":"OK"}`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "OK" {
		t.Errorf("expected OK, got %q", resp)
	}
}

func TestNilRequests(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Positions", func() error { _, err := client.Positions(ctx, nil); return err }},
		{"Trades", func() error { _, err := client.Trades(ctx, nil); return err }},
		{"Activity", func() error { _, err := client.Activity(ctx, nil); return err }},
		{"Holders", func() error { _, err := client.Holders(ctx, nil); return err }},
		{"Value", func() error { _, err := client.Value(ctx, nil); return err }},
		{"ClosedPositions", func() error { _, err := client.ClosedPositions(ctx, nil); return err }},
		{"Traded", func() error { _, err := client.Traded(ctx, nil); return err }},
		{"OpenInterest", func() error { _, err := client.OpenInterest(ctx, nil); return err }},
		{"LiveVolume", func() error { _, err := client.LiveVolume(ctx, nil); return err }},
		{"Leaderboard", func() error { _, err := client.Leaderboard(ctx, nil); return err }},
		{"BuildersLeaderboard", func() error { _, err := client.BuildersLeaderboard(ctx, nil); return err }},
		{"BuildersVolume", func() error { _, err := client.BuildersVolume(ctx, nil); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.fn(), ErrMissingRequest) {
				t.Error("expected ErrMissingRequest")
			}
		})
	}
}

func TestPositionsMissingUser(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Positions(context.Background(), &PositionsRequest{})
	if !errors.Is(err, ErrMissingUser) {
		t.Errorf("expected ErrMissingUser, got %v", err)
	}
}

func TestActivityMissingUser(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Activity(context.Background(), &ActivityRequest{})
	if !errors.Is(err, ErrMissingUser) {
		t.Errorf("expected ErrMissingUser, got %v", err)
	}
}

func TestClosedPositionsMissingUser(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.ClosedPositions(context.Background(), &ClosedPositionsRequest{})
	if !errors.Is(err, ErrMissingUser) {
		t.Errorf("expected ErrMissingUser, got %v", err)
	}
}

func TestPositionsSuccess(t *testing.T) {
	user := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	doer := &staticDoer{responses: map[string]string{
		"/positions": `[{"title":"Test Market","size":"100"}]`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.Positions(context.Background(), &PositionsRequest{User: user})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 position, got %d", len(resp))
	}
	if resp[0].Title != "Test Market" {
		t.Errorf("expected title 'Test Market', got %q", resp[0].Title)
	}
}

func TestTradesSuccess(t *testing.T) {
	doer := &staticDoer{responses: map[string]string{
		"/trades": `[{"side":"BUY","title":"Test"}]`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.Trades(context.Background(), &TradesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 || resp[0].Side != SideBuy {
		t.Errorf("unexpected trades response: %+v", resp)
	}
}

func TestTradedSuccess(t *testing.T) {
	user := common.HexToAddress("0xaaaa")
	doer := &staticDoer{responses: map[string]string{
		"/traded?user=" + user.Hex(): `{"user":"0x000000000000000000000000000000000000aAaA","traded":5}`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.Traded(context.Background(), &TradedRequest{User: user})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Traded != 5 {
		t.Errorf("expected traded=5, got %d", resp.Traded)
	}
}

func TestLiveVolumeRequiresID(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.LiveVolume(context.Background(), &LiveVolumeRequest{ID: 0})
	if err == nil {
		t.Error("expected error for zero ID")
	}
}

func TestLiveVolumeSuccess(t *testing.T) {
	doer := &staticDoer{responses: map[string]string{
		"/live-volume?id=42": `[{"total":"1000","markets":[]}]`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.LiveVolume(context.Background(), &LiveVolumeRequest{ID: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 live volume entry, got %d", len(resp))
	}
}

func TestOpenInterestSuccess(t *testing.T) {
	doer := &staticDoer{responses: map[string]string{
		"/oi": `[{"market":"global","value":"5000"}]`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.OpenInterest(context.Background(), &OpenInterestRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp))
	}
}

func TestHoldersSuccess(t *testing.T) {
	doer := &staticDoer{responses: map[string]string{
		"/holders": `[]`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.Holders(context.Background(), &HoldersRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty holders, got %d", len(resp))
	}
}

func TestBuildersVolumeSuccess(t *testing.T) {
	doer := &staticDoer{responses: map[string]string{
		"/v1/builders/volume": `[{"builder":"test","volume":"100"}]`,
	}}
	client := NewClient(transport.NewClient(doer, "http://example"))
	resp, err := client.BuildersVolume(context.Background(), &BuildersVolumeRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp))
	}
}

// Validation: out-of-range parameters.

func TestPositionsLimitOutOfRange(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Positions(context.Background(), &PositionsRequest{User: user, Limit: intPtr(501)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestPositionsOffsetOutOfRange(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Positions(context.Background(), &PositionsRequest{User: user, Offset: intPtr(10001)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestPositionsTitleTooLong(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	longTitle := strings.Repeat("a", 101)
	_, err := client.Positions(context.Background(), &PositionsRequest{User: user, Title: &longTitle})
	if err == nil {
		t.Error("expected error for title too long")
	}
}

func TestTradesLimitOutOfRange(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Trades(context.Background(), &TradesRequest{Limit: intPtr(10001)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestActivityLimitOutOfRange(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Activity(context.Background(), &ActivityRequest{User: user, Limit: intPtr(501)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestActivityNegativeStart(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Activity(context.Background(), &ActivityRequest{User: user, Start: int64Ptr(-1)})
	if err == nil {
		t.Error("expected error for negative start")
	}
}

func TestActivityNegativeEnd(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Activity(context.Background(), &ActivityRequest{User: user, End: int64Ptr(-1)})
	if err == nil {
		t.Error("expected error for negative end")
	}
}

func TestHoldersLimitOutOfRange(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Holders(context.Background(), &HoldersRequest{Limit: intPtr(21)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestHoldersMinBalanceOutOfRange(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Holders(context.Background(), &HoldersRequest{MinBalance: intPtr(1000000)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestClosedPositionsLimitOutOfRange(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.ClosedPositions(context.Background(), &ClosedPositionsRequest{User: user, Limit: intPtr(51)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestLeaderboardLimitOutOfRange(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Leaderboard(context.Background(), &LeaderboardRequest{Limit: intPtr(51)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestLeaderboardLimitBelowMin(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Leaderboard(context.Background(), &LeaderboardRequest{Limit: intPtr(0)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

func TestBuildersLeaderboardLimitOutOfRange(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.BuildersLeaderboard(context.Background(), &BuildersLeaderboardRequest{Limit: intPtr(51)})
	var bounded BoundedIntError
	if !errors.As(err, &bounded) {
		t.Errorf("expected BoundedIntError, got %v", err)
	}
}

// MarketFilter validation.

func TestMarketFilterBothMarketsAndEvents(t *testing.T) {
	user := common.HexToAddress("0x1111111111111111111111111111111111111111")
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Positions(context.Background(), &PositionsRequest{
		User: user,
		Filter: &MarketFilter{
			Markets:  []common.Hash{common.HexToHash("0x01")},
			EventIDs: []string{"event1"},
		},
	})
	if !errors.Is(err, ErrInvalidMarketFilter) {
		t.Errorf("expected ErrInvalidMarketFilter, got %v", err)
	}
}

// TradeFilter validation.

func TestTradeFilterMissingType(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	_, err := client.Trades(context.Background(), &TradesRequest{
		TradeFilter: &TradeFilter{FilterType: ""},
	})
	if !errors.Is(err, ErrInvalidTradeFilter) {
		t.Errorf("expected ErrInvalidTradeFilter, got %v", err)
	}
}

func TestTradeFilterNegativeAmount(t *testing.T) {
	client := NewClient(transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"))
	negAmount := decimal.NewFromInt(-1)
	_, err := client.Trades(context.Background(), &TradesRequest{
		TradeFilter: &TradeFilter{FilterType: FilterCash, FilterAmount: negAmount},
	})
	if err == nil {
		t.Error("expected error for negative filter amount")
	}
}

// BoundedIntError.

func TestBoundedIntErrorMessage(t *testing.T) {
	err := BoundedIntError{Value: 999, Min: 0, Max: 500, ParamName: "limit"}
	msg := err.Error()
	if !strings.Contains(msg, "limit") || !strings.Contains(msg, "999") {
		t.Errorf("unexpected error message: %s", msg)
	}
}

// FlexibleTime JSON.

func TestFlexibleTimeUnmarshalRFC3339(t *testing.T) {
	var ft FlexibleTime
	if err := json.Unmarshal([]byte(`"2025-01-15T10:30:00Z"`), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ft.Year() != 2025 || ft.Month() != 1 || ft.Day() != 15 {
		t.Errorf("unexpected time: %v", ft.Time)
	}
}

func TestFlexibleTimeUnmarshalDateOnly(t *testing.T) {
	var ft FlexibleTime
	if err := json.Unmarshal([]byte(`"2025-06-01"`), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ft.Year() != 2025 || ft.Month() != 6 || ft.Day() != 1 {
		t.Errorf("unexpected time: %v", ft.Time)
	}
}

func TestFlexibleTimeUnmarshalEmpty(t *testing.T) {
	var ft FlexibleTime
	if err := json.Unmarshal([]byte(`""`), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ft.IsZero() {
		t.Error("expected zero time for empty string")
	}
}

func TestFlexibleTimeUnmarshalNull(t *testing.T) {
	var ft FlexibleTime
	if err := json.Unmarshal([]byte(`null`), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ft.IsZero() {
		t.Error("expected zero time for null")
	}
}

func TestFlexibleTimeUnmarshalInvalid(t *testing.T) {
	var ft FlexibleTime
	err := json.Unmarshal([]byte(`"not-a-date"`), &ft)
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestFlexibleTimeMarshalZero(t *testing.T) {
	ft := FlexibleTime{}
	data, err := json.Marshal(ft)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("expected null, got %s", data)
	}
}

// Market JSON.

func TestMarketUnmarshalGlobal(t *testing.T) {
	var m Market
	if err := json.Unmarshal([]byte(`"global"`), &m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.Global {
		t.Error("expected Global=true")
	}
}

func TestMarketUnmarshalHex(t *testing.T) {
	var m Market
	if err := json.Unmarshal([]byte(`"0xabcdef0000000000000000000000000000000000000000000000000000000001"`), &m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Global {
		t.Error("expected Global=false")
	}
	if m.ID == (common.Hash{}) {
		t.Error("expected non-zero hash")
	}
}

func TestMarketUnmarshal64CharHex(t *testing.T) {
	hex64 := "abcdef0000000000000000000000000000000000000000000000000000000001"
	var m Market
	if err := json.Unmarshal([]byte(`"`+hex64+`"`), &m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID == (common.Hash{}) {
		t.Error("expected non-zero hash")
	}
}

func TestMarketMarshalGlobal(t *testing.T) {
	m := Market{Global: true}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `"global"` {
		t.Errorf("expected \"global\", got %s", data)
	}
}

func TestMarketMarshalRaw(t *testing.T) {
	m := Market{Raw: "test-raw"}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `"test-raw"` {
		t.Errorf("expected \"test-raw\", got %s", data)
	}
}

// IntString JSON.

func TestIntStringUnmarshalNumber(t *testing.T) {
	var i IntString
	if err := json.Unmarshal([]byte(`42`), &i); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(i) != 42 {
		t.Errorf("expected 42, got %d", i)
	}
}

func TestIntStringUnmarshalString(t *testing.T) {
	var i IntString
	if err := json.Unmarshal([]byte(`"42"`), &i); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(i) != 42 {
		t.Errorf("expected 42, got %d", i)
	}
}

func TestIntStringUnmarshalEmptyString(t *testing.T) {
	var i IntString
	if err := json.Unmarshal([]byte(`""`), &i); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(i) != 0 {
		t.Errorf("expected 0, got %d", i)
	}
}

// Helper constructors.

func TestMarketFilterByMarkets(t *testing.T) {
	markets := []common.Hash{common.HexToHash("0x01")}
	f := MarketFilterByMarkets(markets)
	if len(f.Markets) != 1 {
		t.Errorf("expected 1 market, got %d", len(f.Markets))
	}
}

func TestMarketFilterByEventIDs(t *testing.T) {
	f := MarketFilterByEventIDs([]string{"e1", "e2"})
	if len(f.EventIDs) != 2 {
		t.Errorf("expected 2 event IDs, got %d", len(f.EventIDs))
	}
}

func TestTradeFilterConstructors(t *testing.T) {
	amount := decimal.NewFromInt(100)
	cash := TradeFilterCash(amount)
	if cash.FilterType != FilterCash {
		t.Errorf("expected CASH, got %s", cash.FilterType)
	}
	tokens := TradeFilterTokens(amount)
	if tokens.FilterType != FilterTokens {
		t.Errorf("expected TOKENS, got %s", tokens.FilterType)
	}
}
