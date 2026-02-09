package clobtypes

import (
	"encoding/json"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

func TestOrderTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		orderType OrderType
		expected string
	}{
		{"GTC", OrderTypeGTC, "GTC"},
		{"GTD", OrderTypeGTD, "GTD"},
		{"FAK", OrderTypeFAK, "FAK"},
		{"FOK", OrderTypeFOK, "FOK"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.orderType) != tt.expected {
				t.Errorf("OrderType = %s, want %s", tt.orderType, tt.expected)
			}
		})
	}
}

func TestPriceHistoryIntervalConstants(t *testing.T) {
	tests := []struct {
		name     string
		interval PriceHistoryInterval
		expected string
	}{
		{"1m", PriceHistoryInterval1m, "1m"},
		{"1h", PriceHistoryInterval1h, "1h"},
		{"6h", PriceHistoryInterval6h, "6h"},
		{"1d", PriceHistoryInterval1d, "1d"},
		{"1w", PriceHistoryInterval1w, "1w"},
		{"max", PriceHistoryIntervalMax, "max"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.interval) != tt.expected {
				t.Errorf("PriceHistoryInterval = %s, want %s", tt.interval, tt.expected)
			}
		})
	}
}

func TestAssetTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		assetType AssetType
		expected  string
	}{
		{"COLLATERAL", AssetTypeCollateral, "COLLATERAL"},
		{"CONDITIONAL", AssetTypeConditional, "CONDITIONAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.assetType) != tt.expected {
				t.Errorf("AssetType = %s, want %s", tt.assetType, tt.expected)
			}
		})
	}
}

func TestCursorConstants(t *testing.T) {
	if InitialCursor != "MA==" {
		t.Errorf("InitialCursor = %s, want MA==", InitialCursor)
	}
	if EndCursor != "LTE=" {
		t.Errorf("EndCursor = %s, want LTE=", EndCursor)
	}
}

func TestMarketsRequest_JSON(t *testing.T) {
	active := true
	req := MarketsRequest{
		Limit:   10,
		Cursor:  "test_cursor",
		Active:  &active,
		AssetID: "asset123",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded MarketsRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Limit != req.Limit {
		t.Errorf("Limit = %d, want %d", decoded.Limit, req.Limit)
	}
	if decoded.Cursor != req.Cursor {
		t.Errorf("Cursor = %s, want %s", decoded.Cursor, req.Cursor)
	}
	if decoded.Active == nil || *decoded.Active != *req.Active {
		t.Errorf("Active mismatch")
	}
	if decoded.AssetID != req.AssetID {
		t.Errorf("AssetID = %s, want %s", decoded.AssetID, req.AssetID)
	}
}

func TestSignedOrder_JSON(t *testing.T) {
	postOnly := true
	deferExec := false
	order := SignedOrder{
		Order: Order{
			Salt:        types.U256{},
			Signer:      types.Address{},
			Maker:       types.Address{},
			Taker:       types.Address{},
			TokenID:     types.U256{},
			MakerAmount: types.Decimal{},
			TakerAmount: types.Decimal{},
			Expiration:  types.U256{},
			Side:        "BUY",
			FeeRateBps:  types.Decimal{},
			Nonce:       types.U256{},
		},
		Signature: "0xsignature",
		Owner:     "0xowner",
		OrderType: OrderTypeGTC,
		PostOnly:  &postOnly,
		DeferExec: &deferExec,
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify that OrderType, PostOnly, DeferExec are not serialized (json:"-")
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	if _, exists := raw["order_type"]; exists {
		t.Error("OrderType should not be serialized (json:\"-\")")
	}
	if _, exists := raw["post_only"]; exists {
		t.Error("PostOnly should not be serialized (json:\"-\")")
	}
	if _, exists := raw["defer_exec"]; exists {
		t.Error("DeferExec should not be serialized (json:\"-\")")
	}

	// Verify that order, signature, owner are serialized
	if _, exists := raw["order"]; !exists {
		t.Error("Order should be serialized")
	}
	if _, exists := raw["signature"]; !exists {
		t.Error("Signature should be serialized")
	}
	if _, exists := raw["owner"]; !exists {
		t.Error("Owner should be serialized")
	}
}

func TestPricesHistoryResponse_UnmarshalJSON_Array(t *testing.T) {
	// Test direct array format
	jsonData := `[
		{"t": 1234567890, "p": 0.5},
		{"t": 1234567900, "p": 0.6}
	]`

	var resp PricesHistoryResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(resp))
	}

	if resp[0].Timestamp != 1234567890 {
		t.Errorf("Point[0].Timestamp = %d, want 1234567890", resp[0].Timestamp)
	}
	if resp[0].Price != 0.5 {
		t.Errorf("Point[0].Price = %f, want 0.5", resp[0].Price)
	}
	if resp[1].Timestamp != 1234567900 {
		t.Errorf("Point[1].Timestamp = %d, want 1234567900", resp[1].Timestamp)
	}
	if resp[1].Price != 0.6 {
		t.Errorf("Point[1].Price = %f, want 0.6", resp[1].Price)
	}
}

