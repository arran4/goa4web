package adminapi

import (
	"fmt"
	"testing"
)

func TestSignVerify(t *testing.T) {
	signer := NewSigner("k")
	ts, sig := signer.Sign("POST", "/admin/api/shutdown")
	if !signer.Verify("POST", "/admin/api/shutdown", fmt.Sprint(ts), sig) {
		t.Fatalf("signature should verify")
	}
	if signer.Verify("GET", "/admin/api/shutdown", fmt.Sprint(ts), sig) {
		t.Fatalf("unexpected verify for wrong method")
	}
}
