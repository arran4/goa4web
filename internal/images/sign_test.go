package images

import "testing"

func TestSignedAndVerifyRef(t *testing.T) {
	SetSigningKey("k")
	ref := SignedRef("image:abc")
	if ref == "image:abc" {
		t.Fatal("signature not added")
	}
	clean, ok := VerifyRef(ref)
	if !ok {
		t.Fatal("verify failed")
	}
	if clean != "image:abc" {
		t.Fatalf("got %s", clean)
	}
}

func TestMapURLUploading(t *testing.T) {
	SetSigningKey("k")
	got := MapURL("img", "uploading:abc")
	if got != "uploading:abc" {
		t.Fatalf("expected placeholder unchanged, got %s", got)
	}
}