func TestPricesHistoryResponse_UnmarshalJSON_HistoryWrapper(t *testing.T) {
	// Test wrapped format with "history" key
	jsonData := `{
		"history": [
			{"t": 1234567890, "p": 0.5},
			{"t": 1234567900, "p": 0.6}
		]
	}`

	var resp PricesHistoryResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(resp))
	}

	if resp[0].Timestamp != 1234567890 {
		t.Errorf("Point[0].Timestamp = %d, want 1234567890", resp[0].Timestamp)
	}
	if resp[0].Price != 0.5 {
		t.Errorf("Point[0].Price = %f, want 0.5", resp[0].Price)
	}
}

func TestPricesHistoryResponse_UnmarshalJSON_DataWrapper(t *testing.T) {
	// Test wrapped format with "data" key
	jsonData := `{
		"data": [
			{"t": 1234567890, "p": 0.5},
			{"t": 1234567900, "p": 0.6}
		]
	}`

	var resp PricesHistoryResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("Expected 2 points, got %d", len(resp))
	}

	if resp[0].Timestamp != 1234567890 {
		t.Errorf("Point[0].Timestamp = %d, want 1234567890", resp[0].Timestamp)
	}
}

func TestPricesHistoryResponse_UnmarshalJSON_Null(t *testing.T) {
	// Test null value
	jsonData := `null`

	var resp PricesHistoryResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp != nil {
		t.Errorf("Expected nil response for null input, got %v", resp)
	}
}

func TestPricesHistoryResponse_UnmarshalJSON_EmptyArray(t *testing.T) {
	// Test empty array
	jsonData := `[]`

	var resp PricesHistoryResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(resp) != 0 {
		t.Errorf("Expected empty array, got %v", resp)
	}
}

func TestPricesHistoryResponse_UnmarshalJSON_EmptyWrapper(t *testing.T) {
	// Test empty wrapper object
	jsonData := `{}`

	var resp PricesHistoryResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if resp != nil {
		t.Errorf("Expected nil response for empty wrapper, got %v", resp)
	}
}

func TestPricesHistoryResponse_UnmarshalJSON_HistoryPriority(t *testing.T) {
	// Test that "history" takes priority over "data" when both are present
	jsonData := `{
		"history": [{"t": 1111111111, "p": 0.1}],
		"data": [{"t": 2222222222, "p": 0.2}]
	}`

	var resp PricesHistoryResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("Expected 1 point, got %d", len(resp))
	}

	// Should use "history" value, not "data"
	if resp[0].Timestamp != 1111111111 {
		t.Errorf("Point[0].Timestamp = %d, want 1111111111 (from history)", resp[0].Timestamp)
	}
}

func TestOrderBook_JSON(t *testing.T) {
	orderBook := OrderBook{
		MarketID: "market123",
		Bids: []PriceLevel{
			{Price: "0.5", Size: "100"},
			{Price: "0.4", Size: "200"},
		},
		Asks: []PriceLevel{
			{Price: "0.6", Size: "150"},
			{Price: "0.7", Size: "250"},
		},
		Hash: "hash123",
	}

	data, err := json.Marshal(orderBook)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded OrderBook
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.MarketID != orderBook.MarketID {
		t.Errorf("MarketID = %s, want %s", decoded.MarketID, orderBook.MarketID)
	}
	if len(decoded.Bids) != len(orderBook.Bids) {
		t.Errorf("Bids length = %d, want %d", len(decoded.Bids), len(orderBook.Bids))
	}
	if len(decoded.Asks) != len(orderBook.Asks) {
		t.Errorf("Asks length = %d, want %d", len(decoded.Asks), len(orderBook.Asks))
	}
	if decoded.Hash != orderBook.Hash {
		t.Errorf("Hash = %s, want %s", decoded.Hash, orderBook.Hash)
	}
}

func TestMarket_JSON(t *testing.T) {
	market := Market{
		ID:          "market123",
		Question:    "Will it rain tomorrow?",
		ConditionID: "condition123",
		Slug:        "rain-tomorrow",
		Resolution:  "YES",
		EndDate:     "2026-12-31",
		Tokens: []MarketToken{
			{TokenID: "token1", Outcome: "YES", Price: 0.6},
			{TokenID: "token2", Outcome: "NO", Price: 0.4},
		},
		Active: true,
		Closed: false,
	}

	data, err := json.Marshal(market)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Market
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != market.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, market.ID)
	}
	if decoded.Question != market.Question {
		t.Errorf("Question = %s, want %s", decoded.Question, market.Question)
	}
	if decoded.Active != market.Active {
		t.Errorf("Active = %v, want %v", decoded.Active, market.Active)
	}
	if len(decoded.Tokens) != len(market.Tokens) {
		t.Errorf("Tokens length = %d, want %d", len(decoded.Tokens), len(market.Tokens))
	}
}

