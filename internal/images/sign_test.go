package images

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestMapURLUploading(t *testing.T) {
	signer := NewSigner(&config.RuntimeConfig{}, "k")
	got := signer.MapURL("img", "uploading:abc")
	if got != "uploading:abc" {
		t.Fatalf("expected placeholder unchanged, got %s", got)
	}
}

func TestMapURLImage(t *testing.T) {
	cfg := &config.RuntimeConfig{HTTPHostname: "http://mysite"}
	signer := NewSigner(cfg, "k")
	got := signer.MapURL("img", "image:foo")
	if got == "image:foo" || !strings.Contains(got, "/images/image/") {
		t.Fatalf("expected signed image url, got %s", got)
	}
}
