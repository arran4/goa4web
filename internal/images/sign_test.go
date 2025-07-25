package images

import (
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestMapURLUploading(t *testing.T) {
	SetSigningKey("k")
	got := MapURL("img", "uploading:abc", config.RuntimeConfig{})
	if got != "uploading:abc" {
		t.Fatalf("expected placeholder unchanged, got %s", got)
	}
}
