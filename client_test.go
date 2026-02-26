package polymarket

import (
	"errors"
	"testing"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
)

func invalidStreamingConfig() Config {
	cfg := DefaultConfig()
	cfg.BaseURLs.CLOBWS = "://invalid-ws-url"
	cfg.BaseURLs.RTDS = "://invalid-rtds-url"
	return cfg
}

type cloneableWSStub struct {
	ws.Client
	authCalls  int
	cloneCalls int
}

func (s *cloneableWSStub) Authenticate(signer auth.Signer, apiKey *auth.APIKey) ws.Client {
	s.authCalls++
	return s
}

func (s *cloneableWSStub) Clone() ws.Client {
	s.cloneCalls++
	return &cloneableWSStub{Client: s.Client}
}

type nonCloneableWSStub struct {
	ws.Client
	authCalls int
}

func (s *nonCloneableWSStub) Authenticate(signer auth.Signer, apiKey *auth.APIKey) ws.Client {
	s.authCalls++
	return s
}

func TestDefaultConfigRespectsLegacyStreamingEnv(t *testing.T) {
	t.Setenv("CLOB_WS_RECONNECT_MAX", "17")
	t.Setenv("RTDS_WS_PING_INTERVAL_MS", "2300")

	cfg := DefaultConfig()
	if cfg.CLOBWSConfig.ReconnectMax != 17 {
		t.Fatalf("expected CLOB WS reconnect max from env, got %d", cfg.CLOBWSConfig.ReconnectMax)
	}
	if cfg.RTDSConfig.PingInterval != 2300*time.Millisecond {
		t.Fatalf("expected RTDS ping interval from env, got %s", cfg.RTDSConfig.PingInterval)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	cfg := invalidStreamingConfig()
	c := NewClient(
		WithConfig(cfg),
		WithUseServerTime(true),
		WithUserAgent("test-ua"),
		WithCLOB(nil),
		WithGamma(nil),
		WithData(nil),
		WithBridge(nil),
		WithRTDS(nil),
		WithCTF(nil),
	)
	if c.Config.UserAgent != "test-ua" {
		t.Errorf("WithUserAgent failed")
	}
	if !c.Config.UseServerTime {
		t.Errorf("WithUseServerTime failed")
	}
}

func TestAttributionOptions(t *testing.T) {
	cfg := invalidStreamingConfig()
	_ = NewClient(
		WithConfig(cfg),
		WithBuilderConfig(nil),
		WithOfficialGoSDKSupport(),
		WithBuilderAttribution("key", "secret", "pass"),
	)
}

func TestNewClientEReturnsInitError(t *testing.T) {
	cfg := invalidStreamingConfig()

	client, err := NewClientE(WithConfig(cfg))
	if client == nil {
		t.Fatalf("expected client even when init fails")
	}
	if err == nil {
		t.Fatalf("expected initialization error")
	}
	if len(client.InitErrors) == 0 {
		t.Fatalf("expected InitErrors to be populated")
	}
	var initErr *InitError
	if !errors.As(err, &initErr) {
		t.Fatalf("expected joined error to contain InitError, got %T", err)
	}
	foundRTDS := false
	foundWS := false
	for _, e := range client.InitErrors {
		ie, ok := e.(*InitError)
		if !ok {
			continue
		}
		switch ie.Component {
		case "rtds":
			foundRTDS = true
		case "clob_ws":
			foundWS = true
		}
	}
	if !foundRTDS || !foundWS {
		t.Fatalf("expected both RTDS and CLOB WS init errors, got %#v", client.InitErrors)
	}
}

func TestNewClientCollectsInitErrorsWithoutFailing(t *testing.T) {
	cfg := invalidStreamingConfig()

	client := NewClient(WithConfig(cfg))
	if client == nil {
		t.Fatalf("expected client")
	}
	if len(client.InitErrors) == 0 {
		t.Fatalf("expected InitErrors to capture initialization failure")
	}
}

func TestWithAuthDoesNotMutateOriginalClient(t *testing.T) {
	cfg := invalidStreamingConfig()

	base := NewClient(WithConfig(cfg))
	origCLOB := base.CLOB

	signer, err := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	next := base.WithAuth(signer, &auth.APIKey{Key: "k", Secret: "s", Passphrase: "p"})

	if next == base {
		t.Fatalf("WithAuth should return a new root client")
	}
	if base.CLOB != origCLOB {
		t.Fatalf("WithAuth should not mutate original root client CLOB")
	}
	if next.CLOB == origCLOB {
		t.Fatalf("WithAuth should attach auth on a new CLOB client instance")
	}
}

func TestWithAuthClonesWSClientWhenSupported(t *testing.T) {
	cfg := invalidStreamingConfig()
	base := NewClient(WithConfig(cfg))
	origWS := &cloneableWSStub{}
	base.CLOBWS = origWS

	signer, err := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	next := base.WithAuth(signer, &auth.APIKey{Key: "k", Secret: "s", Passphrase: "p"})
	if next == nil || next.CLOBWS == nil {
		t.Fatalf("expected cloned ws client")
	}
	if origWS.authCalls != 0 {
		t.Fatalf("original ws client should not be authenticated in place")
	}
	if origWS.cloneCalls != 1 {
		t.Fatalf("expected original ws client clone path to be used")
	}
	cloned, ok := next.CLOBWS.(*cloneableWSStub)
	if !ok {
		t.Fatalf("expected cloned ws stub type, got %T", next.CLOBWS)
	}
	if cloned == origWS {
		t.Fatalf("expected a distinct ws client instance")
	}
	if cloned.authCalls != 1 {
		t.Fatalf("expected cloned ws client to receive auth")
	}
}

func TestWithAuthLeavesWSUntouchedWhenCloneUnavailable(t *testing.T) {
	cfg := invalidStreamingConfig()
	base := NewClient(WithConfig(cfg))
	origWS := &nonCloneableWSStub{}
	base.CLOBWS = origWS

	signer, err := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	next := base.WithAuth(signer, &auth.APIKey{Key: "k", Secret: "s", Passphrase: "p"})
	if next == nil || next.CLOBWS == nil {
		t.Fatalf("expected ws client")
	}
	if origWS.authCalls != 0 {
		t.Fatalf("non-cloneable ws should not be mutated in place")
	}
	if next.CLOBWS != origWS {
		t.Fatalf("expected ws client to remain unchanged when clone is unavailable")
	}
}
