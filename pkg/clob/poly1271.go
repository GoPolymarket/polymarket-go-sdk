package clob

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
)

const (
	poly1271ExchangeV2Address     = "0xE111180000d2663C0091e4f400237545B87B996B"
	poly1271Bytes32Zero           = "0x0000000000000000000000000000000000000000000000000000000000000000"
	poly1271EIP712DomainType      = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	poly1271OrderType             = "Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"
	poly1271TypedDataSignType     = "TypedDataSign(Order contents,string name,string version,uint256 chainId,address verifyingContract,bytes32 salt)" + poly1271OrderType
	poly1271ExchangeDomainName    = "Polymarket CTF Exchange"
	poly1271ExchangeDomainVersion = "2"
	poly1271DepositWalletName     = "DepositWallet"
	poly1271DepositWalletVersion  = "1"
)

type digestSigner interface {
	SignDigest([]byte) ([]byte, error)
}

type poly1271OrderForHash struct {
	Salt          *big.Int
	Maker         common.Address
	Signer        common.Address
	TokenID       *big.Int
	MakerAmount   *big.Int
	TakerAmount   *big.Int
	Side          *big.Int
	SignatureType *big.Int
	Timestamp     *big.Int
	Metadata      string
	Builder       string
}

func signPoly1271Order(signer auth.Signer, order *clobtypes.Order) (string, error) {
	digestSigner, ok := signer.(digestSigner)
	if !ok {
		return "", fmt.Errorf("POLY_1271 signing requires a signer that can sign raw digests")
	}

	side, err := poly1271Side(order.Side)
	if err != nil {
		return "", err
	}
	sigType := int(auth.SignaturePoly1271)
	if order.SignatureType != nil {
		sigType = *order.SignatureType
	}

	orderForHash := poly1271OrderForHash{
		Salt:          order.Salt.Int,
		Maker:         order.Maker,
		Signer:        order.Signer,
		TokenID:       order.TokenID.Int,
		MakerAmount:   order.MakerAmount.BigInt(),
		TakerAmount:   order.TakerAmount.BigInt(),
		Side:          big.NewInt(int64(side)),
		SignatureType: big.NewInt(int64(sigType)),
		Timestamp:     big.NewInt(order.Timestamp),
		Metadata:      padBytes32(order.Metadata),
		Builder:       padBytes32(order.Builder),
	}

	domainSeparator := poly1271ExchangeDomainSeparator(common.HexToAddress(poly1271ExchangeV2Address), signer.ChainID().Int64())
	contentsHash, err := poly1271OrderStructHash(orderForHash)
	if err != nil {
		return "", err
	}
	walletSalt, err := poly1271ABIHexBytes32(poly1271Bytes32Zero)
	if err != nil {
		return "", err
	}
	typedDataSignHash := crypto.Keccak256(
		poly1271ABIBytes32(poly1271TypedDataSignType),
		contentsHash,
		poly1271ABIBytes32(poly1271DepositWalletName),
		poly1271ABIBytes32(poly1271DepositWalletVersion),
		poly1271ABIUint(signer.ChainID()),
		poly1271ABIAddress(order.Signer),
		walletSalt,
	)
	digest := crypto.Keccak256(append(append([]byte{0x19, 0x01}, domainSeparator...), typedDataSignHash...))
	innerSignature, err := digestSigner.SignDigest(digest)
	if err != nil {
		return "", err
	}
	return wrapPoly1271Signature(innerSignature, domainSeparator, contentsHash), nil
}

func poly1271ExchangeDomainSeparator(exchange common.Address, chainID int64) []byte {
	return crypto.Keccak256(
		poly1271ABIBytes32(poly1271EIP712DomainType),
		poly1271ABIBytes32(poly1271ExchangeDomainName),
		poly1271ABIBytes32(poly1271ExchangeDomainVersion),
		poly1271ABIUint(big.NewInt(chainID)),
		poly1271ABIAddress(exchange),
	)
}

func poly1271OrderStructHash(order poly1271OrderForHash) ([]byte, error) {
	metadata, err := poly1271ABIHexBytes32(order.Metadata)
	if err != nil {
		return nil, err
	}
	builder, err := poly1271ABIHexBytes32(order.Builder)
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(
		poly1271ABIBytes32(poly1271OrderType),
		poly1271ABIUint(order.Salt),
		poly1271ABIAddress(order.Maker),
		poly1271ABIAddress(order.Signer),
		poly1271ABIUint(order.TokenID),
		poly1271ABIUint(order.MakerAmount),
		poly1271ABIUint(order.TakerAmount),
		poly1271ABIUint(order.Side),
		poly1271ABIUint(order.SignatureType),
		poly1271ABIUint(order.Timestamp),
		metadata,
		builder,
	), nil
}

func wrapPoly1271Signature(innerSignature, domainSeparator, contentsHash []byte) string {
	out := make([]byte, 0, len(innerSignature)+len(domainSeparator)+len(contentsHash)+len(poly1271OrderType)+2)
	out = append(out, innerSignature...)
	out = append(out, domainSeparator...)
	out = append(out, contentsHash...)
	out = append(out, []byte(poly1271OrderType)...)
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(poly1271OrderType)))
	out = append(out, length...)
	return "0x" + hex.EncodeToString(out)
}

func poly1271ABIBytes32(value string) []byte {
	return crypto.Keccak256([]byte(value))
}

func poly1271ABIUint(value *big.Int) []byte {
	return common.LeftPadBytes(value.Bytes(), 32)
}

func poly1271ABIAddress(value common.Address) []byte {
	return common.LeftPadBytes(value.Bytes(), 32)
}

func poly1271ABIHexBytes32(value string) ([]byte, error) {
	raw, err := hex.DecodeString(strings.TrimPrefix(value, "0x"))
	if err != nil {
		return nil, err
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("expected bytes32, got %d bytes", len(raw))
	}
	return raw, nil
}

func poly1271Side(side string) (int, error) {
	switch strings.ToUpper(strings.TrimSpace(side)) {
	case "BUY":
		return 0, nil
	case "SELL":
		return 1, nil
	default:
		return 0, fmt.Errorf("order side must be BUY or SELL, got %q", side)
	}
}
