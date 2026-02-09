package rfq

import (
	"math/big"
	"net/url"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func TestRFQRequestItemToDetail(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "req-1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		Side:         "BUY",
		SizeIn:       "10",
		SizeOut:      "5",
		Price:        "0.5",
		Expiry:       123456,
	}

	detail, err := item.ToDetail()
	if err != nil {
		t.Fatalf("ToDetail failed: %v", err)
	}
	if detail.RequestID != "req-1" {
		t.Fatalf("requestID mismatch: %s", detail.RequestID)
	}
	if detail.TokenID == nil || detail.TokenID.String() != "123" {
		t.Fatalf("tokenID mismatch: %v", detail.TokenID)
	}
	if detail.Price.String() != "0.5" {
		t.Fatalf("price mismatch: %s", detail.Price.String())
	}
}

func TestRFQQuoteItemToDetail(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "quote-1",
		RequestID:    "req-1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		Side:         "SELL",
		SizeIn:       "10",
		SizeOut:      "5",
		Price:        "0.5",
	}

	detail, err := item.ToDetail()
	if err != nil {
		t.Fatalf("ToDetail failed: %v", err)
	}
	if detail.QuoteID != "quote-1" {
		t.Fatalf("quoteID mismatch: %s", detail.QuoteID)
	}
	if detail.TokenID == nil || detail.TokenID.String() != "123" {
		t.Fatalf("tokenID mismatch: %v", detail.TokenID)
	}
}

func TestBuildRFQAcceptRequestFromSignedOrder(t *testing.T) {
	signed := clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Salt:        types.U256{Int: big.NewInt(1)},
			Maker:       common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:      common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:     types.U256{Int: big.NewInt(123)},
			MakerAmount: decimal.NewFromInt(100),
			TakerAmount: decimal.NewFromInt(50),
			Side:        "BUY",
			Expiration:  types.U256{Int: big.NewInt(0)},
			FeeRateBps:  decimal.NewFromInt(0),
			Nonce:       types.U256{Int: big.NewInt(10)},
		},
		Signature: "0xsig",
		Owner:     "owner",
	}

	req, err := BuildRFQAcceptRequestFromSignedOrder("req-1", "quote-1", &signed)
	if err != nil {
		t.Fatalf("BuildRFQAcceptRequestFromSignedOrder failed: %v", err)
	}
	if req.RequestID != "req-1" || req.QuoteIDV2 != "quote-1" {
		t.Fatalf("request/quote IDs mismatch")
	}
	if req.TokenID != "123" || req.Nonce != "10" {
		t.Fatalf("order fields mismatch: token=%s nonce=%s", req.TokenID, req.Nonce)
	}
}

func TestBuildRFQAcceptRequest_EmptyRequestID(t *testing.T) {
	signed := &clobtypes.SignedOrder{Signature: "sig", Owner: "owner"}
	_, err := BuildRFQAcceptRequestFromSignedOrder("", "q1", signed)
	if err == nil {
		t.Fatal("expected error for empty requestID")
	}
}

func TestBuildRFQAcceptRequest_EmptyQuoteID(t *testing.T) {
	signed := &clobtypes.SignedOrder{Signature: "sig", Owner: "owner"}
	_, err := BuildRFQAcceptRequestFromSignedOrder("r1", "", signed)
	if err == nil {
		t.Fatal("expected error for empty quoteID")
	}
}

func TestBuildRFQAcceptRequest_NilSigned(t *testing.T) {
	_, err := BuildRFQAcceptRequestFromSignedOrder("r1", "q1", nil)
	if err == nil {
		t.Fatal("expected error for nil signed order")
	}
}

func TestBuildRFQAcceptRequest_EmptySignature(t *testing.T) {
	signed := &clobtypes.SignedOrder{Owner: "owner"}
	_, err := BuildRFQAcceptRequestFromSignedOrder("r1", "q1", signed)
	if err == nil {
		t.Fatal("expected error for empty signature")
	}
}

func TestBuildRFQAcceptRequest_EmptyOwner(t *testing.T) {
	signed := &clobtypes.SignedOrder{Signature: "sig"}
	_, err := BuildRFQAcceptRequestFromSignedOrder("r1", "q1", signed)
	if err == nil {
		t.Fatal("expected error for empty owner")
	}
}

