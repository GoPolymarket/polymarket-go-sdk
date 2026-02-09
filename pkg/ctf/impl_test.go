package ctf

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestNilRequests(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"ConditionID", func() error { _, err := client.ConditionID(ctx, nil); return err }},
		{"CollectionID", func() error { _, err := client.CollectionID(ctx, nil); return err }},
		{"PositionID", func() error { _, err := client.PositionID(ctx, nil); return err }},
		{"PrepareCondition", func() error { _, err := client.PrepareCondition(ctx, nil); return err }},
		{"SplitPosition", func() error { _, err := client.SplitPosition(ctx, nil); return err }},
		{"MergePositions", func() error { _, err := client.MergePositions(ctx, nil); return err }},
		{"RedeemPositions", func() error { _, err := client.RedeemPositions(ctx, nil); return err }},
		{"RedeemNegRisk", func() error { _, err := client.RedeemNegRisk(ctx, nil); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if !errors.Is(err, ErrMissingRequest) {
				t.Errorf("expected ErrMissingRequest, got %v", err)
			}
		})
	}
}

func TestConditionIDMissingOutcomeSlotCount(t *testing.T) {
	client := NewClient()
	_, err := client.ConditionID(context.Background(), &ConditionIDRequest{})
	if !errors.Is(err, ErrMissingU256Value) {
		t.Errorf("expected ErrMissingU256Value, got %v", err)
	}
}

func TestCollectionIDMissingIndexSet(t *testing.T) {
	client := NewClient()
	_, err := client.CollectionID(context.Background(), &CollectionIDRequest{})
	if !errors.Is(err, ErrMissingU256Value) {
		t.Errorf("expected ErrMissingU256Value, got %v", err)
	}
}

