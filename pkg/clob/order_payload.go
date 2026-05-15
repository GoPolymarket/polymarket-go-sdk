package clob

import "github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/types"
)

func buildOrderPayload(order *clobtypes.SignedOrder) (map[string]interface{}, error) {
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}
	orderType := normalizeOrderType(order.OrderType, clobtypes.OrderTypeGTC)
	if order.PostOnly != nil && *order.PostOnly && orderType != clobtypes.OrderTypeGTC && orderType != clobtypes.OrderTypeGTD {
		return nil, fmt.Errorf("postOnly is only supported for GTC and GTD orders")
	}
	orderMap, err := orderWithSignature(order)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"order":     orderMap,
		"owner":     order.Owner,
		"orderType": orderType,
	}
	if order.PostOnly != nil {
		payload["postOnly"] = *order.PostOnly
	}
	if order.DeferExec != nil {
		payload["deferExec"] = *order.DeferExec
	}
	return payload, nil
}

func buildOrdersPayload(orders *clobtypes.SignedOrders) ([]map[string]interface{}, error) {
	if orders == nil {
		return nil, fmt.Errorf("orders are required")
	}
	payloads := make([]map[string]interface{}, 0, len(orders.Orders))
	for idx := range orders.Orders {
		order := orders.Orders[idx]
		payload, err := buildOrderPayload(&order)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}
	return payloads, nil
}

func orderWithSignature(order *clobtypes.SignedOrder) (map[string]interface{}, error) {
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}
	if order.Signature == "" {
		return nil, fmt.Errorf("signature is required")
	}
	if order.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}

	sigType := 0
	if order.Order.SignatureType != nil {
		sigType = *order.Order.SignatureType
	}

	side := strings.ToUpper(order.Order.Side)
	if side != "BUY" && side != "SELL" {
		return nil, fmt.Errorf("invalid order side %q", order.Order.Side)
	}

	payload := map[string]interface{}{
		"salt":          u256String(order.Order.Salt),
		"maker":         order.Order.Maker.Hex(),
		"signer":        order.Order.Signer.Hex(),
		"tokenId":       u256String(order.Order.TokenID),
		"makerAmount":   decimalString(order.Order.MakerAmount),
		"takerAmount":   decimalString(order.Order.TakerAmount),
		"side":          side,
		"expiration":    u256String(order.Order.Expiration),
		"signatureType": sigType,
		"signature":     order.Signature,
	}

	// V2 fields - always include to match EIP-712 signed values
	payload["timestamp"] = order.Order.Timestamp
	payload["metadata"] = padBytes32(order.Order.Metadata)
	payload["builder"] = padBytes32(order.Order.Builder)
	return payload, nil
}

func u256String(value types.U256) string {
	if value.Int == nil {
		return "0"
	}
	return value.Int.String()
}

func decimalString(value types.Decimal) string {
	return value.String()
}

func normalizeOrderType(orderType clobtypes.OrderType, fallback clobtypes.OrderType) clobtypes.OrderType {
	trimmed := strings.TrimSpace(string(orderType))
	if trimmed == "" {
		return fallback
	}
	upper := strings.ToUpper(trimmed)
	return clobtypes.OrderType(upper)
}

// padBytes32 pads a hex string to exactly 32 bytes (right-aligned).
// Returns the zero bytes32 representation for empty or invalid input.
func padBytes32(hexStr string) string {
	zeroBytes32 := "0x0000000000000000000000000000000000000000000000000000000000000000"
	if hexStr == "" {
		return zeroBytes32
	}
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		b, err := hexutil.Decode(hexStr)
		if err != nil || len(b) > 32 {
			return zeroBytes32
		}
		var padded [32]byte
		copy(padded[32-len(b):], b)
		return hexutil.Encode(padded[:])
	}
	return zeroBytes32
}
