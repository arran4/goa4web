package adminapi

import "fmt"
import "testing"

func TestSignVerify(t *testing.T) {
	SetSigningKey("k")
	ts, sig := Sign("POST", "/admin/api/shutdown")
	if !Verify("POST", "/admin/api/shutdown", fmt.Sprint(ts), sig) {
		t.Fatalf("signature should verify")
	}
	if Verify("GET", "/admin/api/shutdown", fmt.Sprint(ts), sig) {
		t.Fatalf("unexpected verify for wrong method")
	}
}