func TestConditionIDDeterministic(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	oracle := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	questionID := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	outcomeSlotCount := big.NewInt(2)

	resp1, err := client.ConditionID(ctx, &ConditionIDRequest{
		Oracle:           oracle,
		QuestionID:       questionID,
		OutcomeSlotCount: outcomeSlotCount,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deterministic: same inputs produce same output.
	resp2, err := client.ConditionID(ctx, &ConditionIDRequest{
		Oracle:           oracle,
		QuestionID:       questionID,
		OutcomeSlotCount: outcomeSlotCount,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp1.ConditionID != resp2.ConditionID {
		t.Errorf("expected deterministic result, got %s and %s", resp1.ConditionID.Hex(), resp2.ConditionID.Hex())
	}

	// Verify manually: keccak256(oracle ++ questionID ++ leftPad32(2))
	buf := make([]byte, 0, 20+32+32)
	buf = append(buf, oracle.Bytes()...)
	buf = append(buf, questionID.Bytes()...)
	buf = append(buf, leftPad32(outcomeSlotCount)...)
	expected := crypto.Keccak256Hash(buf)
	if resp1.ConditionID != expected {
		t.Errorf("expected %s, got %s", expected.Hex(), resp1.ConditionID.Hex())
	}
}

func TestCollectionIDDeterministic(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	parentCollectionID := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	conditionID := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	indexSet := big.NewInt(1)

	resp, err := client.CollectionID(ctx, &CollectionIDRequest{
		ParentCollectionID: parentCollectionID,
		ConditionID:        conditionID,
		IndexSet:           indexSet,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	buf := make([]byte, 0, 32+32+32)
	buf = append(buf, parentCollectionID.Bytes()...)
	buf = append(buf, conditionID.Bytes()...)
	buf = append(buf, leftPad32(indexSet)...)
	expected := crypto.Keccak256Hash(buf)
	if resp.CollectionID != expected {
		t.Errorf("expected %s, got %s", expected.Hex(), resp.CollectionID.Hex())
	}
}

func TestPositionIDDeterministic(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	collateralToken := common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
	collectionID := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd")

	resp, err := client.PositionID(ctx, &PositionIDRequest{
		CollateralToken: collateralToken,
		CollectionID:    collectionID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	buf := make([]byte, 0, 20+32)
	buf = append(buf, collateralToken.Bytes()...)
	buf = append(buf, collectionID.Bytes()...)
	expected := new(big.Int).SetBytes(crypto.Keccak256Hash(buf).Bytes())
	if resp.PositionID.Cmp(expected) != 0 {
		t.Errorf("expected %s, got %s", expected.String(), resp.PositionID.String())
	}
}

func TestConditionIDDifferentInputs(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	resp1, _ := client.ConditionID(ctx, &ConditionIDRequest{
		Oracle:           common.HexToAddress("0x1111111111111111111111111111111111111111"),
		QuestionID:       common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		OutcomeSlotCount: big.NewInt(2),
	})
	resp2, _ := client.ConditionID(ctx, &ConditionIDRequest{
		Oracle:           common.HexToAddress("0x2222222222222222222222222222222222222222"),
		QuestionID:       common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		OutcomeSlotCount: big.NewInt(2),
	})
	if resp1.ConditionID == resp2.ConditionID {
		t.Error("different oracles should produce different condition IDs")
	}
}

func TestTransactionMethodsWithoutBackend(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	t.Run("PrepareCondition", func(t *testing.T) {
		_, err := client.PrepareCondition(ctx, &PrepareConditionRequest{
			OutcomeSlotCount: big.NewInt(2),
		})
		if !errors.Is(err, ErrMissingBackend) {
			t.Errorf("expected ErrMissingBackend, got %v", err)
		}
	})

	t.Run("SplitPosition", func(t *testing.T) {
		_, err := client.SplitPosition(ctx, &SplitPositionRequest{
			Partition: BinaryPartition,
			Amount:    big.NewInt(100),
		})
		if !errors.Is(err, ErrMissingBackend) {
			t.Errorf("expected ErrMissingBackend, got %v", err)
		}
	})

	t.Run("MergePositions", func(t *testing.T) {
		_, err := client.MergePositions(ctx, &MergePositionsRequest{
			Partition: BinaryPartition,
			Amount:    big.NewInt(100),
		})
		if !errors.Is(err, ErrMissingBackend) {
			t.Errorf("expected ErrMissingBackend, got %v", err)
		}
	})

	t.Run("RedeemPositions", func(t *testing.T) {
		_, err := client.RedeemPositions(ctx, &RedeemPositionsRequest{
			IndexSets: BinaryPartition,
		})
		if !errors.Is(err, ErrMissingBackend) {
			t.Errorf("expected ErrMissingBackend, got %v", err)
		}
	})

	t.Run("RedeemNegRisk", func(t *testing.T) {
		_, err := client.RedeemNegRisk(ctx, &RedeemNegRiskRequest{
			Amounts: []*big.Int{big.NewInt(100)},
		})
		if !errors.Is(err, ErrNegRiskAdapter) {
			t.Errorf("expected ErrNegRiskAdapter, got %v", err)
		}
	})
}

func TestTransactionValidation(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	t.Run("PrepareConditionMissingOutcome", func(t *testing.T) {
		_, err := client.PrepareCondition(ctx, &PrepareConditionRequest{})
		if !errors.Is(err, ErrMissingU256Value) {
			t.Errorf("expected ErrMissingU256Value, got %v", err)
		}
	})

	t.Run("SplitPositionMissingAmount", func(t *testing.T) {
		_, err := client.SplitPosition(ctx, &SplitPositionRequest{
			Partition: BinaryPartition,
		})
		if !errors.Is(err, ErrMissingU256Value) {
			t.Errorf("expected ErrMissingU256Value, got %v", err)
		}
	})

	t.Run("SplitPositionMissingPartition", func(t *testing.T) {
		_, err := client.SplitPosition(ctx, &SplitPositionRequest{
			Amount: big.NewInt(100),
		})
		if err == nil {
			t.Error("expected error for missing partition")
		}
	})

	t.Run("MergePositionsMissingAmount", func(t *testing.T) {
		_, err := client.MergePositions(ctx, &MergePositionsRequest{
			Partition: BinaryPartition,
		})
		if !errors.Is(err, ErrMissingU256Value) {
			t.Errorf("expected ErrMissingU256Value, got %v", err)
		}
	})

	t.Run("MergePositionsMissingPartition", func(t *testing.T) {
		_, err := client.MergePositions(ctx, &MergePositionsRequest{
			Amount: big.NewInt(100),
		})
		if err == nil {
			t.Error("expected error for missing partition")
		}
	})

	t.Run("RedeemPositionsMissingIndexSets", func(t *testing.T) {
		_, err := client.RedeemPositions(ctx, &RedeemPositionsRequest{})
		if err == nil {
			t.Error("expected error for missing index sets")
		}
	})

	t.Run("RedeemNegRiskMissingAmounts", func(t *testing.T) {
		_, err := client.RedeemNegRisk(ctx, &RedeemNegRiskRequest{})
		if err == nil {
			t.Error("expected error for missing amounts")
		}
	})
}

func TestLeftPad32(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		result := leftPad32(nil)
		if len(result) != 32 {
			t.Errorf("expected 32 bytes, got %d", len(result))
		}
		for i, b := range result {
			if b != 0 {
				t.Errorf("expected zero at index %d, got %d", i, b)
			}
		}
	})

	t.Run("SmallValue", func(t *testing.T) {
		result := leftPad32(big.NewInt(1))
		if len(result) != 32 {
			t.Errorf("expected 32 bytes, got %d", len(result))
		}
		if result[31] != 1 {
			t.Errorf("expected last byte to be 1, got %d", result[31])
		}
		for i := 0; i < 31; i++ {
			if result[i] != 0 {
				t.Errorf("expected zero at index %d, got %d", i, result[i])
			}
		}
	})

	t.Run("LargeValue", func(t *testing.T) {
		// 32-byte value (max uint256)
		val := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
		result := leftPad32(val)
		if len(result) != 32 {
			t.Errorf("expected 32 bytes, got %d", len(result))
		}
		for i, b := range result {
			if b != 0xff {
				t.Errorf("expected 0xff at index %d, got %d", i, b)
			}
		}
	})
}

func TestNewClientWithBackendUnsupportedChain(t *testing.T) {
	_, err := NewClientWithBackend(nil, nil, 999)
	// nil backend is checked first for supported chains, but unsupported chain returns ErrConfigNotFound
	// Actually backend is checked after config resolution
	if !errors.Is(err, ErrMissingBackend) && !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("expected ErrMissingBackend or ErrConfigNotFound, got %v", err)
	}
}

func TestNewClientWithNegRiskUnsupportedChain(t *testing.T) {
	_, err := NewClientWithNegRisk(nil, nil, 999)
	if !errors.Is(err, ErrMissingBackend) && !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("expected ErrMissingBackend or ErrConfigNotFound, got %v", err)
	}
}

func TestResolveConfig(t *testing.T) {
	t.Run("PolygonStandard", func(t *testing.T) {
		cfg, ok := resolveConfig(PolygonChainID, false)
		if !ok {
			t.Fatal("expected config for Polygon")
		}
		if cfg.ConditionalTokens == (common.Address{}) {
			t.Error("expected non-zero ConditionalTokens address")
		}
	})

	t.Run("PolygonNegRisk", func(t *testing.T) {
		cfg, ok := resolveConfig(PolygonChainID, true)
		if !ok {
			t.Fatal("expected config for Polygon NegRisk")
		}
		if cfg.NegRiskAdapter == nil {
			t.Error("expected NegRiskAdapter address")
		}
	})

	t.Run("AmoyStandard", func(t *testing.T) {
		cfg, ok := resolveConfig(AmoyChainID, false)
		if !ok {
			t.Fatal("expected config for Amoy")
		}
		if cfg.ConditionalTokens == (common.Address{}) {
			t.Error("expected non-zero ConditionalTokens address")
		}
	})

	t.Run("UnsupportedChain", func(t *testing.T) {
		_, ok := resolveConfig(999, false)
		if ok {
			t.Error("expected no config for unsupported chain")
		}
	})
}
