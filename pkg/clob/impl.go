package clob

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/cloberrors"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
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
	signatureType  auth.SignatureType
	authNonce      *int64
	funder         *types.Address
	saltGenerator  SaltGenerator
	cache          *clientCache
	geoblockHost   string
	geoblockClient *transport.Client
	rfq            rfq.Client
	ws             ws.Client
	heartbeat      heartbeat.Client

	heartbeatInterval time.Duration
	heartbeatStop     chan struct{}
	heartbeatMu       sync.Mutex
}

type clientCache struct {
	mu        sync.RWMutex
	tickSizes map[string]float64
	feeRates  map[string]int64
	negRisk   map[string]bool
}

type orderDefaults struct {
	signatureType auth.SignatureType
	funder        *types.Address
	saltGenerator SaltGenerator
}

func newClientCache() *clientCache {
	return &clientCache{
		tickSizes: make(map[string]float64),
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

	if httpClient == nil {
		httpClient = transport.NewClient(nil, BaseURL)
	}

	c := &clientImpl{
		httpClient:     httpClient,
		cache:          newClientCache(),
		geoblockHost:   geoblockHost,
		geoblockClient: nil,
		signatureType:  auth.SignatureEOA,
		authNonce:      nil,
		funder:         nil,
		saltGenerator:  nil,
		// builderCfg is nil by default (Opt-in)
		rfq:       rfq.NewClient(httpClient),
		heartbeat: heartbeat.NewClient(httpClient),
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
	if c.httpClient != nil {
		c.httpClient.SetAuth(signer, apiKey)
	}
	newC := &clientImpl{
		httpClient:        c.httpClient,
		signer:            signer,
		apiKey:            apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
	newC.startHeartbeats()
	return newC
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
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        config,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
}

// PromoteToBuilder switches the client into builder attribution mode.
func (c *clientImpl) PromoteToBuilder(config *auth.BuilderConfig) Client {
	if config == nil {
		return c
	}
	// Stop heartbeats on the old instance before switching.
	c.StopHeartbeats()
	if c.httpClient != nil {
		c.httpClient.SetBuilderConfig(config)
	}
	newC := &clientImpl{
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        config,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
	newC.startHeartbeats()
	return newC
}

// WithSignatureType sets the default signature type for order signing and balance/rewards queries.
func (c *clientImpl) WithSignatureType(sigType auth.SignatureType) Client {
	return &clientImpl{
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     sigType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
}

// WithAuthNonce sets the default nonce used when creating/deriving API keys.
func (c *clientImpl) WithAuthNonce(nonce int64) Client {
	return &clientImpl{
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     c.signatureType,
		authNonce:         &nonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
}

// WithFunder sets the default funder (maker) address used for order creation.
func (c *clientImpl) WithFunder(funder types.Address) Client {
	return &clientImpl{
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            &funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
}

// WithSaltGenerator sets the default salt generator for new orders.
func (c *clientImpl) WithSaltGenerator(gen SaltGenerator) Client {
	return &clientImpl{
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     gen,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
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
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      host,
		geoblockClient:    nil,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
	if c.httpClient != nil {
		newC.geoblockClient = c.httpClient.CloneWithBaseURL(host)
	}
	return newC
}

// WithWS sets the WebSocket client and returns a new client.
func (c *clientImpl) WithWS(ws ws.Client) Client {
	return &clientImpl{
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: c.heartbeatInterval,
	}
}

func (c *clientImpl) WithHeartbeatInterval(interval time.Duration) Client {
	newC := &clientImpl{
		httpClient:        c.httpClient,
		signer:            c.signer,
		apiKey:            c.apiKey,
		builderCfg:        c.builderCfg,
		signatureType:     c.signatureType,
		authNonce:         c.authNonce,
		funder:            c.funder,
		saltGenerator:     c.saltGenerator,
		cache:             c.cache,
		geoblockHost:      c.geoblockHost,
		geoblockClient:    c.geoblockClient,
		rfq:               c.rfq,
		ws:                c.ws,
		heartbeat:         c.heartbeat,
		heartbeatInterval: interval,
	}
	newC.startHeartbeats()
	return newC
}

func (c *clientImpl) orderDefaults() orderDefaults {
	return orderDefaults{
		signatureType: c.signatureType,
		funder:        c.funder,
		saltGenerator: c.saltGenerator,
	}
}

func (c *clientImpl) StopHeartbeats() {
	c.heartbeatMu.Lock()
	defer c.heartbeatMu.Unlock()
	if c.heartbeatStop != nil {
		close(c.heartbeatStop)
		c.heartbeatStop = nil
	}
}

func (c *clientImpl) startHeartbeats() {
	c.heartbeatMu.Lock()
	defer c.heartbeatMu.Unlock()
	if c.heartbeatInterval <= 0 {
		return
	}
	if c.httpClient == nil || c.signer == nil || c.apiKey == nil || c.heartbeat == nil {
		return
	}

	// Stop old heartbeat goroutine if it exists
	if c.heartbeatStop != nil {
		oldStop := c.heartbeatStop
		c.heartbeatStop = nil
		close(oldStop)
		// Wait for old goroutine to exit by using a done channel
		// This prevents race conditions from unlock/sleep/relock pattern
	}

	stop := make(chan struct{})
	c.heartbeatStop = stop
	interval := c.heartbeatInterval
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				_, _ = c.heartbeat.Heartbeat(context.Background(), nil)
			}
		}
	}()
}

func (c *clientImpl) Health(ctx context.Context) (string, error) {
	var resp healthResponse
	err := c.httpClient.Get(ctx, "/", nil, &resp)
	if err != nil {
		return "DOWN", mapError(err)
	}
	return string(resp), nil
}

type healthResponse string

func (h *healthResponse) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*h = ""
		return nil
	}
	var status string
	if err := json.Unmarshal(trimmed, &status); err == nil {
		*h = healthResponse(status)
		return nil
	}
	var payload struct {
		Status string `json:"status"`
		Data   string `json:"data"`
	}
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return err
	}
	if payload.Status != "" {
		*h = healthResponse(payload.Status)
		return nil
	}
	*h = healthResponse(payload.Data)
	return nil
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
	c.cache.tickSizes = make(map[string]float64)
	c.cache.feeRates = make(map[string]int64)
	c.cache.negRisk = make(map[string]bool)
	c.cache.mu.Unlock()
}

func (c *clientImpl) SetTickSize(tokenID string, tickSize float64) {
	if c.cache == nil || tokenID == "" {
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
