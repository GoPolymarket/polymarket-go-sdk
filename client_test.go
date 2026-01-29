package polymarket

import (
	"testing"
)

func TestNewClientWithOptions(t *testing.T) {
	c := NewClient(
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
	_ = NewClient(
		WithOfficialGoSDKSupport(),
		WithBuilderAttribution("key", "secret", "pass"),
	)
}
