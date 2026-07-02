package a4code

import "testing"

func TestSubstringIncludesSelectedImage(t *testing.T) {
	got, err := Substring("[img=image.jpg]", 0, 1)
	if err != nil {
		t.Fatalf("Substring returned error: %v", err)
	}
	if got != "[img=image.jpg]" {
		t.Fatalf("Substring image = %q, want %q", got, "[img=image.jpg]")
	}
}

func TestSubstringIncludesImageBetweenText(t *testing.T) {
	got, err := Substring("a[img=image.jpg]b", 0, 3)
	if err != nil {
		t.Fatalf("Substring returned error: %v", err)
	}
	if got != "a[img=image.jpg]b" {
		t.Fatalf("Substring text image text = %q, want %q", got, "a[img=image.jpg]b")
	}
}
