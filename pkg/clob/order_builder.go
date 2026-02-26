package clob

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

// OrderBuilder helps construct valid orders with correct addresses and nonces.
type OrderBuilder struct {
	client Client
	signer auth.Signer

	tokenID    string
	side       string
	price      decimal.Decimal
	size       decimal.Decimal
	feeRateBps decimal.Decimal
	tickSize   float64
	orderType  clobtypes.OrderType

	// Optional overrides
	maker         *common.Address
	funder        *common.Address
	taker         *common.Address
	nonce         *big.Int
	expiration    *big.Int
	signatureType *auth.SignatureType
	postOnly      *bool

	saltGenerator SaltGenerator

	amount *marketAmount
}

type marketAmount struct {
	kind  string
	value decimal.Decimal
}

const (
	amountUSDC   = "USDC"
	amountShares = "SHARES"
)

const (
	usdcDecimals = int32(6)
	lotSizeScale = int32(2)
)

// SaltGenerator generates salts for new orders.
type SaltGenerator func() (*big.Int, error)

// NewOrderBuilder creates a new order builder.
func NewOrderBuilder(client Client, signer auth.Signer) *OrderBuilder {
	builder := &OrderBuilder{
		client: client,
		signer: signer,
	}
	if provider, ok := client.(interface{ orderDefaults() orderDefaults }); ok {
		defaults := provider.orderDefaults()
		sigType := defaults.signatureType
		builder.signatureType = &sigType
		if defaults.funder != nil {
			builder.funder = defaults.funder
		}
		builder.saltGenerator = defaults.saltGenerator
	}
	return builder
}

// TokenID sets the token ID to trade.
func (b *OrderBuilder) TokenID(tokenID string) *OrderBuilder {
	b.tokenID = tokenID
	return b
}

// Side sets the trade side ("BUY" or "SELL").
func (b *OrderBuilder) Side(side string) *OrderBuilder {
	b.side = side
	return b
}

// Price sets the price per share using a float64.
func (b *OrderBuilder) Price(price float64) *OrderBuilder {
	b.price = decimal.NewFromFloat(price)
	return b
}

// PriceDec sets the price per share using a decimal.Decimal.
func (b *OrderBuilder) PriceDec(price decimal.Decimal) *OrderBuilder {
	b.price = price
	return b
}

// Size sets the number of shares using a float64.
func (b *OrderBuilder) Size(size float64) *OrderBuilder {
	b.size = decimal.NewFromFloat(size)
	return b
}

// SizeDec sets the number of shares using a decimal.Decimal.
func (b *OrderBuilder) SizeDec(size decimal.Decimal) *OrderBuilder {
	b.size = size
	return b
}

// FeeRateBps sets the fee rate in basis points using a float64 (default 0).
func (b *OrderBuilder) FeeRateBps(bps float64) *OrderBuilder {
	b.feeRateBps = decimal.NewFromFloat(bps)
	return b
}

// FeeRateBpsDec sets the fee rate in basis points using a decimal.Decimal.
func (b *OrderBuilder) FeeRateBpsDec(bps decimal.Decimal) *OrderBuilder {
	b.feeRateBps = bps
	return b
}

// TickSize sets a manual tick size override (e.g. "0.01").
func (b *OrderBuilder) TickSize(tickSize float64) *OrderBuilder {
	b.tickSize = tickSize
	return b
}

// Nonce overrides the order nonce.
func (b *OrderBuilder) Nonce(nonce *big.Int) *OrderBuilder {
	b.nonce = nonce
	return b
}

// Maker overrides the maker address.
func (b *OrderBuilder) Maker(maker common.Address) *OrderBuilder {
	b.maker = &maker
	return b
}

// Taker overrides the taker address.
func (b *OrderBuilder) Taker(taker common.Address) *OrderBuilder {
	b.taker = &taker
	return b
}

// clobtypes.OrderType sets the order type (GTC/GTD/FAK/FOK).
func (b *OrderBuilder) OrderType(orderType clobtypes.OrderType) *OrderBuilder {
	b.orderType = orderType
	return b
}

