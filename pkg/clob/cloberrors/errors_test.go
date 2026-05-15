package cloberrors

import (
	"errors"
	"testing"

	sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/errors"
	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/types"
)

func TestFromTypeErr_NilError(t *testing.T) {
	result := FromTypeErr(nil)
	if result != nil {
		t.Errorf("FromTypeErr(nil) = %v, want nil", result)
	}
}

func TestFromTypeErr_ByCode(t *testing.T) {
	tests := []struct {
		name          string
		inputError    *types.Error
		expectedError error
		checkMessage  bool
	}{
		{
			name: "insufficient funds",
			inputError: &types.Error{
				Code:    "INSUFFICIENT_FUNDS",
				Message: "Not enough USDC",
				Status:  400,
			},
			expectedError: sdkerrors.ErrInsufficientFunds,
			checkMessage:  true,
		},
		{
			name: "insufficient balance",
			inputError: &types.Error{
				Code:    "INSUFFICIENT_BALANCE",
				Message: "Balance too low",
				Status:  400,
			},
			expectedError: sdkerrors.ErrInsufficientFunds,
			checkMessage:  true,
		},
		{
			name: "insufficient allowance",
			inputError: &types.Error{
				Code:    "INSUFFICIENT_ALLOWANCE",
				Message: "Allowance not set",
				Status:  400,
			},
			expectedError: sdkerrors.ErrInsufficientFunds,
			checkMessage:  true,
		},
		{
			name: "invalid signature",
			inputError: &types.Error{
				Code:    "INVALID_SIGNATURE",
				Message: "Signature verification failed",
				Status:  401,
			},
			expectedError: sdkerrors.ErrInvalidSignature,
			checkMessage:  true,
		},
		{
			name: "auth invalid signature",
			inputError: &types.Error{
				Code:    "AUTH_INVALID_SIGNATURE",
				Message: "Auth failed",
				Status:  401,
			},
			expectedError: sdkerrors.ErrInvalidSignature,
			checkMessage:  true,
		},
		{
			name: "order not found",
			inputError: &types.Error{
				Code:    "ORDER_NOT_FOUND",
				Message: "Order ID does not exist",
				Status:  404,
			},
			expectedError: sdkerrors.ErrOrderNotFound,
			checkMessage:  true,
		},
		{
			name: "market closed",
			inputError: &types.Error{
				Code:    "MARKET_CLOSED",
				Message: "Market has been resolved",
				Status:  400,
			},
			expectedError: sdkerrors.ErrMarketClosed,
			checkMessage:  true,
		},
		{
			name: "geoblocked",
			inputError: &types.Error{
				Code:    "GEOBLOCKED",
				Message: "Your region is restricted",
				Status:  403,
			},
			expectedError: sdkerrors.ErrGeoblocked,
			checkMessage:  true,
		},
		{
			name: "invalid price",
			inputError: &types.Error{
				Code:    "INVALID_PRICE",
				Message: "Price out of range",
				Status:  400,
			},
			expectedError: sdkerrors.ErrInvalidPrice,
			checkMessage:  true,
		},
		{
			name: "invalid size",
			inputError: &types.Error{
				Code:    "INVALID_SIZE",
				Message: "Size too small",
				Status:  400,
			},
			expectedError: sdkerrors.ErrInvalidSize,
			checkMessage:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromTypeErr(tt.inputError)
			if result == nil {
				t.Fatal("FromTypeErr returned nil")
			}

			if !errors.Is(result, tt.expectedError) {
				t.Errorf("FromTypeErr() error = %v, want %v", result, tt.expectedError)
			}

			if tt.checkMessage && result.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

func TestFromTypeErr_ByStatus(t *testing.T) {
	tests := []struct {
		name          string
		inputError    *types.Error
		expectedError error
	}{
		{
			name: "401 unauthorized",
			inputError: &types.Error{
				Code:    "",
				Message: "Authentication required",
				Status:  401,
			},
			expectedError: sdkerrors.ErrUnauthorized,
		},
		{
			name: "403 forbidden",
			inputError: &types.Error{
				Code:    "",
				Message: "Access denied",
				Status:  403,
			},
			expectedError: sdkerrors.ErrUnauthorized,
		},
		{
			name: "403 with geo in message",
			inputError: &types.Error{
				Code:    "",
				Message: "GEO restriction applied",
				Status:  403,
			},
			expectedError: sdkerrors.ErrGeoblocked,
		},
		{
			name: "400 bad request",
			inputError: &types.Error{
				Code:    "",
				Message: "Invalid request",
				Status:  400,
			},
			expectedError: sdkerrors.ErrBadRequest,
		},
		{
			name: "429 rate limit",
			inputError: &types.Error{
				Code:    "",
				Message: "Too many requests",
				Status:  429,
			},
			expectedError: sdkerrors.ErrRateLimitExceeded,
		},
		{
			name: "500 internal server error",
			inputError: &types.Error{
				Code:    "",
				Message: "Server error",
				Status:  500,
			},
			expectedError: sdkerrors.ErrInternalServerError,
		},
		{
			name: "502 bad gateway",
			inputError: &types.Error{
				Code:    "",
				Message: "Bad gateway",
				Status:  502,
			},
			expectedError: sdkerrors.ErrInternalServerError,
		},
		{
			name: "503 service unavailable",
			inputError: &types.Error{
				Code:    "",
				Message: "Service unavailable",
				Status:  503,
			},
			expectedError: sdkerrors.ErrInternalServerError,
		},
		{
			name: "504 gateway timeout",
			inputError: &types.Error{
				Code:    "",
				Message: "Gateway timeout",
				Status:  504,
			},
			expectedError: sdkerrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromTypeErr(tt.inputError)
			if result == nil {
				t.Fatal("FromTypeErr returned nil")
			}

			if !errors.Is(result, tt.expectedError) {
				t.Errorf("FromTypeErr() error = %v, want %v", result, tt.expectedError)
			}
		})
	}
}

func TestFromTypeErr_UnknownError(t *testing.T) {
	t.Run("unknown code and status", func(t *testing.T) {
		inputError := &types.Error{
			Code:    "UNKNOWN_ERROR",
			Message: "Something went wrong",
			Status:  418, // I'm a teapot
		}

		result := FromTypeErr(inputError)
		if result == nil {
			t.Fatal("FromTypeErr returned nil")
		}

		// Should return the original error
		if result != inputError {
			t.Errorf("FromTypeErr() should return original error for unknown codes/statuses")
		}
	})
}

func TestFromTypeErr_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"lowercase", "insufficient_funds"},
		{"uppercase", "INSUFFICIENT_FUNDS"},
		{"mixed case", "Insufficient_Funds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputError := &types.Error{
				Code:    tt.code,
				Message: "Test message",
				Status:  400,
			}

			result := FromTypeErr(inputError)
			if !errors.Is(result, sdkerrors.ErrInsufficientFunds) {
				t.Errorf("FromTypeErr() should be case-insensitive for code matching")
			}
		})
	}
}

