package sign

import (
	"fmt"
	"testing"
	"time"
)

func TestSignAndVerifyNoExpiry(t *testing.T) {
	s := &Signer{Key: "k"}
	_, sig := s.Sign("data", time.Unix(0, 0))
	if !s.Verify("data", "0", sig) {
		t.Fatalf("verify failed for no expiry")
	}
}

func TestSignCustomExpiry(t *testing.T) {
	s := &Signer{Key: "k"}
	exp := time.Now().Add(2 * time.Hour)
	ts, sig := s.Sign("x", exp)
	if !s.Verify("x", fmt.Sprintf("%d", ts), sig) {
		t.Fatalf("verify failed")
	}
}
