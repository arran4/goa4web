package sign

import (
	"fmt"
	"testing"
	"time"
)

func TestSignAndVerifyNoExpiry(t *testing.T) {
	s := &Signer{Key: "k"}
	_, sig := s.Sign("data", WithExpiry(time.Unix(0, 0)))
	if valid, err := s.Verify("data", sig, WithExpiryTimestamp("0")); !valid || err != nil {
		t.Fatalf("verify failed for no expiry: %v", err)
	}
}

func TestSignCustomExpiry(t *testing.T) {
	s := &Signer{Key: "k"}
	exp := time.Now().Add(2 * time.Hour)
	ts, sig := s.Sign("x", WithExpiry(exp))
	if valid, err := s.Verify("x", sig, WithExpiryTimestamp(fmt.Sprintf("%d", ts))); !valid || err != nil {
		t.Fatalf("verify failed: %v", err)
	}
}
