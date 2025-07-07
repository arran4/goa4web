package runtimeconfig

import "testing"

func TestResolveSMTPStartTLSPrecedence(t *testing.T) {
	if !resolveSMTPStartTLS("", "", "") {
		t.Fatalf("default should be true")
	}
	if resolveSMTPStartTLS("", "", "0") {
		t.Fatalf("env false")
	}
	if resolveSMTPStartTLS("", "0", "1") {
		t.Fatalf("file overrides env")
	}
	if !resolveSMTPStartTLS("1", "0", "0") {
		t.Fatalf("cli overrides file and env")
	}
}
