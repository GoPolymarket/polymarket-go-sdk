package polymarket

import (
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
}

// NewClient creates a new root client with optional overrides.
func NewClient(opts ...Option) *Client {
	// 1. Initialize with default configuration
	c := &Client{Config: DefaultConfig()}

	// 2. Initialize default transports
	// We need these base transports to be ready so Options can modify them
	clobTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.CLOB)
	clobTransport.SetUserAgent(c.Config.UserAgent)
	clobTransport.SetUseServerTime(c.Config.UseServerTime)

	gammaTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Gamma)
	gammaTransport.SetUserAgent(c.Config.UserAgent)

	dataTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Data)
	dataTransport.SetUserAgent(c.Config.UserAgent)

	bridgeTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Bridge)
	bridgeTransport.SetUserAgent(c.Config.UserAgent)

	// 3. Initialize default clients
	c.clob = clob.NewClientWithGeoblock(clobTransport, c.Config.BaseURLs.Geoblock)
	c.gamma = gamma.NewClient(gammaTransport)
	c.data = data.NewClient(dataTransport)
	c.bridge = bridge.NewClient(bridgeTransport)
	c.ctf = ctf.NewClient()
	
	// Default WS URL
	wsURL := c.Config.BaseURLs.CLOBWS
	if wsURL == "" {
		wsURL = ws.ProdBaseURL
	}
	c.clobws, _ = ws.NewClient(wsURL, nil, nil)

	// 4. Apply Options (Overrides)
	// Now that clients exist, options like WithBuilderAttribution can modify them
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithAuth returns a new client with auth credentials applied to all sub-clients.
func (c *Client) WithAuth(signer auth.Signer, apiKey *auth.APIKey) *Client {
	if c.CLOB != nil {
		c.CLOB = c.CLOB.WithAuth(signer, apiKey)
	}
	return c
}
