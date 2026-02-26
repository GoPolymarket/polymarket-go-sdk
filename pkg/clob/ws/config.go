package ws

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// ClientConfig controls runtime behavior of the CLOB WebSocket client.
type ClientConfig struct {
	Debug               bool
	DisablePing         bool
	Reconnect           bool
	ReconnectDelay      time.Duration
	ReconnectMaxDelay   time.Duration
	ReconnectMultiplier float64
	ReconnectMax        int
	HeartbeatInterval   time.Duration
	HeartbeatTimeout    time.Duration
	ReadTimeout         time.Duration
}

// DefaultClientConfig returns stable defaults independent from process environment variables.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Debug:               false,
		DisablePing:         false,
		Reconnect:           true,
		ReconnectDelay:      2 * time.Second,
		ReconnectMaxDelay:   30 * time.Second,
		ReconnectMultiplier: 2.0,
		ReconnectMax:        5,
		HeartbeatInterval:   10 * time.Second,
		HeartbeatTimeout:    30 * time.Second,
		ReadTimeout:         DefaultReadTimeout,
	}
}

// ClientConfigFromEnv preserves legacy behavior for users who configured WS behavior via env vars.
func ClientConfigFromEnv() ClientConfig {
	cfg := DefaultClientConfig()

	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT")); raw != "" {
		cfg.Reconnect = raw != "0" && strings.ToLower(raw) != "false"
	}
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_DELAY_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			cfg.ReconnectDelay = time.Duration(ms) * time.Millisecond
		}
	}
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_MAX_DELAY_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			cfg.ReconnectMaxDelay = time.Duration(ms) * time.Millisecond
		}
	}
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_BACKOFF_MULTIPLIER")); raw != "" {
		if mult, err := strconv.ParseFloat(raw, 64); err == nil && mult > 0 {
			cfg.ReconnectMultiplier = mult
		}
	}
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_MAX")); raw != "" {
		if max, err := strconv.Atoi(raw); err == nil {
			cfg.ReconnectMax = max
		}
	}
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_HEARTBEAT_INTERVAL_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			cfg.HeartbeatInterval = time.Duration(ms) * time.Millisecond
		}
	}
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_HEARTBEAT_TIMEOUT_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			cfg.HeartbeatTimeout = time.Duration(ms) * time.Millisecond
		}
	} else if cfg.HeartbeatInterval > 0 {
		cfg.HeartbeatTimeout = cfg.HeartbeatInterval * 3
	}
	cfg.Debug = os.Getenv("CLOB_WS_DEBUG") != ""
	cfg.DisablePing = os.Getenv("CLOB_WS_DISABLE_PING") != ""
	return cfg.normalize()
}

func (c ClientConfig) normalize() ClientConfig {
	if c.ReconnectDelay <= 0 {
		c.ReconnectDelay = 2 * time.Second
	}
	if c.ReconnectMaxDelay <= 0 {
		c.ReconnectMaxDelay = 30 * time.Second
	}
	if c.ReconnectMultiplier <= 0 {
		c.ReconnectMultiplier = 2.0
	}
	if c.ReconnectMax < 0 {
		c.ReconnectMax = 5
	}
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = 10 * time.Second
	}
	if c.HeartbeatTimeout <= 0 {
		c.HeartbeatTimeout = c.HeartbeatInterval * 3
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = DefaultReadTimeout
	}
	return c
}