func TestBuildRFQAcceptRequest_NilTokenID(t *testing.T) {
	signed := &clobtypes.SignedOrder{
		Order: clobtypes.Order{
			TokenID: types.U256{},
			Nonce:   types.U256{Int: big.NewInt(1)},
			Salt:    types.U256{Int: big.NewInt(1)},
		},
		Signature: "sig",
		Owner:     "owner",
	}
	_, err := BuildRFQAcceptRequestFromSignedOrder("r1", "q1", signed)
	if err == nil {
		t.Fatal("expected error for nil tokenID")
	}
}

func TestBuildRFQAcceptRequest_NilExpiration(t *testing.T) {
	signed := &clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Salt:        types.U256{Int: big.NewInt(1)},
			Maker:       common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:      common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:     types.U256{Int: big.NewInt(123)},
			MakerAmount: decimal.NewFromInt(100),
			TakerAmount: decimal.NewFromInt(50),
			Side:        "BUY",
			Expiration:  types.U256{}, // nil Int
			FeeRateBps:  decimal.NewFromInt(0),
			Nonce:       types.U256{Int: big.NewInt(10)},
		},
		Signature: "0xsig",
		Owner:     "owner",
	}
	req, err := BuildRFQAcceptRequestFromSignedOrder("r1", "q1", signed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Expiration != "0" {
		t.Fatalf("expected expiration 0, got %s", req.Expiration)
	}
}

func TestBuildRFQApproveQuoteFromSignedOrder(t *testing.T) {
	signed := &clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Salt:        types.U256{Int: big.NewInt(1)},
			Maker:       common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:      common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:     types.U256{Int: big.NewInt(123)},
			MakerAmount: decimal.NewFromInt(100),
			TakerAmount: decimal.NewFromInt(50),
			Side:        "BUY",
			Expiration:  types.U256{Int: big.NewInt(999)},
			FeeRateBps:  decimal.NewFromInt(0),
			Nonce:       types.U256{Int: big.NewInt(10)},
		},
		Signature: "0xsig",
		Owner:     "owner",
	}
	req, err := BuildRFQApproveQuoteFromSignedOrder("r1", "q1", signed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.RequestID != "r1" || req.QuoteIDV2 != "q1" {
		t.Fatalf("IDs mismatch")
	}
	if req.Expiration != "999" {
		t.Fatalf("expected expiration 999, got %s", req.Expiration)
	}
	if req.TokenID != "123" {
		t.Fatalf("expected tokenID 123, got %s", req.TokenID)
	}
}

