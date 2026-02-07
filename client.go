package polymarket

import (
	"net/http"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/bridge"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/ctf"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/data"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/gamma"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/rtds"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

// Client aggregates service clients behind a shared configuration.
type Client struct {
	Config Config

	CLOB   clob.Client
	CLOBWS ws.Client
	Gamma  gamma.Client
	Data   data.Client
	Bridge bridge.Client
	RTDS   rtds.Client
	CTF    ctf.Client

	builderCfg *auth.BuilderConfig
}

// NewClient creates a new root client with optional overrides.
func NewClient(opts ...Option) *Client {
	// 1. Initialize with default configuration
	c := &Client{Config: DefaultConfig()}

	// 2. Apply Options (Config overrides)
	for _, opt := range opts {
		opt(c)
	}

	// 3. Ensure a default HTTP client with timeout if none was provided.
	if c.Config.HTTPClient == nil && c.Config.Timeout > 0 {
		c.Config.HTTPClient = &http.Client{Timeout: c.Config.Timeout}
	}

	// 4. Initialize default transports and clients (if not overridden)
	if c.CLOB == nil {
		clobTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.CLOB)
		clobTransport.SetUserAgent(c.Config.UserAgent)
		clobTransport.SetUseServerTime(c.Config.UseServerTime)
		c.CLOB = clob.NewClientWithGeoblock(clobTransport, c.Config.BaseURLs.Geoblock)
	}
	if c.Gamma == nil {
		gammaTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Gamma)
		gammaTransport.SetUserAgent(c.Config.UserAgent)
		c.Gamma = gamma.NewClient(gammaTransport)
	}
	if c.Data == nil {
		dataTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Data)
		dataTransport.SetUserAgent(c.Config.UserAgent)
		c.Data = data.NewClient(dataTransport)
	}
	if c.Bridge == nil {
		bridgeTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Bridge)
		bridgeTransport.SetUserAgent(c.Config.UserAgent)
		c.Bridge = bridge.NewClient(bridgeTransport)
	}
	if c.RTDS == nil {
		rtdsURL := c.Config.BaseURLs.RTDS
		if rtdsURL == "" {
			rtdsURL = rtds.ProdURL
		}
		c.RTDS, _ = rtds.NewClient(rtdsURL)
	}
	if c.CTF == nil {
		c.CTF = ctf.NewClient()
	}
	if c.CLOBWS == nil {
		// Default WS URL
		wsURL := c.Config.BaseURLs.CLOBWS
		if wsURL == "" {
			wsURL = ws.ProdBaseURL
		}
		c.CLOBWS, _ = ws.NewClient(wsURL, nil, nil)
	}

	// 5. Apply builder attribution if configured
	if c.builderCfg != nil && c.CLOB != nil {
		c.CLOB = c.CLOB.WithBuilderConfig(c.builderCfg)
	}

	return c
}

// WithAuth returns a new client with auth credentials applied to all sub-clients.
func (c *Client) WithAuth(signer auth.Signer, apiKey *auth.APIKey) *Client {
	if c.CLOB != nil {
		c.CLOB = c.CLOB.WithAuth(signer, apiKey)
	}
	if c.CLOBWS != nil {
		c.CLOBWS = c.CLOBWS.Authenticate(signer, apiKey)
	}
	return c
}
