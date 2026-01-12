package sign

import (
	"fmt"
	"testing"
	"time"
)

func TestSignAndVerifyNoExpiry(t *testing.T) {
	s := &Signer{Key: "k"}
	sig := s.Sign("data", WithExpiryTime(time.Unix(0, 0)))
	if valid, err := s.Verify("data", sig, WithExpiryTimestamp("0")); !valid || err != nil {
		t.Fatalf("verify failed for no expiry: %v", err)
	}
}

func TestSignCustomExpiry(t *testing.T) {
	s := &Signer{Key: "k"}
	exp := time.Now().Add(2 * time.Hour)
	sig := s.Sign("x", WithExpiryTime(exp))
	if valid, err := s.Verify("x", sig, WithExpiryTimestamp(fmt.Sprintf("%d", exp.Unix()))); !valid || err != nil {
		t.Fatalf("verify failed: %v", err)
	}
}
