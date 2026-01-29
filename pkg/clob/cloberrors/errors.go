package cloberrors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

var (
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrOrderNotFound        = errors.New("order not found")
	ErrMarketClosed         = errors.New("market closed")
	ErrInternalServerError  = errors.New("internal server error")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrBadRequest           = errors.New("bad request")
)

// FromTypeErr maps a generic types.Error to a specific error if possible.
func FromTypeErr(err *types.Error) error {
	if err == nil {
		return nil
	}

	// Map by Code if available (most reliable)
	switch strings.ToUpper(err.Code) {
	case "INSUFFICIENT_FUNDS":
		return fmt.Errorf("%w: %s", ErrInsufficientFunds, err.Message)
	case "INVALID_SIGNATURE":
		return fmt.Errorf("%w: %s", ErrInvalidSignature, err.Message)
	case "ORDER_NOT_FOUND":
		return fmt.Errorf("%w: %s", ErrOrderNotFound, err.Message)
	case "MARKET_CLOSED":
		return fmt.Errorf("%w: %s", ErrMarketClosed, err.Message)
	}

	// Map by Status
	switch err.Status {
	case 401:
		return fmt.Errorf("%w: %s", ErrUnauthorized, err.Message)
	case 403:
		// Often used for geoblocking or auth issues
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
