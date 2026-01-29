package polymarket

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BaseURLs.CLOB == "" {
		t.Errorf("default CLOB URL empty")
	}
}
