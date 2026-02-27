package clob

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

var benchSigner auth.Signer

func init() {
	var err error
	benchSigner, err = auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	if err != nil {
		panic(err)
	}
}

func BenchmarkNewOrderBuilder(b *testing.B) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewOrderBuilder(stub, benchSigner).
			TokenID("123456789012345678901234567890").
			Side("BUY").
			Price(0.55).
			Size(100).
			OrderType(clobtypes.OrderTypeGTC)
	}
}

func BenchmarkOrderBuilderBuild(b *testing.B) {
	stub := newStubClient()
	stub.tickSize = 0.01
	stub.feeRate = 0
	stub.book = clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.55", Size: "1000"},
		},
		Bids: []clobtypes.PriceLevel{
			{Price: "0.54", Size: "1000"},
		},
	}

	builder := NewOrderBuilder(stub, benchSigner).
		TokenID("123456789012345678901234567890").
		Side("BUY").
		Price(0.55).
		Size(100).
		OrderType(clobtypes.OrderTypeGTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Build()
	}
}

func BenchmarkJSONMarshalOrder(b *testing.B) {
	order := map[string]interface{}{
		"exchange":     "0x4bFb41d5Bfc6C6Cfb53D0a3A6c14a6B4D7a7E8d",
		"maker":        "0x1234567890123456789012345678901234567890",
		"token_id":     "123456789012345678901234567890",
		"side":         "BUY",
		"price":        "0.55",
		"size":         "100",
		"fee_rate_bps": 0,
		"nonce":        12345678,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(order)
	}
}

func BenchmarkJSONUnmarshalOrder(b *testing.B) {
	data := `{
		"exchange":"0x4bFb41d5Bfc6C6Cfb53D0a3A6c14a6B4D7a7E8d",
		"maker":"0x1234567890123456789012345678901234567890",
		"token_id":"123456789012345678901234567890",
		"side":"BUY",
		"price":"0.55",
		"size":"100",
		"fee_rate_bps":0,
		"nonce":12345678
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var order map[string]interface{}
		_ = json.Unmarshal([]byte(data), &order)
	}
}

func BenchmarkJSONMarshalOrderBook(b *testing.B) {
	book := clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.55", Size: "1000"},
			{Price: "0.56", Size: "2000"},
			{Price: "0.57", Size: "3000"},
			{Price: "0.58", Size: "4000"},
			{Price: "0.59", Size: "5000"},
		},
		Bids: []clobtypes.PriceLevel{
			{Price: "0.54", Size: "1000"},
			{Price: "0.53", Size: "2000"},
			{Price: "0.52", Size: "3000"},
			{Price: "0.51", Size: "4000"},
			{Price: "0.50", Size: "5000"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(book)
	}
}

func BenchmarkJSONUnmarshalOrderBook(b *testing.B) {
	data := `{
		"asks":[
			{"price":"0.55","size":"1000"},
			{"price":"0.56","size":"2000"},
			{"price":"0.57","size":"3000"},
			{"price":"0.58","size":"4000"},
			{"price":"0.59","size":"5000"}
		],
		"bids":[
			{"price":"0.54","size":"1000"},
			{"price":"0.53","size":"2000"},
			{"price":"0.52","size":"3000"},
			{"price":"0.51","size":"4000"},
			{"price":"0.50","size":"5000"}
		]
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var book clobtypes.OrderBookResponse
		_ = json.Unmarshal([]byte(data), &book)
	}
}

func BenchmarkPrivateKeySigner(b *testing.B) {
	pk := "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = auth.NewPrivateKeySigner(pk, 137)
	}
}

func BenchmarkHMACSign(b *testing.B) {
	secret := "dGVzdF9zZWNyZXQ"
	message := "POST\n/v1/orders\n{\"token_id\":\"123\",\"side\":\"BUY\",\"price\":\"0.55\"}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = auth.SignHMAC(secret, message)
	}
}

func BenchmarkDecimalNewFromFloat(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = decimal.NewFromFloat(0.55)
	}
}

func BenchmarkCommonHexToAddress(b *testing.B) {
	addr := "0x1234567890123456789012345678901234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = common.HexToAddress(addr)
	}
}

func BenchmarkOrderBookAnalysis(b *testing.B) {
	book := clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.55", Size: "1000"},
			{Price: "0.56", Size: "2000"},
			{Price: "0.57", Size: "3000"},
			{Price: "0.58", Size: "4000"},
			{Price: "0.59", Size: "5000"},
			{Price: "0.60", Size: "6000"},
			{Price: "0.61", Size: "7000"},
			{Price: "0.62", Size: "8000"},
			{Price: "0.63", Size: "9000"},
			{Price: "0.64", Size: "10000"},
		},
		Bids: []clobtypes.PriceLevel{
			{Price: "0.54", Size: "1000"},
			{Price: "0.53", Size: "2000"},
			{Price: "0.52", Size: "3000"},
			{Price: "0.51", Size: "4000"},
			{Price: "0.50", Size: "5000"},
			{Price: "0.49", Size: "6000"},
			{Price: "0.48", Size: "7000"},
			{Price: "0.47", Size: "8000"},
			{Price: "0.46", Size: "9000"},
			{Price: "0.45", Size: "10000"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzeOrderBook(book)
	}
}

func analyzeOrderBook(book clobtypes.OrderBookResponse) {
	var bidVol, askVol float64
	for _, b := range book.Bids {
		size, _ := strconv.ParseFloat(b.Size, 64)
		bidVol += size
	}
	for _, a := range book.Asks {
		size, _ := strconv.ParseFloat(a.Size, 64)
		askVol += size
	}
	_ = bidVol
	_ = askVol
}

func BenchmarkParsePriceLevel(b *testing.B) {
	levels := []clobtypes.PriceLevel{
		{Price: "0.55", Size: "1000"},
		{Price: "0.56", Size: "2000"},
		{Price: "0.57", Size: "3000"},
		{Price: "0.58", Size: "4000"},
		{Price: "0.59", Size: "5000"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, level := range levels {
			price, _ := strconv.ParseFloat(level.Price, 64)
			size, _ := strconv.ParseFloat(level.Size, 64)
			_ = price * size
		}
	}
}
