package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// Address represents an EVM address.
type Address = common.Address

// U256 represents a 256-bit unsigned integer using big.Int.
type U256 struct {
	*big.Int
}

// Decimal represents a fixed-point decimal using shopspring/decimal.
type Decimal = decimal.Decimal

// Pagination represents simple pagination controls.
type Pagination struct {
	Limit  int
	Offset int
}

// Error represents a standard API error.
type Error struct {
	Status  int    `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("api error: %s (code=%s, status=%d)", e.Message, e.Code, e.Status)
	}
	return fmt.Sprintf("api error: %s (status=%d)", e.Message, e.Status)
}

// MarshalJSON encodes the U256 as a decimal string.
func (u U256) MarshalJSON() ([]byte, error) {
	if u.Int == nil {
		return []byte("null"), nil
	}
	return json.Marshal(u.Int.String())
}

// UnmarshalJSON parses a U256 from a decimal or hex string/number.
func (u *U256) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		u.Int = nil
		return nil
	}

	var s string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
	} else {
		var num json.Number
		if err := json.Unmarshal(data, &num); err != nil {
			return err
		}
		s = num.String()
	}

	s = strings.TrimSpace(s)
	if s == "" {
		u.Int = nil
		return nil
	}

	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		base = 16
		s = s[2:]
	}

	value, ok := new(big.Int).SetString(s, base)
	if !ok {
		return fmt.Errorf("invalid U256 value: %q", s)
	}
	u.Int = value
	return nil
}