// PostOnly sets the post-only flag for limit orders.
func (b *OrderBuilder) PostOnly(postOnly bool) *OrderBuilder {
	b.postOnly = &postOnly
	return b
}

// ExpirationUnix sets the expiration timestamp (seconds since epoch) for GTD orders.
func (b *OrderBuilder) ExpirationUnix(timestamp int64) *OrderBuilder {
	b.expiration = big.NewInt(timestamp)
	return b
}

// AmountUSDC sets the amount for a market order in USDC.
func (b *OrderBuilder) AmountUSDC(amount float64) *OrderBuilder {
	b.amount = &marketAmount{
		kind:  amountUSDC,
		value: decimal.NewFromFloat(amount),
	}
	return b
}

// AmountShares sets the amount for a market order in shares.
func (b *OrderBuilder) AmountShares(amount float64) *OrderBuilder {
	b.amount = &marketAmount{
		kind:  amountShares,
		value: decimal.NewFromFloat(amount),
	}
	return b
}

// Build constructs the clobtypes.Order object using a background context.
func (b *OrderBuilder) Build() (*clobtypes.Order, error) {
	return b.BuildWithContext(context.Background())
}

// BuildWithContext constructs the clobtypes.Order object using the provided context for API lookups.
func (b *OrderBuilder) BuildWithContext(ctx context.Context) (*clobtypes.Order, error) {
	order, err := b.buildLimit(ctx)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// BuildSignable constructs a limit order and returns it with order type metadata.
func (b *OrderBuilder) BuildSignable() (*clobtypes.SignableOrder, error) {
	return b.BuildSignableWithContext(context.Background())
}

// BuildSignableWithContext constructs a limit order and returns it with order type metadata.
func (b *OrderBuilder) BuildSignableWithContext(ctx context.Context) (*clobtypes.SignableOrder, error) {
	order, err := b.buildLimit(ctx)
	if err != nil {
		return nil, err
	}

	orderType := normalizeOrderType(b.orderType, clobtypes.OrderTypeGTC)
	if b.expiration != nil && b.expiration.Sign() > 0 && orderType != clobtypes.OrderTypeGTD {
		return nil, fmt.Errorf("expiration is only supported for GTD orders")
	}
	if orderType == clobtypes.OrderTypeGTD && (b.expiration == nil || b.expiration.Sign() == 0) {
		return nil, fmt.Errorf("GTD orders require a non-zero expiration")
	}
	if b.postOnly != nil && *b.postOnly && orderType != clobtypes.OrderTypeGTC && orderType != clobtypes.OrderTypeGTD {
		return nil, fmt.Errorf("postOnly is only supported for GTC and GTD orders")
	}

	return &clobtypes.SignableOrder{
		Order:     order,
		OrderType: orderType,
		PostOnly:  b.postOnly,
	}, nil
}

// BuildMarket constructs a market order and returns it with order type metadata.
func (b *OrderBuilder) BuildMarket() (*clobtypes.SignableOrder, error) {
	return b.BuildMarketWithContext(context.Background())
}

// BuildMarketWithContext constructs a market order and returns it with order type metadata.
func (b *OrderBuilder) BuildMarketWithContext(ctx context.Context) (*clobtypes.SignableOrder, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if b.tokenID == "" {
		return nil, fmt.Errorf("token_id is required")
	}
	side := strings.ToUpper(strings.TrimSpace(b.side))
	if side != "BUY" && side != "SELL" {
		return nil, fmt.Errorf("side must be BUY or SELL")
	}
	if b.amount == nil {
		return nil, fmt.Errorf("amount is required for market orders")
	}
	if b.amount.value.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	amountScale := decimalPlaces(b.amount.value)
	switch b.amount.kind {
	case amountShares:
		if amountScale > lotSizeScale {
			return nil, fmt.Errorf("amount has too many decimal places (max %d)", lotSizeScale)
		}
	case amountUSDC:
		if amountScale > usdcDecimals {
			return nil, fmt.Errorf("amount has too many decimal places (max %d)", usdcDecimals)
		}
	default:
		return nil, fmt.Errorf("unsupported market order amount")
	}

	orderType := normalizeOrderType(b.orderType, clobtypes.OrderTypeFAK)
	if orderType != clobtypes.OrderTypeFAK && orderType != clobtypes.OrderTypeFOK {
		return nil, fmt.Errorf("market orders require FAK or FOK order type")
	}
	if b.postOnly != nil && *b.postOnly {
		return nil, fmt.Errorf("postOnly is not supported for market orders")
	}

	if side == "SELL" && b.amount.kind == amountUSDC {
		return nil, fmt.Errorf("sell market orders must specify amount in shares")
	}

	tokenIDInt, ok := new(big.Int).SetString(b.tokenID, 10)
	if !ok {
		return nil, fmt.Errorf("invalid token_id format")
	}

	tickSize, err := b.resolveTickSize(ctx, b.tokenID)
	if err != nil {
		return nil, err
	}
	tickScale := decimalPlaces(tickSize)

	var price decimal.Decimal
	if b.price.Sign() < 0 {
		return nil, fmt.Errorf("price must be positive")
	}
	if b.price.Sign() > 0 {
		price = b.price
		if decimalPlaces(price) > tickScale {
			return nil, fmt.Errorf("price has too many decimal places for tick size %s", tickSize.String())
		}
	} else {
		var err error
		price, err = b.resolveMarketPrice(ctx, side, orderType, b.amount)
		if err != nil {
			return nil, err
		}
	}
	price = price.Truncate(tickScale)
	one := decimal.NewFromInt(1)
	if price.LessThan(tickSize) || price.GreaterThan(one.Sub(tickSize)) {
		return nil, fmt.Errorf("price %s is out of bounds for tick size %s", price.String(), tickSize.String())
	}

	feeRateBps, err := b.resolveFeeRateBps(ctx, b.tokenID)
	if err != nil {
		return nil, err
	}

	truncScale := tickScale + lotSizeScale
	rawAmount := b.amount.value
	var makerAmount, takerAmount decimal.Decimal

	switch {
	case side == "BUY" && b.amount.kind == amountUSDC:
		takerAmount = rawAmount.Div(price).Truncate(truncScale)
		makerAmount = rawAmount
	case side == "BUY" && b.amount.kind == amountShares:
		takerAmount = rawAmount
		makerAmount = rawAmount.Mul(price).Truncate(truncScale)
	case side == "SELL" && b.amount.kind == amountShares:
		makerAmount = rawAmount
		takerAmount = rawAmount.Mul(price).Truncate(truncScale)
	default:
		return nil, fmt.Errorf("unsupported market order amount")
	}

	makerFixed := toFixedDecimal(makerAmount)
	takerFixed := toFixedDecimal(takerAmount)

	sigType := int(auth.SignatureEOA)
	if b.signatureType != nil {
		sigType = int(*b.signatureType)
	}

	var maker common.Address
	if b.maker != nil {
		maker = *b.maker
	} else if b.funder != nil {
		if sigType == int(auth.SignatureEOA) {
			return nil, fmt.Errorf("funder requires non-EOA signature type")
		}
		if *b.funder == (common.Address{}) {
			return nil, fmt.Errorf("funder cannot be zero address")
		}
		maker = *b.funder
	} else {
		derived, err := deriveMakerFromSignature(b.signer, sigType)
		if err != nil {
			return nil, err
		}
		maker = derived
	}

	taker := common.HexToAddress("0x0000000000000000000000000000000000000000")
	if b.taker != nil {
		taker = *b.taker
	}

	nonce := big.NewInt(0)
	if b.nonce != nil {
		nonce = b.nonce
	}

	salt, err := b.generateSalt()
	if err != nil {
		return nil, err
	}

	order := &clobtypes.Order{
		Salt:          types.U256{Int: salt},
		Signer:        b.signer.Address(),
		Maker:         maker,
		Taker:         taker,
		TokenID:       types.U256{Int: tokenIDInt},
		MakerAmount:   types.Decimal(makerFixed),
		TakerAmount:   types.Decimal(takerFixed),
		Expiration:    types.U256{Int: big.NewInt(0)},
		Side:          side,
		FeeRateBps:    types.Decimal(decimal.NewFromInt(feeRateBps)),
		Nonce:         types.U256{Int: nonce},
		SignatureType: &sigType,
	}

	return &clobtypes.SignableOrder{
		Order:     order,
		OrderType: orderType,
	}, nil
}

func (b *OrderBuilder) buildLimit(ctx context.Context) (*clobtypes.Order, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if b.tokenID == "" {
		return nil, fmt.Errorf("token_id is required")
	}
	side := strings.ToUpper(strings.TrimSpace(b.side))
	if side != "BUY" && side != "SELL" {
		return nil, fmt.Errorf("side must be BUY or SELL")
	}
	if b.price.Sign() <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}
	if b.size.Sign() <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}

	tokenIDInt, ok := new(big.Int).SetString(b.tokenID, 10)
	if !ok {
		return nil, fmt.Errorf("invalid token_id format")
	}

	tickSize, err := b.resolveTickSize(ctx, b.tokenID)
	if err != nil {
		return nil, err
	}
	tickScale := decimalPlaces(tickSize)

	price := b.price
	if decimalPlaces(price) > tickScale {
		return nil, fmt.Errorf("price has too many decimal places for tick size %s", tickSize.String())
	}
	one := decimal.NewFromInt(1)
	if price.LessThan(tickSize) || price.GreaterThan(one.Sub(tickSize)) {
		return nil, fmt.Errorf("price %s is out of bounds for tick size %s", price.String(), tickSize.String())
	}

	size := b.size
	if decimalPlaces(size) > lotSizeScale {
		return nil, fmt.Errorf("size has too many decimal places (max %d)", lotSizeScale)
	}
	if size.Sign() <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}

	feeRateBps, err := b.resolveFeeRateBps(ctx, b.tokenID)
	if err != nil {
		return nil, err
	}

	truncScale := tickScale + lotSizeScale
	var makerAmount, takerAmount decimal.Decimal
	if side == "BUY" {
		takerAmount = size
		makerAmount = size.Mul(price).Truncate(truncScale)
	} else {
		makerAmount = size
		takerAmount = size.Mul(price).Truncate(truncScale)
	}

	makerFixed := toFixedDecimal(makerAmount)
	takerFixed := toFixedDecimal(takerAmount)

	sigType := int(auth.SignatureEOA)
	if b.signatureType != nil {
		sigType = int(*b.signatureType)
	}

	var maker common.Address
	if b.maker != nil {
		maker = *b.maker
	} else if b.funder != nil {
		if sigType == int(auth.SignatureEOA) {
			return nil, fmt.Errorf("funder requires non-EOA signature type")
		}
		if *b.funder == (common.Address{}) {
			return nil, fmt.Errorf("funder cannot be zero address")
		}
		maker = *b.funder
	} else {
		derived, err := deriveMakerFromSignature(b.signer, sigType)
		if err != nil {
			return nil, err
		}
		maker = derived
	}

	taker := common.HexToAddress("0x0000000000000000000000000000000000000000")
	if b.taker != nil {
		taker = *b.taker
	}

	nonce := big.NewInt(0)
	if b.nonce != nil {
		nonce = b.nonce
	}

	salt, err := b.generateSalt()
	if err != nil {
		return nil, err
	}

	expiration := big.NewInt(0)
	if b.expiration != nil {
		if b.expiration.Sign() < 0 {
			return nil, fmt.Errorf("expiration must be non-negative")
		}
		expiration = b.expiration
	}

	return &clobtypes.Order{
		Salt:          types.U256{Int: salt},
		Signer:        b.signer.Address(),
		Maker:         maker,
		Taker:         taker,
		TokenID:       types.U256{Int: tokenIDInt},
		MakerAmount:   types.Decimal(makerFixed),
		TakerAmount:   types.Decimal(takerFixed),
		Expiration:    types.U256{Int: expiration},
		Side:          side,
		FeeRateBps:    types.Decimal(decimal.NewFromInt(feeRateBps)),
		Nonce:         types.U256{Int: nonce},
		SignatureType: &sigType,
	}, nil
}
