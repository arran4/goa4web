package linksign

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestMapURLExternalLinkSigned(t *testing.T) {
	cfg := &config.RuntimeConfig{HTTPHostname: "http://mysite"}
	signer := NewSigner(cfg, "k", 0)
	got := signer.MapURL("a", "https://example.com/foo")
	if got == "https://example.com/foo" || !strings.Contains(got, "/goto?u=") {
		t.Fatalf("expected signed redirect, got %s", got)
	}
}

func TestMapURLInternalLinkUnchanged(t *testing.T) {
	cfg := &config.RuntimeConfig{HTTPHostname: "http://mysite"}
	signer := NewSigner(cfg, "k", 0)
	got := signer.MapURL("a", "http://mysite/bar")
	if got != "http://mysite/bar" {
		t.Fatalf("expected unchanged link, got %s", got)
	}
}