func TestBuildRFQApproveQuote_EmptyRequestID(t *testing.T) {
	_, err := BuildRFQApproveQuoteFromSignedOrder("", "q1", &clobtypes.SignedOrder{Signature: "s", Owner: "o"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildRFQApproveQuote_NilSigned(t *testing.T) {
	_, err := BuildRFQApproveQuoteFromSignedOrder("r1", "q1", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildRFQApproveQuote_EmptySignature(t *testing.T) {
	_, err := BuildRFQApproveQuoteFromSignedOrder("r1", "q1", &clobtypes.SignedOrder{Owner: "o"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildRFQApproveQuote_EmptyOwner(t *testing.T) {
	_, err := BuildRFQApproveQuoteFromSignedOrder("r1", "q1", &clobtypes.SignedOrder{Signature: "s"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildRFQApproveQuote_NilNonce(t *testing.T) {
	signed := &clobtypes.SignedOrder{
		Order: clobtypes.Order{
			TokenID: types.U256{Int: big.NewInt(1)},
			Nonce:   types.U256{},
			Salt:    types.U256{Int: big.NewInt(1)},
		},
		Signature: "sig",
		Owner:     "owner",
	}
	_, err := BuildRFQApproveQuoteFromSignedOrder("r1", "q1", signed)
	if err == nil {
		t.Fatal("expected error for nil nonce")
	}
}

// --------------- parse helpers ---------------

func TestParseBigIntString_Empty(t *testing.T) {
	v, err := parseBigIntString("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != nil {
		t.Fatal("expected nil")
	}
}

func TestParseBigIntString_Valid(t *testing.T) {
	v, err := parseBigIntString("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.String() != "12345" {
		t.Fatalf("expected 12345, got %s", v.String())
	}
}

func TestParseBigIntString_Invalid(t *testing.T) {
	_, err := parseBigIntString("not-a-number")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseDecimalString_Empty(t *testing.T) {
	v, err := parseDecimalString("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v.IsZero() {
		t.Fatal("expected zero")
	}
}

func TestParseDecimalString_Valid(t *testing.T) {
	v, err := parseDecimalString("1.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.String() != "1.5" {
		t.Fatalf("expected 1.5, got %s", v.String())
	}
}

func TestParseDecimalString_Invalid(t *testing.T) {
	_, err := parseDecimalString("abc")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseAddress_Empty(t *testing.T) {
	addr, err := parseAddress("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != (common.Address{}) {
		t.Fatal("expected zero address")
	}
}

func TestParseAddress_Valid(t *testing.T) {
	addr, err := parseAddress("0x0000000000000000000000000000000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != common.HexToAddress("0x0000000000000000000000000000000000000001") {
		t.Fatal("address mismatch")
	}
}

func TestParseAddress_Invalid(t *testing.T) {
	_, err := parseAddress("not-an-address")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --------------- ToDetail edge cases ---------------

func TestRFQRequestItem_ToDetail_FallbackID(t *testing.T) {
	item := RFQRequestItem{
		ID:          "fallback-id",
		UserAddress: "0x0000000000000000000000000000000000000001",
	}
	detail, err := item.ToDetail()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.RequestID != "fallback-id" {
		t.Fatalf("expected fallback-id, got %s", detail.RequestID)
	}
}

func TestRFQRequestItem_ToDetail_InvalidAddress(t *testing.T) {
	item := RFQRequestItem{
		RequestID:   "r1",
		UserAddress: "bad-addr",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error for invalid address")
	}
}

func TestRFQRequestItem_ToDetail_InvalidProxy(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "r1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "bad-addr",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error for invalid proxy address")
	}
}

func TestRFQRequestItem_ToDetail_InvalidToken(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "r1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "not-a-number",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestRFQRequestItem_ToDetail_InvalidComplement(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "r1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error for invalid complement")
	}
}

func TestRFQRequestItem_ToDetail_InvalidSizeIn(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "r1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		SizeIn:       "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error for invalid sizeIn")
	}
}

func TestRFQRequestItem_ToDetail_InvalidSizeOut(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "r1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		SizeIn:       "10",
		SizeOut:      "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error for invalid sizeOut")
	}
}

func TestRFQRequestItem_ToDetail_InvalidPrice(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "r1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		SizeIn:       "10",
		SizeOut:      "5",
		Price:        "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error for invalid price")
	}
}

func TestRFQQuoteItem_ToDetail_FallbackID(t *testing.T) {
	item := RFQQuoteItem{
		ID:          "fallback-id",
		UserAddress: "0x0000000000000000000000000000000000000001",
	}
	detail, err := item.ToDetail()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.QuoteID != "fallback-id" {
		t.Fatalf("expected fallback-id, got %s", detail.QuoteID)
	}
}

func TestRFQQuoteItem_ToDetail_InvalidAddress(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:     "q1",
		UserAddress: "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRFQQuoteItem_ToDetail_InvalidProxy(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "q1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRFQQuoteItem_ToDetail_InvalidToken(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "q1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRFQQuoteItem_ToDetail_InvalidComplement(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "q1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRFQQuoteItem_ToDetail_InvalidSizeIn(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "q1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		SizeIn:       "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRFQQuoteItem_ToDetail_InvalidSizeOut(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "q1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		SizeIn:       "10",
		SizeOut:      "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRFQQuoteItem_ToDetail_InvalidPrice(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "q1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		SizeIn:       "10",
		SizeOut:      "5",
		Price:        "bad",
	}
	_, err := item.ToDetail()
	if err == nil {
		t.Fatal("expected error")
	}
}

// --------------- applyRFQPagination ---------------

func TestApplyRFQPagination_NilValues(t *testing.T) {
	// Should not panic
	applyRFQPagination(nil, 0, "", "")
}

func TestApplyRFQPagination_WithLimit(t *testing.T) {
	q := make(url.Values)
	applyRFQPagination(&q, 10, "", "")
	if q.Get("limit") != "10" {
		t.Fatalf("expected limit 10, got %s", q.Get("limit"))
	}
}

func TestApplyRFQPagination_ZeroLimit(t *testing.T) {
	q := make(url.Values)
	applyRFQPagination(&q, 0, "", "")
	if q.Get("limit") != "" {
		t.Fatalf("expected no limit, got %s", q.Get("limit"))
	}
}

func TestApplyRFQPagination_CursorSetsOffset(t *testing.T) {
	q := make(url.Values)
	applyRFQPagination(&q, 0, "", "cursor123")
	if q.Get("cursor") != "cursor123" {
		t.Fatalf("expected cursor123, got %s", q.Get("cursor"))
	}
	if q.Get("offset") != "cursor123" {
		t.Fatalf("expected offset=cursor123, got %s", q.Get("offset"))
	}
}

func TestApplyRFQPagination_OffsetOnly(t *testing.T) {
	q := make(url.Values)
	applyRFQPagination(&q, 0, "off1", "")
	if q.Get("offset") != "off1" {
		t.Fatalf("expected off1, got %s", q.Get("offset"))
	}
	if q.Get("cursor") != "" {
		t.Fatalf("expected no cursor, got %s", q.Get("cursor"))
	}
}

// --------------- applyRFQFilters ---------------

func TestApplyRFQFilters_NilValues(t *testing.T) {
	// Should not panic
	applyRFQFilters(nil, "", nil, nil, nil, "", "", "", "", "", "", "", "")
}

func TestApplyRFQFilters_AllFields(t *testing.T) {
	q := make(url.Values)
	applyRFQFilters(&q,
		RFQStateActive,
		[]string{"r1", "r2"},
		[]string{"q1"},
		[]string{"m1", "m2"},
		"1", "100",
		"10", "1000",
		"0.1", "0.9",
		RFQSortBySize,
		RFQSortDirDesc,
	)
	if q.Get("state") != "active" {
		t.Fatalf("expected active, got %s", q.Get("state"))
	}
	if q.Get("requestIds") != "r1,r2" {
		t.Fatalf("expected r1,r2, got %s", q.Get("requestIds"))
	}
	if q.Get("quoteIds") != "q1" {
		t.Fatalf("expected q1, got %s", q.Get("quoteIds"))
	}
	if q.Get("markets") != "m1,m2" {
		t.Fatalf("expected m1,m2, got %s", q.Get("markets"))
	}
	if q.Get("sizeMin") != "1" {
		t.Fatalf("expected 1, got %s", q.Get("sizeMin"))
	}
	if q.Get("size_min") != "1" {
		t.Fatalf("expected 1, got %s", q.Get("size_min"))
	}
	if q.Get("sizeMax") != "100" {
		t.Fatalf("expected 100, got %s", q.Get("sizeMax"))
	}
	if q.Get("sizeUsdcMin") != "10" {
		t.Fatalf("expected 10, got %s", q.Get("sizeUsdcMin"))
	}
	if q.Get("sizeUsdcMax") != "1000" {
		t.Fatalf("expected 1000, got %s", q.Get("sizeUsdcMax"))
	}
	if q.Get("priceMin") != "0.1" {
		t.Fatalf("expected 0.1, got %s", q.Get("priceMin"))
	}
	if q.Get("priceMax") != "0.9" {
		t.Fatalf("expected 0.9, got %s", q.Get("priceMax"))
	}
	if q.Get("sortBy") != "size" {
		t.Fatalf("expected size, got %s", q.Get("sortBy"))
	}
	if q.Get("sortDir") != "desc" {
		t.Fatalf("expected desc, got %s", q.Get("sortDir"))
	}
}

func TestApplyRFQFilters_EmptyFields(t *testing.T) {
	q := make(url.Values)
	applyRFQFilters(&q, "", nil, nil, nil, "", "", "", "", "", "", "", "")
	if len(q) != 0 {
		t.Fatalf("expected empty query, got %v", q)
	}
}
