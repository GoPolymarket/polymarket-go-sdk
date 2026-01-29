package ctf

import (
	"context"
	"testing"
)

func TestCtfMethods(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	t.Run("ConditionID", func(t *testing.T) {
		_, _ = client.ConditionID(ctx, &ConditionIDRequest{})
	})

	t.Run("CollectionID", func(t *testing.T) {
		_, _ = client.CollectionID(ctx, &CollectionIDRequest{})
	})

	t.Run("PositionID", func(t *testing.T) {
		_, _ = client.PositionID(ctx, &PositionIDRequest{})
	})
}
