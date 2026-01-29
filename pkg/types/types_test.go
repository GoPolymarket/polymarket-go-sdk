package types

import (
	"math/big"
	"testing"
)

func TestU256(t *testing.T) {
	u := U256{Int: big.NewInt(100)}
	raw, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if string(raw) != `"100"` {
		t.Errorf("expected \"100\", got %s", string(raw))
	}

	var u2 U256
	err = u2.UnmarshalJSON([]byte(`"200"`))
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if u2.Int.Int64() != 200 {
		t.Errorf("expected 200, got %d", u2.Int.Int64())
	}
}

func TestAddress(t *testing.T) {
	addrStr := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
	var a Address
	err := a.UnmarshalJSON([]byte(`"` + addrStr + `"`))
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if a.String() != addrStr {
		t.Errorf("expected %s, got %s", addrStr, a.String())
	}
}