func TestCancelOrderRequest_JSON(t *testing.T) {
	req := CancelOrderRequest{
		OrderID: "order123",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify the JSON uses "orderId" (not "order_id")
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	if _, exists := raw["orderId"]; !exists {
		t.Error("Expected 'orderId' field in JSON")
	}
	if _, exists := raw["order_id"]; exists {
		t.Error("Should not have 'order_id' field in JSON")
	}
}

func TestCancelOrdersRequest_JSON(t *testing.T) {
	req := CancelOrdersRequest{
		OrderIDs: []string{"order1", "order2", "order3"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify the JSON uses "orderIds" (not "order_ids")
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	if _, exists := raw["orderIds"]; !exists {
		t.Error("Expected 'orderIds' field in JSON")
	}
}

func TestBalanceAllowanceRequest_JSON(t *testing.T) {
	sigType := 1
	req := BalanceAllowanceRequest{
		Asset:         "USDC",
		AssetType:     AssetTypeCollateral,
		TokenID:       "token123",
		SignatureType: &sigType,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded BalanceAllowanceRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Asset != req.Asset {
		t.Errorf("Asset = %s, want %s", decoded.Asset, req.Asset)
	}
	if decoded.AssetType != req.AssetType {
		t.Errorf("AssetType = %s, want %s", decoded.AssetType, req.AssetType)
	}
	if decoded.TokenID != req.TokenID {
		t.Errorf("TokenID = %s, want %s", decoded.TokenID, req.TokenID)
	}
	if decoded.SignatureType == nil || *decoded.SignatureType != *req.SignatureType {
		t.Error("SignatureType mismatch")
	}
}

func TestTradesResponse_JSON(t *testing.T) {
	resp := TradesResponse{
		Data: []Trade{
			{ID: "trade1", Price: "0.5", Size: "100", Side: "BUY", Timestamp: 1234567890},
			{ID: "trade2", Price: "0.6", Size: "200", Side: "SELL", Timestamp: 1234567900},
		},
		NextCursor: "cursor123",
		Limit:      10,
		Count:      2,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded TradesResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded.Data) != len(resp.Data) {
		t.Errorf("Data length = %d, want %d", len(decoded.Data), len(resp.Data))
	}
	if decoded.NextCursor != resp.NextCursor {
		t.Errorf("NextCursor = %s, want %s", decoded.NextCursor, resp.NextCursor)
	}
	if decoded.Limit != resp.Limit {
		t.Errorf("Limit = %d, want %d", decoded.Limit, resp.Limit)
	}
	if decoded.Count != resp.Count {
		t.Errorf("Count = %d, want %d", decoded.Count, resp.Count)
	}
}

func TestPricesHistoryRequest_JSON(t *testing.T) {
	req := PricesHistoryRequest{
		Market:     "market123",
		TokenID:    "token123",
		Interval:   PriceHistoryInterval1h,
		StartTs:    1234567890,
		EndTs:      1234567900,
		Resolution: "1h",
		Fidelity:   100,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded PricesHistoryRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Market != req.Market {
		t.Errorf("Market = %s, want %s", decoded.Market, req.Market)
	}
	if decoded.Interval != req.Interval {
		t.Errorf("Interval = %s, want %s", decoded.Interval, req.Interval)
	}
	if decoded.StartTs != req.StartTs {
		t.Errorf("StartTs = %d, want %d", decoded.StartTs, req.StartTs)
	}
	if decoded.Fidelity != req.Fidelity {
		t.Errorf("Fidelity = %d, want %d", decoded.Fidelity, req.Fidelity)
	}
}

func TestNotification_JSON(t *testing.T) {
	notification := Notification{
		ID:      "notif123",
		Title:   "Test Notification",
		Content: "This is a test",
	}

	data, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Notification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != notification.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, notification.ID)
	}
	if decoded.Title != notification.Title {
		t.Errorf("Title = %s, want %s", decoded.Title, notification.Title)
	}
	if decoded.Content != notification.Content {
		t.Errorf("Content = %s, want %s", decoded.Content, notification.Content)
	}
}

func TestGeoblockResponse_JSON(t *testing.T) {
	resp := GeoblockResponse{
		Blocked: true,
		IP:      "192.168.1.1",
		Country: "US",
		Region:  "CA",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded GeoblockResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Blocked != resp.Blocked {
		t.Errorf("Blocked = %v, want %v", decoded.Blocked, resp.Blocked)
	}
	if decoded.IP != resp.IP {
		t.Errorf("IP = %s, want %s", decoded.IP, resp.IP)
	}
	if decoded.Country != resp.Country {
		t.Errorf("Country = %s, want %s", decoded.Country, resp.Country)
	}
}

func TestAPIKeyResponse_JSON(t *testing.T) {
	resp := APIKeyResponse{
		APIKey:     "key123",
		Secret:     "secret123",
		Passphrase: "pass123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded APIKeyResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.APIKey != resp.APIKey {
		t.Errorf("APIKey = %s, want %s", decoded.APIKey, resp.APIKey)
	}
	if decoded.Secret != resp.Secret {
		t.Errorf("Secret = %s, want %s", decoded.Secret, resp.Secret)
	}
	if decoded.Passphrase != resp.Passphrase {
		t.Errorf("Passphrase = %s, want %s", decoded.Passphrase, resp.Passphrase)
	}
}
