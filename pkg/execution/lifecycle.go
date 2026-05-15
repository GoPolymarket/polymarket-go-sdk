package execution

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"
)

// LifecycleState is the unified order lifecycle model shared by strategy/runtime layers.
type LifecycleState string

const (
	LifecycleStateCreated  LifecycleState = "created"
	LifecycleStateAccepted LifecycleState = "accepted"
	LifecycleStatePartial  LifecycleState = "partial"
	LifecycleStateFilled   LifecycleState = "filled"
	LifecycleStateCanceled LifecycleState = "canceled"
	LifecycleStateRejected LifecycleState = "rejected"
)

// LifecycleSource marks where a lifecycle event was observed.
type LifecycleSource string

const (
	LifecycleSourcePlace  LifecycleSource = "place"
	LifecycleSourceCancel LifecycleSource = "cancel"
	LifecycleSourceQuery  LifecycleSource = "query"
	LifecycleSourceReplay LifecycleSource = "replay"
)

var (
	errLifecycleOrderIDRequired = errors.New("execution lifecycle: order_id is required")
	errLifecycleUnknownStatus   = errors.New("execution lifecycle: unknown order status")
)

// LifecycleEvent represents one normalized order lifecycle transition.
type LifecycleEvent struct {
	OrderID    string          `json:"order_id"`
	State      LifecycleState  `json:"state"`
	Source     LifecycleSource `json:"source"`
	RawStatus  string          `json:"raw_status,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
}

// NewCreatedEvent emits a deterministic local "created" event.
func NewCreatedEvent(orderID string, at time.Time, source LifecycleSource) (LifecycleEvent, error) {
	id := strings.TrimSpace(orderID)
	if id == "" {
		return LifecycleEvent{}, errLifecycleOrderIDRequired
	}
	return LifecycleEvent{
		OrderID:    id,
		State:      LifecycleStateCreated,
		Source:     source,
		RawStatus:  string(LifecycleStateCreated),
		OccurredAt: normalizeEventTime(at),
	}, nil
}

// EventFromOrderResponse converts upstream order status to the unified lifecycle state.
func EventFromOrderResponse(order clobtypes.OrderResponse, source LifecycleSource) (LifecycleEvent, error) {
	id := strings.TrimSpace(order.ID)
	if id == "" {
		return LifecycleEvent{}, errLifecycleOrderIDRequired
	}
	state, err := NormalizeLifecycleState(order.Status)
	if err != nil {
		return LifecycleEvent{}, err
	}
	return LifecycleEvent{
		OrderID:    id,
		State:      state,
		Source:     source,
		RawStatus:  strings.ToLower(strings.TrimSpace(order.Status)),
		OccurredAt: time.Now().UTC(),
	}, nil
}

// EventFromCancelResponse maps cancel acknowledgement into the unified lifecycle model.
func EventFromCancelResponse(orderID string, cancel clobtypes.CancelResponse, source LifecycleSource, at time.Time) (LifecycleEvent, error) {
	id := strings.TrimSpace(orderID)
	if id == "" {
		return LifecycleEvent{}, errLifecycleOrderIDRequired
	}

	state := LifecycleStateCanceled
	raw := ""
	if _, ok := cancel.NotCanceled[id]; ok {
		state = LifecycleStateRejected
		if msg, ok := cancel.NotCanceled[id]; ok {
			raw = strings.ToLower(msg)
		}
	}

	return LifecycleEvent{
		OrderID:    id,
		State:      state,
		Source:     source,
		RawStatus:  raw,
		OccurredAt: normalizeEventTime(at),
	}, nil
}

// NormalizeLifecycleState collapses exchange-specific order status strings into the six-state model.
func NormalizeLifecycleState(rawStatus string) (LifecycleState, error) {
	s := strings.ToLower(strings.TrimSpace(rawStatus))
	switch {
	case s == "":
		return "", fmt.Errorf("%w: empty", errLifecycleUnknownStatus)
	case strings.Contains(s, "create"):
		return LifecycleStateCreated, nil
	case strings.Contains(s, "partial"), strings.Contains(s, "partially"), s == "matched":
		return LifecycleStatePartial, nil
	case strings.Contains(s, "fill"), strings.Contains(s, "complete"), strings.Contains(s, "execut"):
		return LifecycleStateFilled, nil
	case strings.Contains(s, "cancel"):
		return LifecycleStateCanceled, nil
	case strings.Contains(s, "reject"), strings.Contains(s, "fail"), strings.Contains(s, "error"):
		return LifecycleStateRejected, nil
	case strings.Contains(s, "accept"), strings.Contains(s, "open"), strings.Contains(s, "live"), strings.Contains(s, "submit"), strings.Contains(s, "queue"), strings.Contains(s, "post"):
		return LifecycleStateAccepted, nil
	default:
		return "", fmt.Errorf("%w: %s", errLifecycleUnknownStatus, s)
	}
}

func normalizeEventTime(at time.Time) time.Time {
	if at.IsZero() {
		return time.Now().UTC()
	}
	return at.UTC()
}
