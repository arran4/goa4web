package handlers

import (
	"testing"
)

func TestHashSessionID(t *testing.T) {
	h1 := HashSessionID("abc123")
	h2 := HashSessionID("abc123")
	if h1 != h2 {
		t.Fatalf("expected deterministic hash got %s and %s", h1, h2)
	}
	if len(h1) != 12 {
		t.Fatalf("expected 12 char hash got %d", len(h1))
	}
	if h1 == "abc123" {
		t.Fatalf("hash should not equal original")
	}
	if HashSessionID("") != "" {
		t.Fatalf("expected empty result for empty ID")
	}
}
