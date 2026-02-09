// Package auth provides cryptographic primitives for Polymarket authentication.
// It supports L1 (EIP-712) signing for account management and L2 (HMAC) signing
// for high-frequency trading operations. It also includes utilities for
// deterministic wallet derivation (Proxy and Gnosis Safe).
package auth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/pkg/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// ClobAuthDomain is the EIP-712 domain used for CLOB authentication requests.
var ClobAuthDomain = &apitypes.TypedDataDomain{
	Name:    "ClobAuthDomain",
	Version: "1",
}

// ClobAuthTypes defines the EIP-712 type schemas for CLOB authentication.
var ClobAuthTypes = apitypes.Types{
	"EIP712Domain": {
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
	},
	"ClobAuth": {
		{Name: "address", Type: "address"},
		{Name: "timestamp", Type: "string"},
		{Name: "nonce", Type: "uint256"},
		{Name: "message", Type: "string"},
	},
}

// APIKey represents the Layer 2 credentials used for HMAC-signed requests.
// These are typically created or derived using a Layer 1 (EIP-712) signature.
type APIKey struct {
	Key        string
	Secret     string
	Passphrase string
}

// Signer defines the interface for an EIP-712 capable signing entity.
// It can be implemented by a local private key, a hardware wallet, or a KMS.
type Signer interface {
	// Address returns the Ethereum address of the signer.
	Address() common.Address
	// ChainID returns the network ID this signer is configured for.
	ChainID() *big.Int
	// SignTypedData signs data according to the EIP-712 specification.
	SignTypedData(domain *apitypes.TypedDataDomain, types apitypes.Types, message apitypes.TypedDataMessage, primaryType string) ([]byte, error)
}

// SignatureType indicates the wallet type used for signature verification on the CLOB.
type SignatureType int

const (
	// SignatureEOA indicates a standard Externally Owned Account (EOA).
	SignatureEOA SignatureType = 0
	// SignatureProxy indicates a signature from a Polymarket Proxy wallet.
	// Magic.link wallets use the same proxy factory, so SignatureProxy covers both.
	SignatureProxy SignatureType = 1
	// SignatureGnosisSafe indicates a signature from a Gnosis Safe multisig.
	SignatureGnosisSafe SignatureType = 2
)

// Supported chain IDs for Polymarket operations.
const (
	PolygonChainID int64 = 137
	AmoyChainID    int64 = 80002
)

// Constants for Proxy Wallet Derivation.
const (
	// ProxyFactoryAddress is the factory contract for Polymarket Proxy wallets (Magic/email).
	ProxyFactoryAddress = "0xaB45c5A4B0c941a2F231C04C3f49182e1A254052"
	// SafeFactoryAddress is the factory contract for Gnosis Safe wallets.
	SafeFactoryAddress = "0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b"

	// ProxyInitCodeHash is the init code hash for Polymarket Proxy wallets.
	ProxyInitCodeHash = "0xd21df8dc65880a8606f09fe0ce3df9b8869287ab0b058be05aa9e8af6330a00b"
	// SafeInitCodeHash is the init code hash for Gnosis Safe wallets.
	SafeInitCodeHash = "0x2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf"
)

type walletConfig struct {
	ProxyFactory *common.Address
	SafeFactory  common.Address
}

var walletConfigs = map[int64]walletConfig{
	PolygonChainID: {
		ProxyFactory: ptrAddress(common.HexToAddress(ProxyFactoryAddress)),
		SafeFactory:  common.HexToAddress(SafeFactoryAddress),
	},
	AmoyChainID: {
		ProxyFactory: nil,
		SafeFactory:  common.HexToAddress(SafeFactoryAddress),
	},
}

var (
	// Use unified error definitions from pkg/errors
	ErrMissingSigner          = sdkerrors.ErrMissingSigner
	ErrMissingCreds           = sdkerrors.ErrMissingCreds
	ErrMissingBuilderConfig   = sdkerrors.ErrMissingBuilderConfig
	ErrProxyWalletUnsupported = sdkerrors.ErrProxyWalletUnsupported
	ErrSafeWalletUnsupported  = sdkerrors.ErrSafeWalletUnsupported
)

