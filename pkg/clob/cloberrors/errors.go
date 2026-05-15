// Package cloberrors provides error mapping utilities for the Polymarket CLOB.
// It converts generic HTTP errors from the transport layer into structured,
// recognizable error types from pkg/errors for programmatic error handling.
package cloberrors

import (
	"fmt"
	"strings"

	sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/errors"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/types"
)

// FromTypeErr maps a generic types.Error (from transport layer) to a specific
// structured error type from pkg/errors.
func FromTypeErr(err *types.Error) error {
	if err == nil {
		return nil
	}

	// Map by Code if available (most reliable)
	code := strings.ToUpper(err.Code)
	switch code {
	case "INSUFFICIENT_FUNDS", "INSUFFICIENT_BALANCE", "INSUFFICIENT_ALLOWANCE":
		return fmt.Errorf("%w: %s", sdkerrors.ErrInsufficientFunds, err.Message)
	case "INVALID_SIGNATURE", "AUTH_INVALID_SIGNATURE":
		return fmt.Errorf("%w: %s", sdkerrors.ErrInvalidSignature, err.Message)
	case "ORDER_NOT_FOUND":
		return fmt.Errorf("%w: %s", sdkerrors.ErrOrderNotFound, err.Message)
	case "MARKET_CLOSED":
		return fmt.Errorf("%w: %s", sdkerrors.ErrMarketClosed, err.Message)
	case "GEOBLOCKED":
		return fmt.Errorf("%w: %s", sdkerrors.ErrGeoblocked, err.Message)
	case "INVALID_PRICE":
		return fmt.Errorf("%w: %s", sdkerrors.ErrInvalidPrice, err.Message)
	case "INVALID_SIZE":
		return fmt.Errorf("%w: %s", sdkerrors.ErrInvalidSize, err.Message)
	}

	// Fallback mapping by Status
	switch err.Status {
	case 401:
		return fmt.Errorf("%w: %s", sdkerrors.ErrUnauthorized, err.Message)
	case 403:
		if strings.Contains(strings.ToUpper(err.Message), "GEO") {
			return fmt.Errorf("%w: %s", sdkerrors.ErrGeoblocked, err.Message)
		}
		return fmt.Errorf("%w: %s", sdkerrors.ErrUnauthorized, err.Message)
	case 400:
		return fmt.Errorf("%w: %s", sdkerrors.ErrBadRequest, err.Message)
	case 429:
		return sdkerrors.ErrRateLimitExceeded
	case 500, 502, 503, 504:
		return fmt.Errorf("%w: %s", sdkerrors.ErrInternalServerError, err.Message)
	}

	return err
}