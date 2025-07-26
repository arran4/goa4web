package images

import (
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
