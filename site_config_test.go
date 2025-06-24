package goa4web

import "testing"

func TestResolveFeedsEnabledPrecedence(t *testing.T) {
	if !resolveFeedsEnabled("", "", "") {
		t.Fatalf("default should be true")
	}
	if resolveFeedsEnabled("", "", "0") {
		t.Fatalf("env false")
	}
	if resolveFeedsEnabled("", "0", "1") {
		t.Fatalf("file overrides env")
	}
	if !resolveFeedsEnabled("1", "0", "0") {
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
