package rtds

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// ClientConfig controls RTDS WebSocket reconnect and heartbeat behavior.
type ClientConfig struct {
	Reconnect      bool
	ReconnectDelay time.Duration
	ReconnectMax   int
	PingInterval   time.Duration
}

// DefaultClientConfig returns deterministic defaults without reading environment variables.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Reconnect:      true,
		ReconnectDelay: 2 * time.Second,
		ReconnectMax:   5,
		PingInterval:   5 * time.Second,
	}
}

// ClientConfigFromEnv keeps backward compatibility with the old env-driven initialization behavior.
func ClientConfigFromEnv() ClientConfig {
	cfg := DefaultClientConfig()
	if raw := strings.TrimSpace(os.Getenv("RTDS_WS_RECONNECT")); raw != "" {
		cfg.Reconnect = raw != "0" && strings.ToLower(raw) != "false"
	}
	if raw := strings.TrimSpace(os.Getenv("RTDS_WS_RECONNECT_DELAY_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			cfg.ReconnectDelay = time.Duration(ms) * time.Millisecond
		}
	}
	if raw := strings.TrimSpace(os.Getenv("RTDS_WS_RECONNECT_MAX")); raw != "" {
		if max, err := strconv.Atoi(raw); err == nil {
			cfg.ReconnectMax = max
		}
	}
	if raw := strings.TrimSpace(os.Getenv("RTDS_WS_PING_INTERVAL_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			cfg.PingInterval = time.Duration(ms) * time.Millisecond
		}
	}
	return cfg.normalize()
}

func (c ClientConfig) normalize() ClientConfig {
	if c.ReconnectDelay <= 0 {
		c.ReconnectDelay = 2 * time.Second
	}
	if c.ReconnectMax < 0 {
		c.ReconnectMax = 5
	}
	if c.PingInterval <= 0 {
		c.PingInterval = 5 * time.Second
	}
	return c
}
