// Package errors provides unified error definitions for the SDK.
// All errors are defined in this package with error codes for easy identification.
package errors

import (
	"fmt"
)

// ErrorCode represents a unique error code for each error type.
type ErrorCode string

// Error codes for all SDK errors
const (
	// Authentication and Authorization error codes (AUTH-xxx)
	CodeMissingSigner        ErrorCode = "AUTH-001"
	CodeMissingCreds         ErrorCode = "AUTH-002"
	CodeMissingBuilderConfig ErrorCode = "AUTH-003"
	CodeInvalidSignature     ErrorCode = "AUTH-004"
	CodeUnauthorized         ErrorCode = "AUTH-005"

	// Wallet derivation error codes (WALLET-xxx)
	CodeProxyWalletUnsupported ErrorCode = "WALLET-001"
	CodeSafeWalletUnsupported  ErrorCode = "WALLET-002"

	// CLOB API error codes (CLOB-xxx)
	CodeInsufficientFunds ErrorCode = "CLOB-001"
	CodeRateLimitExceeded ErrorCode = "CLOB-002"
	CodeOrderNotFound     ErrorCode = "CLOB-003"
	CodeMarketClosed      ErrorCode = "CLOB-004"
	CodeGeoblocked        ErrorCode = "CLOB-005"
	CodeInvalidPrice      ErrorCode = "CLOB-006"
	CodeInvalidSize       ErrorCode = "CLOB-007"
	CodeBatchSizeExceeded ErrorCode = "CLOB-008"

	// HTTP and Network error codes (NET-xxx)
	CodeInternalServerError ErrorCode = "NET-001"
	CodeBadRequest          ErrorCode = "NET-002"
	CodeCircuitOpen         ErrorCode = "NET-003"
	CodeTooManyRequests     ErrorCode = "NET-004"

	// Data API error codes (DATA-xxx)
	CodeMissingRequest      ErrorCode = "DATA-001"
	CodeMissingUser         ErrorCode = "DATA-002"
	CodeInvalidMarketFilter ErrorCode = "DATA-003"
	CodeInvalidTradeFilter  ErrorCode = "DATA-004"

	// WebSocket error codes (WS-xxx)
	CodeInvalidSubscription ErrorCode = "WS-001"

	// CTF (Conditional Token Framework) error codes (CTF-xxx)
	CodeMissingU256Value  ErrorCode = "CTF-001"
	CodeMissingBackend    ErrorCode = "CTF-002"
	CodeMissingTransactor ErrorCode = "CTF-003"
	CodeNegRiskAdapter    ErrorCode = "CTF-004"
	CodeConfigNotFound    ErrorCode = "CTF-005"

	// Bridge error codes (BRIDGE-xxx)
	CodeMissingFromAddress     ErrorCode = "BRIDGE-001"
	CodeMissingDepositAddress  ErrorCode = "BRIDGE-002"
	CodeWithdrawUnsupported    ErrorCode = "BRIDGE-003"
	CodeMissingWithdrawRequest ErrorCode = "BRIDGE-004"
	CodeMissingWithdrawAddress ErrorCode = "BRIDGE-005"
)

// SDKError represents a structured error with code and message.
type SDKError struct {
	Code    ErrorCode
	Message string
}

