package rtds

import (
	"encoding/json"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

// ConnectionState represents RTDS connection status.
type ConnectionState string

const (
	ConnectionDisconnected ConnectionState = "disconnected"
	ConnectionConnected    ConnectionState = "connected"
)

// ConnectionStateEvent captures connection transitions.
type ConnectionStateEvent struct {
	State    ConnectionState `json:"state"`
	Recorded int64           `json:"recorded"`
}

// RtdsMessage is the raw RTDS message wrapper.
type RtdsMessage struct {
	Topic     string          `json:"topic"`
	MsgType   string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// SubscriptionAction indicates subscribe/unsubscribe.
type SubscriptionAction string

const (
	SubscribeAction   SubscriptionAction = "subscribe"
	UnsubscribeAction SubscriptionAction = "unsubscribe"
)

// Subscription describes a single RTDS subscription.
type Subscription struct {
	Topic    string      `json:"topic"`
	MsgType  string      `json:"type"`
	Filters  interface{} `json:"filters,omitempty"`
	ClobAuth *ClobAuth   `json:"clob_auth,omitempty"`
}

// MarshalJSON customizes filters encoding to align with RTDS expectations.
// For non-Chainlink topics, JSON strings are parsed and sent as raw JSON.
// Chainlink filters are always serialized as a JSON string.
func (s Subscription) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"topic": s.Topic,
		"type":  s.MsgType,
	}
	if s.Filters != nil {
		filters := s.Filters
		if raw, ok := s.Filters.(string); ok && s.Topic != string(ChainlinkPrice) {
			var parsed interface{}
			if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
				filters = parsed
			}
		}
		payload["filters"] = filters
	}
	if s.ClobAuth != nil {
		payload["clob_auth"] = s.ClobAuth
	}
	return json.Marshal(payload)
}

// SubscriptionRequest is the top-level RTDS subscribe/unsubscribe payload.
type SubscriptionRequest struct {
	Action        SubscriptionAction `json:"action"`
	Subscriptions []Subscription     `json:"subscriptions"`
}

// ClobAuth carries CLOB credentials for authenticated comment streams.
type ClobAuth struct {
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// EventType represents RTDS topic categories.
type EventType string

const (
	CryptoPrice    EventType = "crypto_prices"
	ChainlinkPrice EventType = "crypto_prices_chainlink"
	Comments       EventType = "comments"
	Activity       EventType = "activity"
)

// BaseEvent carries message metadata.
type BaseEvent struct {
	Topic            EventType `json:"-"`
	MessageType      string    `json:"-"`
	MessageTimestamp int64     `json:"-"`
}

// CryptoPriceEvent is a Binance price update payload.
type CryptoPriceEvent struct {
	BaseEvent
	Symbol    string        `json:"symbol"`
	Timestamp int64         `json:"timestamp"`
	Value     types.Decimal `json:"value"`
}

// ChainlinkPriceEvent is a Chainlink price update payload.
type ChainlinkPriceEvent struct {
	BaseEvent
	Symbol    string        `json:"symbol"`
	Timestamp int64         `json:"timestamp"`
	Value     types.Decimal `json:"value"`
}

// CommentType enumerates comment event types.
type CommentType string

const (
	CommentCreated  CommentType = "comment_created"
	CommentRemoved  CommentType = "comment_removed"
	ReactionCreated CommentType = "reaction_created"
	ReactionRemoved CommentType = "reaction_removed"
)

// CommentProfile describes the comment author.
type CommentProfile struct {
	BaseAddress           types.Address  `json:"baseAddress"`
	DisplayUsernamePublic bool           `json:"displayUsernamePublic,omitempty"`
	Name                  string         `json:"name"`
	ProxyWallet           *types.Address `json:"proxyWallet,omitempty"`
	Pseudonym             *string        `json:"pseudonym,omitempty"`
}

// CommentEvent is a comment stream payload.
type CommentEvent struct {
	BaseEvent
	ID               string         `json:"id"`
	Body             string         `json:"body"`
	CreatedAt        time.Time      `json:"createdAt"`
	ParentCommentID  *string        `json:"parentCommentID,omitempty"`
	ParentEntityID   int64          `json:"parentEntityID"`
	ParentEntityType string         `json:"parentEntityType"`
	Profile          CommentProfile `json:"profile"`
	ReactionCount    int64          `json:"reactionCount,omitempty"`
	ReplyAddress     *types.Address `json:"replyAddress,omitempty"`
	ReportCount      int64          `json:"reportCount,omitempty"`
	UserAddress      types.Address  `json:"userAddress"`
}

// OrdersMatchedEvent is an activity stream payload for matched orders.
type OrdersMatchedEvent struct {
	BaseEvent
	Asset           string  `json:"asset"`
	Bio             string  `json:"bio"`
	ConditionID     string  `json:"conditionId"`
	EventSlug       string  `json:"eventSlug"`
	Icon            string  `json:"icon"`
	Name            string  `json:"name"`
	Outcome         string  `json:"outcome"`
	OutcomeIndex    int     `json:"outcomeIndex"`
	Price           float64 `json:"price"`
	ProfileImage    string  `json:"profileImage"`
	ProxyWallet     string  `json:"proxyWallet"`
	Pseudonym       string  `json:"pseudonym"`
	Side            string  `json:"side"`
	Size            float64 `json:"size"`
	Slug            string  `json:"slug"`
	Timestamp       int64   `json:"timestamp"`
	Title           string  `json:"title"`
	TransactionHash string  `json:"transactionHash"`
}

// CommentFilter configures the comments subscription.
type CommentFilter struct {
	Type    *CommentType
	Auth    *auth.APIKey
	Filters interface{}
}
