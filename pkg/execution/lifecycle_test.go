package execution

import (
	"testing"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func TestNormalizeLifecycleState(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    LifecycleState
		wantErr bool
	}{
		{name: "created", raw: "created", want: LifecycleStateCreated},
		{name: "accepted live", raw: "LIVE", want: LifecycleStateAccepted},
		{name: "partial", raw: "partially_filled", want: LifecycleStatePartial},
		{name: "filled", raw: "filled", want: LifecycleStateFilled},
		{name: "canceled", raw: "cancelled", want: LifecycleStateCanceled},
		{name: "rejected", raw: "rejected", want: LifecycleStateRejected},
		{name: "failed maps to rejected", raw: "failed", want: LifecycleStateRejected},
		{name: "unknown", raw: "mystery", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeLifecycleState(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestEventFromOrderResponse(t *testing.T) {
	event, err := EventFromOrderResponse(clobtypes.OrderResponse{
		ID:     "ord-1",
		Status: "live",
	}, LifecycleSourceQuery)
	if err != nil {
		t.Fatalf("event from order: %v", err)
	}
	if event.OrderID != "ord-1" {
		t.Fatalf("unexpected order id: %q", event.OrderID)
	}
	if event.State != LifecycleStateAccepted {
		t.Fatalf("unexpected lifecycle state: %q", event.State)
	}
	if event.Source != LifecycleSourceQuery {
		t.Fatalf("unexpected source: %q", event.Source)
	}
	if event.RawStatus != "live" {
		t.Fatalf("unexpected raw status: %q", event.RawStatus)
	}
}

func TestEventFromOrderResponseRequiresOrderID(t *testing.T) {
	_, err := EventFromOrderResponse(clobtypes.OrderResponse{
		Status: "filled",
	}, LifecycleSourceReplay)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestEventFromCancelResponse(t *testing.T) {
	event, err := EventFromCancelResponse("ord-2", clobtypes.CancelResponse{
		Status: "ok",
	}, LifecycleSourceCancel, time.Unix(1710000000, 0))
	if err != nil {
		t.Fatalf("event from cancel: %v", err)
	}
	if event.State != LifecycleStateCanceled {
		t.Fatalf("unexpected state: %q", event.State)
	}
	if event.OrderID != "ord-2" {
		t.Fatalf("unexpected order id: %q", event.OrderID)
	}
	if event.OccurredAt.IsZero() {
		t.Fatalf("expected occurred_at to be set")
	}
}

func TestNewCreatedEvent(t *testing.T) {
	at := time.Unix(1710000000, 0)
	event, err := NewCreatedEvent("ord-3", at, LifecycleSourcePlace)
	if err != nil {
		t.Fatalf("new created event: %v", err)
	}
	if event.State != LifecycleStateCreated {
		t.Fatalf("unexpected state: %q", event.State)
	}
	if event.Source != LifecycleSourcePlace {
		t.Fatalf("unexpected source: %q", event.Source)
	}
	if !event.OccurredAt.Equal(at.UTC()) {
		t.Fatalf("unexpected occurred_at: %s", event.OccurredAt)
	}
}