// Error implements the error interface.
func (e *SDKError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Is implements error comparison for errors.Is.
func (e *SDKError) Is(target error) bool {
	t, ok := target.(*SDKError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New creates a new SDKError with the given code and message.
func New(code ErrorCode, message string) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
	}
}

// Authentication and Authorization errors
var (
	// ErrMissingSigner is returned when a signer is required but not provided.
	ErrMissingSigner = New(CodeMissingSigner, "signer is required")
	// ErrMissingCreds is returned when API credentials are required but not provided.
	ErrMissingCreds = New(CodeMissingCreds, "api credentials are required")
	// ErrMissingBuilderConfig is returned when builder config is required but not provided.
	ErrMissingBuilderConfig = New(CodeMissingBuilderConfig, "builder config is required")
	// ErrInvalidSignature is returned when a signature is invalid.
	ErrInvalidSignature = New(CodeInvalidSignature, "invalid signature")
	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = New(CodeUnauthorized, "unauthorized")
)

// Wallet derivation errors
var (
	// ErrProxyWalletUnsupported is returned when proxy wallet derivation is not supported on the chain.
	ErrProxyWalletUnsupported = New(CodeProxyWalletUnsupported, "proxy wallet derivation not supported on this chain")
	// ErrSafeWalletUnsupported is returned when safe wallet derivation is not supported on the chain.
	ErrSafeWalletUnsupported = New(CodeSafeWalletUnsupported, "safe wallet derivation not supported on this chain")
)

// CLOB API errors
var (
	// ErrInsufficientFunds is returned when the user has insufficient funds.
	ErrInsufficientFunds = New(CodeInsufficientFunds, "insufficient funds")
	// ErrRateLimitExceeded is returned when the rate limit is exceeded.
	ErrRateLimitExceeded = New(CodeRateLimitExceeded, "rate limit exceeded")
	// ErrOrderNotFound is returned when an order is not found.
	ErrOrderNotFound = New(CodeOrderNotFound, "order not found")
	// ErrMarketClosed is returned when a market is closed.
	ErrMarketClosed = New(CodeMarketClosed, "market closed")
	// ErrGeoblocked is returned when the user is geoblocked.
	ErrGeoblocked = New(CodeGeoblocked, "geoblocked")
	// ErrInvalidPrice is returned when a price is invalid.
	ErrInvalidPrice = New(CodeInvalidPrice, "invalid price")
	// ErrInvalidSize is returned when a size is invalid.
	ErrInvalidSize = New(CodeInvalidSize, "invalid size")
	// ErrBatchSizeExceeded is returned when a batch request exceeds the maximum allowed size.
	ErrBatchSizeExceeded = New(CodeBatchSizeExceeded, "batch size exceeds maximum")
)

// HTTP and Network errors
var (
	// ErrInternalServerError is returned when the server returns a 5xx error.
	ErrInternalServerError = New(CodeInternalServerError, "internal server error")
	// ErrBadRequest is returned when the request is malformed.
	ErrBadRequest = New(CodeBadRequest, "bad request")
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = New(CodeCircuitOpen, "circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests are made in half-open state.
	ErrTooManyRequests = New(CodeTooManyRequests, "too many requests in half-open state")
)

// Data API errors
var (
	// ErrMissingRequest is returned when a request is required but not provided.
	ErrMissingRequest = New(CodeMissingRequest, "request is required")
	// ErrMissingUser is returned when a user is required but not provided.
	ErrMissingUser = New(CodeMissingUser, "user is required")
	// ErrInvalidMarketFilter is returned when market filter is invalid.
	ErrInvalidMarketFilter = New(CodeInvalidMarketFilter, "market filter cannot include both markets and event IDs")
	// ErrInvalidTradeFilter is returned when trade filter is invalid.
	ErrInvalidTradeFilter = New(CodeInvalidTradeFilter, "trade filter requires both filter type and amount")
)

// WebSocket errors
var (
	// ErrInvalidSubscription is returned when a subscription is invalid.
	ErrInvalidSubscription = New(CodeInvalidSubscription, "invalid subscription")
)

// CTF (Conditional Token Framework) errors
var (
	// ErrMissingU256Value is returned when a numeric value is missing.
	ErrMissingU256Value = New(CodeMissingU256Value, "missing numeric value")
	// ErrMissingBackend is returned when CTF backend is required but not provided.
	ErrMissingBackend = New(CodeMissingBackend, "ctf backend is required")
	// ErrMissingTransactor is returned when CTF transactor is required but not provided.
	ErrMissingTransactor = New(CodeMissingTransactor, "ctf transactor is required")
	// ErrNegRiskAdapter is returned when neg risk adapter is not configured.
	ErrNegRiskAdapter = New(CodeNegRiskAdapter, "neg risk adapter is not configured")
	// ErrConfigNotFound is returned when CTF contract config is not found for chain ID.
	ErrConfigNotFound = New(CodeConfigNotFound, "ctf contract config not found for chain ID")
)

// Bridge errors
var (
	// ErrMissingFromAddress is returned when bridge transactor is missing from address.
	ErrMissingFromAddress = New(CodeMissingFromAddress, "bridge transactor missing from address")
	// ErrMissingDepositAddress is returned when EVM deposit address is missing.
	ErrMissingDepositAddress = New(CodeMissingDepositAddress, "evm deposit address is missing")
	// ErrWithdrawUnsupported is returned when bridge withdraw is not supported via the API.
	ErrWithdrawUnsupported = New(CodeWithdrawUnsupported, "bridge withdraw is not supported via the API; use WithdrawTo for on-chain transfers")
	// ErrMissingWithdrawRequest is returned when withdraw request is required but not provided.
	ErrMissingWithdrawRequest = New(CodeMissingWithdrawRequest, "withdraw request is required")
	// ErrMissingWithdrawAddress is returned when withdraw destination is required but not provided.
	ErrMissingWithdrawAddress = New(CodeMissingWithdrawAddress, "withdraw destination is required")
)