// Authentication header keys used by Polymarket API.
const (
	HeaderPolyAddress           = "POLY_ADDRESS"
	HeaderPolySignature         = "POLY_SIGNATURE"
	HeaderPolyTimestamp         = "POLY_TIMESTAMP"
	HeaderPolyNonce             = "POLY_NONCE"
	HeaderPolyAPIKey            = "POLY_API_KEY"
	HeaderPolyPassphrase        = "POLY_PASSPHRASE"
	HeaderPolyBuilderAPIKey     = "POLY_BUILDER_API_KEY"
	HeaderPolyBuilderPassphrase = "POLY_BUILDER_PASSPHRASE"
	HeaderPolyBuilderSignature  = "POLY_BUILDER_SIGNATURE"
	HeaderPolyBuilderTimestamp  = "POLY_BUILDER_TIMESTAMP"
)

// PrivateKeySigner implements the Signer interface using a local ECDSA private key.
type PrivateKeySigner struct {
	key     *ecdsa.PrivateKey
	address common.Address
	chainID *big.Int
}

// NewPrivateKeySigner creates a new signer from a hex-encoded private key.
// The hexKey may optionally include a "0x" prefix.
func NewPrivateKeySigner(hexKey string, chainID int64) (*PrivateKeySigner, error) {
	if len(hexKey) > 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return &PrivateKeySigner{
		key:     key,
		address: crypto.PubkeyToAddress(key.PublicKey),
		chainID: big.NewInt(chainID),
	}, nil
}

// Address returns the public address of the private key.
func (s *PrivateKeySigner) Address() common.Address {
	return s.address
}

// ChainID returns the network ID this signer is configured for.
func (s *PrivateKeySigner) ChainID() *big.Int {
	return s.chainID
}

