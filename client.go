package polymarket

import (
	"errors"
	"fmt"
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
	InitErrors []error
}

// InitError records a non-fatal client initialization failure for a sub-service.
type InitError struct {
	Component string
	Err       error
}

func (e *InitError) Error() string {
	return fmt.Sprintf("init %s client: %v", e.Component, e.Err)
}

func (e *InitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewClient creates a new root client with optional overrides.
func NewClient(opts ...Option) *Client {
	c, _ := newClient(false, opts...)
	return c
}

// NewClientE creates a new root client and returns an aggregated error if any sub-client fails to initialize.
func NewClientE(opts ...Option) (*Client, error) {
	return newClient(true, opts...)
}

func newClient(strict bool, opts ...Option) (*Client, error) {
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
		rtdsClient, err := rtds.NewClientWithConfig(rtdsURL, c.Config.RTDSConfig)
		if err != nil {
			c.InitErrors = append(c.InitErrors, &InitError{Component: "rtds", Err: err})
		} else {
			c.RTDS = rtdsClient
		}
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
		wsClient, err := ws.NewClientWithConfig(wsURL, nil, nil, c.Config.CLOBWSConfig)
		if err != nil {
			c.InitErrors = append(c.InitErrors, &InitError{Component: "clob_ws", Err: err})
		} else {
			c.CLOBWS = wsClient
		}
	}

	// 5. Apply builder attribution if configured
	if c.builderCfg != nil && c.CLOB != nil {
		c.CLOB = c.CLOB.WithBuilderConfig(c.builderCfg)
	}

	if strict && len(c.InitErrors) > 0 {
		return c, errors.Join(c.InitErrors...)
	}
	return c, nil
}

// WithAuth returns a new client with auth credentials applied to all sub-clients.
// For best WebSocket behavior, call this before opening WS subscriptions.
func (c *Client) WithAuth(signer auth.Signer, apiKey *auth.APIKey) *Client {
	if c == nil {
		return nil
	}

	type wsCloner interface {
		Clone() ws.Client
	}

	next := *c
	if c.CLOB != nil {
		next.CLOB = c.CLOB.WithAuth(signer, apiKey)
	}
	if c.CLOBWS != nil {
		if cloner, ok := c.CLOBWS.(wsCloner); ok {
			if cloned := cloner.Clone(); cloned != nil {
				next.CLOBWS = cloned.Authenticate(signer, apiKey)
			} else {
				next.CLOBWS = c.CLOBWS
			}
		} else {
			// Unknown implementations are left untouched to avoid in-place auth mutation.
			next.CLOBWS = c.CLOBWS
		}
	}
	return &next
}