func TestFromTypeErr_CodePriorityOverStatus(t *testing.T) {
	t.Run("code takes priority", func(t *testing.T) {
		// Code says insufficient funds, but status is 401
		inputError := &types.Error{
			Code:    "INSUFFICIENT_FUNDS",
			Message: "Not enough balance",
			Status:  401,
		}

		result := FromTypeErr(inputError)
		if !errors.Is(result, sdkerrors.ErrInsufficientFunds) {
			t.Errorf("Code should take priority over status in error mapping")
		}
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("errors can be unwrapped", func(t *testing.T) {
		inputError := &types.Error{
			Code:    "INSUFFICIENT_FUNDS",
			Message: "Not enough USDC",
			Status:  400,
		}

		result := FromTypeErr(inputError)

		// Should be able to use errors.Is
		if !errors.Is(result, sdkerrors.ErrInsufficientFunds) {
			t.Error("Error should be unwrappable with errors.Is")
		}

		// Error message should contain the original message
		if result.Error() == "" {
			t.Error("Error message should not be empty")
		}
	})
}

func TestAllDefinedErrors(t *testing.T) {
	// Ensure all defined errors are non-nil
	definedErrors := []struct {
		name string
		err  error
	}{
		{"sdkerrors.ErrInsufficientFunds", sdkerrors.ErrInsufficientFunds},
		{"sdkerrors.ErrInvalidSignature", sdkerrors.ErrInvalidSignature},
		{"sdkerrors.ErrRateLimitExceeded", sdkerrors.ErrRateLimitExceeded},
		{"sdkerrors.ErrOrderNotFound", sdkerrors.ErrOrderNotFound},
		{"sdkerrors.ErrMarketClosed", sdkerrors.ErrMarketClosed},
		{"sdkerrors.ErrInternalServerError", sdkerrors.ErrInternalServerError},
		{"sdkerrors.ErrUnauthorized", sdkerrors.ErrUnauthorized},
		{"sdkerrors.ErrBadRequest", sdkerrors.ErrBadRequest},
		{"sdkerrors.ErrGeoblocked", sdkerrors.ErrGeoblocked},
		{"sdkerrors.ErrInvalidPrice", sdkerrors.ErrInvalidPrice},
		{"sdkerrors.ErrInvalidSize", sdkerrors.ErrInvalidSize},
	}

	for _, tt := range definedErrors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s should have a non-empty error message", tt.name)
			}
		})
	}
}