// BuildL1Headers creates the L1 authentication headers required for API key management.
// It generates an EIP-712 signature over a standard authentication message.
func BuildL1Headers(signer Signer, timestamp int64, nonce int64) (http.Header, error) {
	if signer == nil {
		return nil, ErrMissingSigner
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	domain := &apitypes.TypedDataDomain{
		Name:    ClobAuthDomain.Name,
		Version: ClobAuthDomain.Version,
		ChainId: (*math.HexOrDecimal256)(signer.ChainID()),
	}

	message := apitypes.TypedDataMessage{
		"address":   signer.Address().Hex(),
		"timestamp": fmt.Sprintf("%d", timestamp),
		"nonce":     (*math.HexOrDecimal256)(big.NewInt(nonce)),
		"message":   "This message attests that I control the given wallet",
	}

	sig, err := signer.SignTypedData(domain, ClobAuthTypes, message, "ClobAuth")
	if err != nil {
		return nil, fmt.Errorf("sign clob auth: %w", err)
	}

	headers := http.Header{}
	headers.Set(HeaderPolyAddress, signer.Address().Hex())
	headers.Set(HeaderPolySignature, hexutil.Encode(sig))
	headers.Set(HeaderPolyTimestamp, fmt.Sprintf("%d", timestamp))
	headers.Set(HeaderPolyNonce, fmt.Sprintf("%d", nonce))
	return headers, nil
}

// SignHMAC calculates the HMAC-SHA256 signature used for Layer 2 authentication.
// The message is typically constructed as timestamp + method + path + body.
func SignHMAC(secret string, message string) (string, error) {
	decodedSecret, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, decodedSecret)
	h.Write([]byte(message))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func decodeSecret(secret string) ([]byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(secret)
	if err == nil {
		return decoded, nil
	}
	decoded, err = base64.RawURLEncoding.DecodeString(secret)
	if err == nil {
		return decoded, nil
	}
	decoded, err = base64.StdEncoding.DecodeString(secret)
	if err == nil {
		return decoded, nil
	}
	decoded, err = base64.RawStdEncoding.DecodeString(secret)
	if err == nil {
		return decoded, nil
	}
	return nil, fmt.Errorf("invalid base64 secret: %w", err)
}

// BuildL2Headers returns the headers required for an HMAC-authenticated L2 request.
func BuildL2Headers(signer Signer, apiKey *APIKey, method, path string, body *string, timestamp int64) (http.Header, error) {
	if signer == nil {
		return nil, ErrMissingSigner
	}
	if apiKey == nil {
		return nil, ErrMissingCreds
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	message := fmt.Sprintf("%d%s%s", timestamp, method, path)
	if body != nil && *body != "" {
		message += strings.ReplaceAll(*body, "'", "\"")
	}

	sig, err := SignHMAC(apiKey.Secret, message)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set(HeaderPolyAddress, signer.Address().Hex())
	headers.Set(HeaderPolyAPIKey, apiKey.Key)
	headers.Set(HeaderPolyPassphrase, apiKey.Passphrase)
	headers.Set(HeaderPolyTimestamp, fmt.Sprintf("%d", timestamp))
	headers.Set(HeaderPolySignature, sig)
	return headers, nil
}

// BuilderCredentials represents credentials for a Builder account.
type BuilderCredentials struct {
	Key        string
	Secret     string
	Passphrase string
}

// BuilderHTTPDoer defines an interface for executing HTTP requests for remote signing.
type BuilderHTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// BuilderRemoteConfig configures a remote signing service for builder attribution.
type BuilderRemoteConfig struct {
	// Host is the endpoint of the remote signing service.
	Host string
	// Token is an optional bearer token for authenticating with the signer.
	Token string
	// HTTPClient allows providing a custom client for signing requests.
	HTTPClient BuilderHTTPDoer
}

// BuilderConfig holds configuration for either local or remote builder attribution.
type BuilderConfig struct {
	Local  *BuilderCredentials
	Remote *BuilderRemoteConfig
}

// IsValid returns true if the configuration has sufficient credentials.
func (c *BuilderConfig) IsValid() bool {
	if c == nil {
		return false
	}
	if c.Local != nil {
		return c.Local.Key != "" && c.Local.Secret != "" && c.Local.Passphrase != ""
	}
	if c.Remote != nil {
		return c.Remote.Host != ""
	}
	return false
}

// Headers returns the attribution headers for a given request.
func (c *BuilderConfig) Headers(ctx context.Context, method, path string, body *string, timestamp int64) (http.Header, error) {
	if c == nil {
		return nil, ErrMissingBuilderConfig
	}
	if c.Local != nil {
		return buildBuilderHeadersLocal(c.Local, method, path, body, timestamp)
	}
	if c.Remote != nil {
		return buildBuilderHeadersRemote(ctx, c.Remote, method, path, body, timestamp)
	}
	return nil, ErrMissingBuilderConfig
}

func buildBuilderHeadersLocal(creds *BuilderCredentials, method, path string, body *string, timestamp int64) (http.Header, error) {
	if creds == nil {
		return nil, ErrMissingBuilderConfig
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	message := fmt.Sprintf("%d%s%s", timestamp, method, path)
	if body != nil && *body != "" {
		message += strings.ReplaceAll(*body, "'", "\"")
	}
	sig, err := SignHMAC(creds.Secret, message)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set(HeaderPolyBuilderAPIKey, creds.Key)
	headers.Set(HeaderPolyBuilderPassphrase, creds.Passphrase)
	headers.Set(HeaderPolyBuilderTimestamp, fmt.Sprintf("%d", timestamp))
	headers.Set(HeaderPolyBuilderSignature, sig)
	return headers, nil
}

func buildBuilderHeadersRemote(ctx context.Context, remote *BuilderRemoteConfig, method, path string, body *string, timestamp int64) (http.Header, error) {
	if remote == nil || remote.Host == "" {
		return nil, ErrMissingBuilderConfig
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	payload := map[string]interface{}{
		"method":    method,
		"path":      path,
		"body":      "",
		"timestamp": timestamp,
	}
	if body != nil {
		payload["body"] = *body
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal builder payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, remote.Host, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("builder request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if remote.Token != "" {
		req.Header.Set("Authorization", "Bearer "+remote.Token)
	}

	client := remote.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("builder request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("builder signer error: status %d", resp.StatusCode)
	}

	var rawHeaders map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&rawHeaders); err != nil {
		return nil, fmt.Errorf("decode builder headers: %w", err)
	}

	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := rawHeaders[k]; ok && v != "" {
				return v
			}
		}
		return ""
	}

	builderKey := get(HeaderPolyBuilderAPIKey, "poly_builder_api_key", "POLY_BUILDER_API_KEY")
	builderPass := get(HeaderPolyBuilderPassphrase, "poly_builder_passphrase", "POLY_BUILDER_PASSPHRASE")
	builderSig := get(HeaderPolyBuilderSignature, "poly_builder_signature", "POLY_BUILDER_SIGNATURE")
	builderTs := get(HeaderPolyBuilderTimestamp, "poly_builder_timestamp", "POLY_BUILDER_TIMESTAMP")

	if builderKey == "" || builderPass == "" || builderSig == "" || builderTs == "" {
		return nil, fmt.Errorf("invalid builder headers response")
	}

	headers := http.Header{}
	headers.Set(HeaderPolyBuilderAPIKey, builderKey)
	headers.Set(HeaderPolyBuilderPassphrase, builderPass)
	headers.Set(HeaderPolyBuilderSignature, builderSig)
	headers.Set(HeaderPolyBuilderTimestamp, builderTs)
	return headers, nil
}

// DeriveProxyWallet calculates the deterministic Proxy Wallet address for an EOA.
// Corresponds to the `derive_proxy_wallet` logic in official clients.
// Defaults to Polygon Mainnet.
func DeriveProxyWallet(eoa common.Address) (common.Address, error) {
	return DeriveProxyWalletForChain(eoa, PolygonChainID)
}

// DeriveProxyWalletForChain calculates the deterministic Proxy Wallet address for an EOA on a specific chain.
func DeriveProxyWalletForChain(eoa common.Address, chainID int64) (common.Address, error) {
	cfg, ok := walletConfigs[chainID]
	if !ok || cfg.ProxyFactory == nil {
		return common.Address{}, ErrProxyWalletUnsupported
	}
	// salt = keccak256(eoa)
	salt := crypto.Keccak256(eoa.Bytes())

	initCodeHash, err := hexutil.Decode(ProxyInitCodeHash)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid proxy init code hash: %w", err)
	}

	// address = keccak256(0xff + factory + salt + initCodeHash)[12:]
	address := crypto.CreateAddress2(*cfg.ProxyFactory, common.BytesToHash(salt), initCodeHash)

	return address, nil
}

// DeriveSafeWallet calculates the deterministic Gnosis Safe address for an EOA.
// Corresponds to the `derive_safe_wallet` logic in official clients.
// Defaults to Polygon Mainnet.
func DeriveSafeWallet(eoa common.Address) (common.Address, error) {
	return DeriveSafeWalletForChain(eoa, PolygonChainID)
}

// DeriveSafeWalletForChain calculates the deterministic Gnosis Safe address for an EOA on a specific chain.
func DeriveSafeWalletForChain(eoa common.Address, chainID int64) (common.Address, error) {
	cfg, ok := walletConfigs[chainID]
	if !ok {
		return common.Address{}, ErrSafeWalletUnsupported
	}
	// salt = keccak256(left_pad_32(eoa))
	paddedEOA := common.LeftPadBytes(eoa.Bytes(), 32)
	salt := crypto.Keccak256(paddedEOA)

	initCodeHash, err := hexutil.Decode(SafeInitCodeHash)
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid safe init code hash: %w", err)
	}

	// address = keccak256(0xff + factory + salt + initCodeHash)[12:]
	address := crypto.CreateAddress2(cfg.SafeFactory, common.BytesToHash(salt), initCodeHash)

	return address, nil
}

func ptrAddress(addr common.Address) *common.Address {
	return &addr
}

// SignTypedData signs EIP-712 typed data. It ensures the V value is correctly adjusted
// for compatibility with Ethereum's expected 27/28 values.
func (s *PrivateKeySigner) SignTypedData(domain *apitypes.TypedDataDomain, types apitypes.Types, message apitypes.TypedDataMessage, primaryType string) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types:       types,
		PrimaryType: primaryType,
		Domain:      *domain,
		Message:     message,
	}

	sighash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash typed data: %w", err)
	}

	signature, err := crypto.Sign(sighash, s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}

	if signature[64] < 27 {
		signature[64] += 27
	}

	return signature, nil
}
