package clob

import (
	"context"
	"sync"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/cloberrors"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/heartbeat"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/rfq"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

// clientImpl implements the Client interface.
type clientImpl struct {
	httpClient     *transport.Client
	signer         auth.Signer
	apiKey         *auth.APIKey
	builderCfg     *auth.BuilderConfig
	cache          *clientCache
	geoblockHost   string
	geoblockClient *transport.Client
	rfq            rfq.Client
	ws             ws.Client
	heartbeat      heartbeat.Client
}

type clientCache struct {
	mu        sync.RWMutex
	tickSizes map[string]string
	feeRates  map[string]int64
	negRisk   map[string]bool
}

func newClientCache() *clientCache {
	return &clientCache{
		tickSizes: make(map[string]string),
		feeRates:  make(map[string]int64),
		negRisk:   make(map[string]bool),
	}
}

// NewClient creates a new CLOB client.
func NewClient(httpClient *transport.Client) Client {
	return NewClientWithGeoblock(httpClient, "")
}

// NewClientWithGeoblock creates a new CLOB client with an explicit geoblock host.
func NewClientWithGeoblock(httpClient *transport.Client, geoblockHost string) Client {
	if geoblockHost == "" {
		geoblockHost = DefaultGeoblockHost
	}

	c := &clientImpl{
		httpClient:     httpClient,
		cache:          newClientCache(),
		geoblockHost:   geoblockHost,
		geoblockClient: nil,
		// builderCfg is nil by default (Opt-in)
		rfq:            rfq.NewClient(httpClient),
		heartbeat:      heartbeat.NewClient(httpClient),
	}
	if httpClient != nil {
		c.geoblockClient = httpClient.CloneWithBaseURL(geoblockHost)
	}
	return c
}

func (c *clientImpl) RFQ() rfq.Client {
	return c.rfq
}

func (c *clientImpl) WS() ws.Client {
	return c.ws
}

func (c *clientImpl) Heartbeat() heartbeat.Client {
	return c.heartbeat
}

// WithAuth returns a new Client with the provided signer and API credentials.
func (c *clientImpl) WithAuth(signer auth.Signer, apiKey *auth.APIKey) Client {
	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         signer,
		apiKey:         apiKey,
		builderCfg:     c.builderCfg,
		cache:          c.cache,
		geoblockHost:   c.geoblockHost,
		geoblockClient: c.geoblockClient,
		rfq:            c.rfq,
		ws:             c.ws,
		heartbeat:      c.heartbeat,
	}
}

// WithBuilderConfig sets the builder attribution config.
func (c *clientImpl) WithBuilderConfig(config *auth.BuilderConfig) Client {
	// If config is nil, we might want to disable it or revert to default.
	// For now, let's assume the user knows what they are doing.
	// We also need to update the underlying transport's builder config.
	if c.httpClient != nil {
		c.httpClient.SetBuilderConfig(config)
	}
	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         c.signer,
		apiKey:         c.apiKey,
		builderCfg:     config,
		cache:          c.cache,
		geoblockHost:   c.geoblockHost,
		geoblockClient: c.geoblockClient,
		rfq:            c.rfq,
		ws:             c.ws,
		heartbeat:      c.heartbeat,
	}
}

// WithUseServerTime configures the transport to use server time for timestamps.
func (c *clientImpl) WithUseServerTime(use bool) Client {
	if c.httpClient != nil {
		c.httpClient.SetUseServerTime(use)
	}
	return c
}

// WithGeoblockHost sets the geoblock host.
func (c *clientImpl) WithGeoblockHost(host string) Client {
	newC := &clientImpl{
		httpClient:     c.httpClient,
		signer:         c.signer,
		apiKey:         c.apiKey,
		builderCfg:     c.builderCfg,
		cache:          c.cache,
		geoblockHost:   host,
		geoblockClient: nil,
		rfq:            c.rfq,
		ws:             c.ws,
		heartbeat:      c.heartbeat,
	}
	if c.httpClient != nil {
		newC.geoblockClient = c.httpClient.CloneWithBaseURL(host)
	}
	return newC
}

// WithWS sets the WebSocket client and returns a new client.
func (c *clientImpl) WithWS(ws ws.Client) Client {
	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         c.signer,
		apiKey:         c.apiKey,
		builderCfg:     c.builderCfg,
		cache:          c.cache,
		geoblockHost:   c.geoblockHost,
		geoblockClient: c.geoblockClient,
		rfq:            c.rfq,
		ws:             ws,
		heartbeat:      c.heartbeat,
	}
}

func (c *clientImpl) Health(ctx context.Context) (string, error) {
	var resp struct {
		Status string `json:"status"`
	}
	err := c.httpClient.Get(ctx, "/time", nil, &resp)
	if err != nil {
		return "DOWN", err
	}
	return "UP", nil
}

func (c *clientImpl) Time(ctx context.Context) (clobtypes.TimeResponse, error) {
	var ts int64
	err := c.httpClient.Get(ctx, "/time", nil, &ts)
	if err != nil {
		return clobtypes.TimeResponse{}, err
	}
	return clobtypes.TimeResponse{Timestamp: ts}, nil
}

func (c *clientImpl) Geoblock(ctx context.Context) (clobtypes.GeoblockResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	geo := c.geoblockClient
	if geo == nil {
		host := c.geoblockHost
		if host == "" {
			host = DefaultGeoblockHost
		}
		geo = transport.NewClient(nil, host)
	}
	var resp clobtypes.GeoblockResponse
	err := geo.Get(ctx, "/api/geoblock", nil, &resp)
	return resp, err
}

func (c *clientImpl) InvalidateCaches() {
	if c.cache == nil {
		return
	}
	c.cache.mu.Lock()
	c.cache.tickSizes = make(map[string]string)
	c.cache.feeRates = make(map[string]int64)
	c.cache.negRisk = make(map[string]bool)
	c.cache.mu.Unlock()
}

func (c *clientImpl) SetTickSize(tokenID, tickSize string) {
	if c.cache == nil || tokenID == "" || tickSize == "" {
		return
	}
	c.cache.mu.Lock()
	c.cache.tickSizes[tokenID] = tickSize
	c.cache.mu.Unlock()
}

func (c *clientImpl) SetNegRisk(tokenID string, negRisk bool) {
	if c.cache == nil || tokenID == "" {
		return
	}
	c.cache.mu.Lock()
	c.cache.negRisk[tokenID] = negRisk
	c.cache.mu.Unlock()
}

func (c *clientImpl) SetFeeRateBps(tokenID string, feeRateBps int64) {
	if c.cache == nil || tokenID == "" || feeRateBps <= 0 {
		return
	}
	c.cache.mu.Lock()
	c.cache.feeRates[tokenID] = feeRateBps
	c.cache.mu.Unlock()
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if apiErr, ok := err.(*types.Error); ok {
		return cloberrors.FromTypeErr(apiErr)
	}
	return err
}
