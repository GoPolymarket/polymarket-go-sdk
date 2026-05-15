package rfq

import (
	"fmt"
	"math/big"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// RFQRequestDetail provides parsed/typed RFQ request fields.
type RFQRequestDetail struct {
	RequestID    string
	UserAddress  common.Address
	ProxyAddress common.Address
	Condition    string
	TokenID      *big.Int
	Complement   *big.Int
	Side         string
	SizeIn       decimal.Decimal
	SizeOut      decimal.Decimal
	Price        decimal.Decimal
	Expiry       int64
}

// RFQQuoteDetail provides parsed/typed RFQ quote fields.
type RFQQuoteDetail struct {
	QuoteID      string
	RequestID    string
	UserAddress  common.Address
	ProxyAddress common.Address
	Condition    string
	TokenID      *big.Int
	Complement   *big.Int
	Side         string
	SizeIn       decimal.Decimal
	SizeOut      decimal.Decimal
	Price        decimal.Decimal
}

// BuildRFQAcceptRequestFromSignedOrder builds an RFQ accept payload from a signed order.
func BuildRFQAcceptRequestFromSignedOrder(requestID, quoteID string, signed *clobtypes.SignedOrder) (*RFQAcceptRequest, error) {
	if requestID == "" || quoteID == "" {
		return nil, fmt.Errorf("requestID and quoteID are required")
	}
	if signed == nil {
		return nil, fmt.Errorf("signed order is required")
	}
	if signed.Signature == "" {
		return nil, fmt.Errorf("signature is required")
	}
	if signed.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}

	order := signed.Order
	if order.TokenID.Int == nil || order.Salt.Int == nil {
		return nil, fmt.Errorf("order token/salt are required")
	}

	expiration := "0"
	if order.Expiration.Int != nil {
		expiration = order.Expiration.Int.String()
	}

	req := &RFQAcceptRequest{
		RequestID:   requestID,
		QuoteID:     quoteID,
		QuoteIDV2:   quoteID,
		MakerAmount: order.MakerAmount.String(),
		TakerAmount: order.TakerAmount.String(),
		TokenID:     order.TokenID.Int.String(),
		Maker:       order.Maker.Hex(),
		Signer:      order.Signer.Hex(),
		Expiration:  expiration,
		Side:        order.Side,
		Signature:   signed.Signature,
		Salt:        order.Salt.Int.String(),
		Owner:       signed.Owner,
		Timestamp:   order.Timestamp,
		Metadata:    order.Metadata,
		Builder:     order.Builder,
	}
	return req, nil
}

// BuildRFQApproveQuoteFromSignedOrder builds an RFQ approve payload from a signed order.
func BuildRFQApproveQuoteFromSignedOrder(requestID, quoteID string, signed *clobtypes.SignedOrder) (*RFQApproveQuote, error) {
	if requestID == "" || quoteID == "" {
		return nil, fmt.Errorf("requestID and quoteID are required")
	}
	if signed == nil {
		return nil, fmt.Errorf("signed order is required")
	}
	if signed.Signature == "" {
		return nil, fmt.Errorf("signature is required")
	}
	if signed.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}

	order := signed.Order
	if order.TokenID.Int == nil || order.Salt.Int == nil {
		return nil, fmt.Errorf("order token/salt are required")
	}

	expiration := "0"
	if order.Expiration.Int != nil {
		expiration = order.Expiration.Int.String()
	}

	req := &RFQApproveQuote{
		RequestID:   requestID,
		QuoteID:     quoteID,
		QuoteIDV2:   quoteID,
		MakerAmount: order.MakerAmount.String(),
		TakerAmount: order.TakerAmount.String(),
		TokenID:     order.TokenID.Int.String(),
		Maker:       order.Maker.Hex(),
		Signer:      order.Signer.Hex(),
		Expiration:  expiration,
		Side:        order.Side,
		Signature:   signed.Signature,
		Salt:        order.Salt.Int.String(),
		Owner:       signed.Owner,
		Timestamp:   order.Timestamp,
		Metadata:    order.Metadata,
		Builder:     order.Builder,
	}
	return req, nil
}

func (r RFQRequestItem) ToDetail() (RFQRequestDetail, error) {
	requestID := r.RequestID
	if requestID == "" {
		requestID = r.ID
	}
	user, err := parseAddress(r.UserAddress)
	if err != nil {
		return RFQRequestDetail{}, err
	}
	proxy, err := parseAddress(r.ProxyAddress)
	if err != nil {
		return RFQRequestDetail{}, err
	}
	tokenID, err := parseBigIntString(r.Token)
	if err != nil {
		return RFQRequestDetail{}, err
	}
	complement, err := parseBigIntString(r.Complement)
	if err != nil {
		return RFQRequestDetail{}, err
	}
	sizeIn, err := parseDecimalString(r.SizeIn)
	if err != nil {
		return RFQRequestDetail{}, err
	}
	sizeOut, err := parseDecimalString(r.SizeOut)
	if err != nil {
		return RFQRequestDetail{}, err
	}
	price, err := parseDecimalString(r.Price)
	if err != nil {
		return RFQRequestDetail{}, err
	}

	return RFQRequestDetail{
		RequestID:    requestID,
		UserAddress:  user,
		ProxyAddress: proxy,
		Condition:    r.Condition,
		TokenID:      tokenID,
		Complement:   complement,
		Side:         r.Side,
		SizeIn:       sizeIn,
		SizeOut:      sizeOut,
		Price:        price,
		Expiry:       r.Expiry,
	}, nil
}

func (r RFQQuoteItem) ToDetail() (RFQQuoteDetail, error) {
	quoteID := r.QuoteID
	if quoteID == "" {
		quoteID = r.ID
	}
	user, err := parseAddress(r.UserAddress)
	if err != nil {
		return RFQQuoteDetail{}, err
	}
	proxy, err := parseAddress(r.ProxyAddress)
	if err != nil {
		return RFQQuoteDetail{}, err
	}
	tokenID, err := parseBigIntString(r.Token)
	if err != nil {
		return RFQQuoteDetail{}, err
	}
	complement, err := parseBigIntString(r.Complement)
	if err != nil {
		return RFQQuoteDetail{}, err
	}
	sizeIn, err := parseDecimalString(r.SizeIn)
	if err != nil {
		return RFQQuoteDetail{}, err
	}
	sizeOut, err := parseDecimalString(r.SizeOut)
	if err != nil {
		return RFQQuoteDetail{}, err
	}
	price, err := parseDecimalString(r.Price)
	if err != nil {
		return RFQQuoteDetail{}, err
	}

	return RFQQuoteDetail{
		QuoteID:      quoteID,
		RequestID:    r.RequestID,
		UserAddress:  user,
		ProxyAddress: proxy,
		Condition:    r.Condition,
		TokenID:      tokenID,
		Complement:   complement,
		Side:         r.Side,
		SizeIn:       sizeIn,
		SizeOut:      sizeOut,
		Price:        price,
	}, nil
}

func parseBigIntString(value string) (*big.Int, error) {
	if value == "" {
		return nil, nil
	}
	parsed, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer %q", value)
	}
	return parsed, nil
}

func parseDecimalString(value string) (decimal.Decimal, error) {
	if value == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(value)
}

func parseAddress(value string) (common.Address, error) {
	if value == "" {
		return common.Address{}, nil
	}
	if !common.IsHexAddress(value) {
		return common.Address{}, fmt.Errorf("invalid address %q", value)
	}
	return common.HexToAddress(value), nil
}
