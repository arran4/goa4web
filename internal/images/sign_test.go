package images

import "testing"

func TestMapURLUploading(t *testing.T) {
	SetSigningKey("k")
	got := MapURL("img", "uploading:abc")
	if got != "uploading:abc" {
		t.Fatalf("expected placeholder unchanged, got %s", got)
	}
}
