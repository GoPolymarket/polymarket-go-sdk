// Package cloberrors defines structured error types for the Polymarket CLOB.
// It allows developers to handle specific failure modes (e.g., insufficient funds)
// programmatically using standard Go error wrapping.
package cloberrors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

var (
	// ErrInsufficientFunds is returned when the maker has insufficient collateral or tokens.
	ErrInsufficientFunds = errors.New("insufficient funds")
	// ErrInvalidSignature is returned when the EIP-712 or HMAC signature is invalid.
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrRateLimitExceeded is returned when the user exceeds the API rate limits (HTTP 429).
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	// ErrOrderNotFound is returned when a requested order ID does not exist.
	ErrOrderNotFound = errors.New("order not found")
	// ErrMarketClosed is returned when attempting to trade in a resolved or closed market.
	ErrMarketClosed = errors.New("market closed")
	// ErrInternalServerError is returned when Polymarket returns a 5xx status code.
	ErrInternalServerError = errors.New("internal server error")
	// ErrUnauthorized is returned for authentication failures (HTTP 401/403).
	ErrUnauthorized = errors.New("unauthorized")
	// ErrBadRequest is returned for malformed requests (HTTP 400).
	ErrBadRequest = errors.New("bad request")
	// ErrGeoblocked is returned when the user's IP is restricted from trading.
	ErrGeoblocked = errors.New("geoblocked")
	// ErrPriceTooHigh or too low relative to current market.
	ErrInvalidPrice = errors.New("invalid price")
	// ErrInvalidSize is returned when order size is too small or precision is incorrect.
	ErrInvalidSize = errors.New("invalid size")
)

// FromTypeErr maps a generic types.Error (from transport layer) to a specific, 
// recognizable error type defined in this package.
func FromTypeErr(err *types.Error) error {
	if err == nil {
		return nil
	}

	// Map by Code if available (most reliable)
	code := strings.ToUpper(err.Code)
	switch code {
	case "INSUFFICIENT_FUNDS", "INSUFFICIENT_BALANCE", "INSUFFICIENT_ALLOWANCE":
		return fmt.Errorf("%w: %s", ErrInsufficientFunds, err.Message)
	case "INVALID_SIGNATURE", "AUTH_INVALID_SIGNATURE":
		return fmt.Errorf("%w: %s", ErrInvalidSignature, err.Message)
	case "ORDER_NOT_FOUND":
		return fmt.Errorf("%w: %s", ErrOrderNotFound, err.Message)
	case "MARKET_CLOSED":
		return fmt.Errorf("%w: %s", ErrMarketClosed, err.Message)
	case "GEOBLOCKED":
		return fmt.Errorf("%w: %s", ErrGeoblocked, err.Message)
	case "INVALID_PRICE":
		return fmt.Errorf("%w: %s", ErrInvalidPrice, err.Message)
	case "INVALID_SIZE":
		return fmt.Errorf("%w: %s", ErrInvalidSize, err.Message)
	}

	// Fallback mapping by Status
	switch err.Status {
	case 401:
		return fmt.Errorf("%w: %s", ErrUnauthorized, err.Message)
	case 403:
		if strings.Contains(strings.ToUpper(err.Message), "GEO") {
			return fmt.Errorf("%w: %s", ErrGeoblocked, err.Message)
		}
		return fmt.Errorf("%w: %s", ErrUnauthorized, err.Message)
	case 400:
		return fmt.Errorf("%w: %s", ErrBadRequest, err.Message)
	case 429:
		return ErrRateLimitExceeded
	case 500, 502, 503, 504:
		return fmt.Errorf("%w: %s", ErrInternalServerError, err.Message)
	}

	return err
}