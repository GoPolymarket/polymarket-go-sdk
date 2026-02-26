package polymarket

import (
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/rtds"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

// BaseURLs defines per-service base endpoints.
type BaseURLs struct {
	CLOB     string
	CLOBWS   string
	Geoblock string
	Gamma    string
	Data     string
	Bridge   string
	RTDS     string
	CTF      string
}

// Config holds shared SDK configuration.
type Config struct {
	BaseURLs      BaseURLs
	HTTPClient    transport.Doer
	UserAgent     string
	Timeout       time.Duration
	UseServerTime bool
	CLOBWSConfig  ws.ClientConfig
	RTDSConfig    rtds.ClientConfig
}

// DefaultConfig returns default service endpoints.
func DefaultConfig() Config {
	return Config{
		BaseURLs: BaseURLs{
			CLOB:     "https://clob.polymarket.com",
			CLOBWS:   "wss://ws-subscriptions-clob.polymarket.com",
			Geoblock: "https://polymarket.com",
			Gamma:    "https://gamma-api.polymarket.com",
			Data:     "https://data-api.polymarket.com",
			Bridge:   "https://bridge.polymarket.com",
			RTDS:     "wss://ws-live-data.polymarket.com",
			CTF:      "",
		},
		UserAgent:     "github.com/GoPolymarket/polymarket-go-sdk",
		Timeout:       30 * time.Second,
		UseServerTime: false,
		// Keep legacy env-driven behavior for backward compatibility at the root client level.
		CLOBWSConfig: ws.ClientConfigFromEnv(),
		RTDSConfig:   rtds.ClientConfigFromEnv(),
	}
}
