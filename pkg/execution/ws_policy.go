package execution

import (
	"math"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/ws"
)

const (
	defaultWSReconnectDelay      = 2 * time.Second
	defaultWSReconnectMaxDelay   = 30 * time.Second
	defaultWSReconnectMultiplier = 2.0
	defaultWSReconnectMax        = 5
	defaultWSHeartbeatInterval   = 10 * time.Second
	defaultWSHeartbeatTimeout    = 30 * time.Second
)

// WSPolicy defines shared websocket reconnect + heartbeat behavior.
type WSPolicy struct {
	DisablePing bool

	ReconnectEnabled     bool
	ReconnectBaseDelay   time.Duration
	ReconnectMaxDelay    time.Duration
	ReconnectMultiplier  float64
	ReconnectMaxAttempts int

	HeartbeatInterval time.Duration
	HeartbeatTimeout  time.Duration
}

// DefaultWSPolicy returns default reconnect + heartbeat parameters.
func DefaultWSPolicy() WSPolicy {
	return WSPolicy{
		DisablePing: false,

		ReconnectEnabled:     true,
		ReconnectBaseDelay:   defaultWSReconnectDelay,
		ReconnectMaxDelay:    defaultWSReconnectMaxDelay,
		ReconnectMultiplier:  defaultWSReconnectMultiplier,
		ReconnectMaxAttempts: defaultWSReconnectMax,

		HeartbeatInterval: defaultWSHeartbeatInterval,
		HeartbeatTimeout:  defaultWSHeartbeatTimeout,
	}
}

// NextReconnectDelay returns delay for a reconnect attempt and whether reconnect should continue.
//
// attempt is 1-based.
func (p WSPolicy) NextReconnectDelay(attempt int) (time.Duration, bool) {
	n := p.withDefaults()
	if !n.ReconnectEnabled {
		return 0, false
	}
	if attempt <= 0 {
		attempt = 1
	}
	if n.ReconnectMaxAttempts >= 0 && attempt > n.ReconnectMaxAttempts {
		return 0, false
	}

	pow := math.Pow(n.ReconnectMultiplier, float64(attempt-1))
	delay := time.Duration(float64(n.ReconnectBaseDelay) * pow)
	if delay > n.ReconnectMaxDelay {
		delay = n.ReconnectMaxDelay
	}
	if delay <= 0 {
		delay = n.ReconnectBaseDelay
	}
	return delay, true
}

// IsHeartbeatExpired returns true when last pong exceeds timeout window.
func (p WSPolicy) IsHeartbeatExpired(lastPong, now time.Time) bool {
	n := p.withDefaults()
	if lastPong.IsZero() {
		return true
	}
	nowUTC := now.UTC()
	if nowUTC.IsZero() {
		nowUTC = time.Now().UTC()
	}
	return nowUTC.Sub(lastPong.UTC()) > n.HeartbeatTimeout
}

// ToCLOBConfig converts policy into clob ws client config.
func (p WSPolicy) ToCLOBConfig() ws.ClientConfig {
	n := p.withDefaults()
	return ws.ClientConfig{
		DisablePing:         n.DisablePing,
		Reconnect:           n.ReconnectEnabled,
		ReconnectDelay:      n.ReconnectBaseDelay,
		ReconnectMaxDelay:   n.ReconnectMaxDelay,
		ReconnectMultiplier: n.ReconnectMultiplier,
		ReconnectMax:        n.ReconnectMaxAttempts,
		HeartbeatInterval:   n.HeartbeatInterval,
		HeartbeatTimeout:    n.HeartbeatTimeout,
	}
}

func (p WSPolicy) withDefaults() WSPolicy {
	n := p
	if n.ReconnectBaseDelay <= 0 {
		n.ReconnectBaseDelay = defaultWSReconnectDelay
	}
	if n.ReconnectMaxDelay <= 0 {
		n.ReconnectMaxDelay = defaultWSReconnectMaxDelay
	}
	if n.ReconnectMultiplier <= 0 {
		n.ReconnectMultiplier = defaultWSReconnectMultiplier
	}
	if n.ReconnectMaxAttempts < 0 {
		n.ReconnectMaxAttempts = defaultWSReconnectMax
	}
	if n.HeartbeatInterval <= 0 {
		n.HeartbeatInterval = defaultWSHeartbeatInterval
	}
	if n.HeartbeatTimeout <= 0 {
		n.HeartbeatTimeout = defaultWSHeartbeatTimeout
	}
	if n.HeartbeatTimeout < n.HeartbeatInterval {
		n.HeartbeatTimeout = n.HeartbeatInterval * 3
	}
	return n
}
