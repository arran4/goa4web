package runtimeconfig

import "testing"

func TestResolveBoolPrecedence(t *testing.T) {
	if !resolveBool(true, "", "", "") {
		t.Fatalf("default should be true")
	}
	if resolveBool(true, "", "", "0") {
		t.Fatalf("env false")
	}
	if resolveBool(true, "", "0", "1") {
		t.Fatalf("file overrides env")
	}
	if !resolveBool(true, "1", "0", "0") {
		t.Fatalf("cli overrides file and env")
	}
}

func TestParseBool(t *testing.T) {
	if b, ok := parseBool("yes"); !ok || !b {
		t.Fatalf("yes should be true")
	}
	if b, ok := parseBool("off"); !ok || b {
		t.Fatalf("off should be false")
	}
	if _, ok := parseBool(""); ok {
		t.Fatalf("empty should not be ok")
	}
}
