package clob

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func (b *OrderBuilder) resolveTickSize(ctx context.Context, tokenID string) (decimal.Decimal, error) {
	var override *decimal.Decimal
	if b.tickSize != 0 {
		parsed := decimal.NewFromFloat(b.tickSize)
		override = &parsed
	}

	hasClient := clientHasTransport(b.client)
	if hasClient {
		resp, err := b.client.TickSize(ctx, &clobtypes.TickSizeRequest{TokenID: tokenID})
		if err != nil {
			if override != nil {
				return *override, nil
			}
			return decimal.Decimal{}, fmt.Errorf("tick size lookup failed: %w", err)
		}
		minTick := decimal.NewFromFloat(resp.MinimumTickSize)

		if override != nil {
			if override.Cmp(minTick) < 0 {
				return decimal.Decimal{}, fmt.Errorf("tick size %s is smaller than minimum %s", override.String(), minTick.String())
			}
			return *override, nil
		}
		return minTick, nil
	}

	if override != nil {
		return *override, nil
	}
	return decimal.Decimal{}, fmt.Errorf("tick size is required (set TickSize or provide a client)")
}

func (b *OrderBuilder) resolveFeeRateBps(ctx context.Context, tokenID string) (int64, error) {
	userFee, err := parseFeeRateBps(b.feeRateBps)
	if err != nil {
		return 0, err
	}

	if !clientHasTransport(b.client) {
		return userFee, nil
	}

	resp, err := b.client.FeeRate(ctx, &clobtypes.FeeRateRequest{TokenID: tokenID})
	if err != nil {
		if userFee > 0 {
			return userFee, nil
		}
		return 0, fmt.Errorf("fee rate lookup failed: %w", err)
	}

	marketFee := int64(resp.BaseFee)
	if marketFee == 0 && resp.FeeRate != "" {
		parsed, err := decimal.NewFromString(resp.FeeRate)
		if err != nil {
			return 0, fmt.Errorf("invalid fee rate response: %w", err)
		}
		marketFee = parsed.IntPart()
	}

	if marketFee > 0 && userFee > 0 && userFee != marketFee {
		return 0, fmt.Errorf("invalid fee rate %d, market fee rate is %d", userFee, marketFee)
	}
	if marketFee > 0 {
		return marketFee, nil
	}
	return userFee, nil
}

func (b *OrderBuilder) resolveMarketPrice(ctx context.Context, side string, orderType clobtypes.OrderType, amount *marketAmount) (decimal.Decimal, error) {
	if amount == nil {
		return decimal.Decimal{}, fmt.Errorf("amount is required")
	}
	if b.client == nil || !clientHasTransport(b.client) {
		return decimal.Decimal{}, fmt.Errorf("client is required to fetch order book")
	}
	book, err := b.client.OrderBook(ctx, &clobtypes.BookRequest{TokenID: b.tokenID})
	if err != nil {
		return decimal.Decimal{}, err
	}

	var levels []clobtypes.PriceLevel
	switch side {
	case "BUY":
		levels = book.Asks
	case "SELL":
		levels = book.Bids
	default:
		return decimal.Decimal{}, fmt.Errorf("invalid side %q", side)
	}

	if len(levels) == 0 {
		return decimal.Decimal{}, fmt.Errorf("no opposing orders")
	}

	firstPrice, err := decimal.NewFromString(levels[0].Price)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid price level: %w", err)
	}

	sum := decimal.Zero
	var cutoff *decimal.Decimal
	for i := len(levels) - 1; i >= 0; i-- {
		level := levels[i]
		levelPrice, err := decimal.NewFromString(level.Price)
		if err != nil {
			return decimal.Decimal{}, fmt.Errorf("invalid price level: %w", err)
		}
		levelSize, err := decimal.NewFromString(level.Size)
		if err != nil {
			return decimal.Decimal{}, fmt.Errorf("invalid size level: %w", err)
		}

		if amount.kind == amountUSDC {
			sum = sum.Add(levelSize.Mul(levelPrice))
		} else {
			sum = sum.Add(levelSize)
		}

		if sum.GreaterThanOrEqual(amount.value) {
			cutoff = &levelPrice
			break
		}
	}

	if cutoff != nil {
		return *cutoff, nil
	}
	if orderType == clobtypes.OrderTypeFOK {
		return decimal.Decimal{}, fmt.Errorf("insufficient liquidity to fill order")
	}
	return firstPrice, nil
}

func clientHasTransport(client Client) bool {
	if client == nil {
		return false
	}
	if impl, ok := client.(*clientImpl); ok {
		if impl == nil {
			return false
		}
		return impl.httpClient != nil
	}
	return true
}

func decimalPlaces(d decimal.Decimal) int32 {
	exp := d.Exponent()
	if exp < 0 {
		return -exp
	}
	return 0
}

func toFixedDecimal(d decimal.Decimal) decimal.Decimal {
	trimmed := d.Truncate(usdcDecimals)
	return trimmed.Shift(usdcDecimals).Truncate(0)
}

func parseFeeRateBps(dec decimal.Decimal) (int64, error) {
	if dec.Sign() <= 0 {
		return 0, nil
	}
	intPart := dec.Truncate(0)
	if !intPart.Equal(dec) {
		return 0, fmt.Errorf("fee rate must be an integer bps value")
	}
	return intPart.IntPart(), nil
}

func generateSalt() (*big.Int, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}
	raw := binary.BigEndian.Uint64(buf[:])
	raw &= (1 << 53) - 1
	return new(big.Int).SetUint64(raw), nil
}

func (b *OrderBuilder) generateSalt() (*big.Int, error) {
	if b.saltGenerator != nil {
		return b.saltGenerator()
	}
	return generateSalt()
}

func deriveMakerFromSignature(signer auth.Signer, sigType int) (common.Address, error) {
	if signer == nil {
		return common.Address{}, fmt.Errorf("signer is required")
	}
	chainID := int64(0)
	if signer.ChainID() != nil {
		chainID = signer.ChainID().Int64()
	}
	switch sigType {
	case int(auth.SignatureProxy):
		proxy, err := auth.DeriveProxyWalletForChain(signer.Address(), chainID)
		if err != nil && chainID == 0 {
			proxy, err = auth.DeriveProxyWallet(signer.Address())
		}
		if err != nil {
			return common.Address{}, fmt.Errorf("failed to derive proxy wallet: %w", err)
		}
		return proxy, nil
	case int(auth.SignatureGnosisSafe):
		safe, err := auth.DeriveSafeWalletForChain(signer.Address(), chainID)
		if err != nil && chainID == 0 {
			safe, err = auth.DeriveSafeWallet(signer.Address())
		}
		if err != nil {
			return common.Address{}, fmt.Errorf("failed to derive safe wallet: %w", err)
		}
		return safe, nil
	default:
		return signer.Address(), nil
	}
}

// UseProxy sets the order to use the user's Proxy Wallet.
func (b *OrderBuilder) UseProxy() *OrderBuilder {
	t := auth.SignatureProxy
	b.signatureType = &t
	return b
}

// UseSafe sets the order to use the user's Gnosis Safe.
func (b *OrderBuilder) UseSafe() *OrderBuilder {
	t := auth.SignatureGnosisSafe
	b.signatureType = &t
	return b
}
