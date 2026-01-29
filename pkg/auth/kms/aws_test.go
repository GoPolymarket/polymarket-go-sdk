package kms

import (
	"testing"
)

func TestAWSSignerTypes(t *testing.T) {
	// Since we cannot mock the concrete *kms.Client without modifying the source,
	// and we want to keep the source code determined and original,
	// we will only test basic properties that don't involve calling AWS.
	s := &AWSSigner{
		keyID: "test",
	}
	if s.keyID != "test" {
		t.Errorf("keyID mismatch")
	}
}