package a4code

import (
	"testing"
)

func TestSnip(t *testing.T) {
	s := "This is a long string"
	if Snip(s, 10) != "This is a..." {
		t.Errorf("Snip failed")
	}
	if Snip(s, 100) != s {
		t.Errorf("Snip failed")
	}
	if Snip("", 10) != "" {
		t.Errorf("Snip failed")
	}
	if Snip("short", 10) != "short" {
		t.Errorf("Snip failed")
	}
	if Snip("exactexact", 10) != "exactexact" {
		t.Errorf("Snip failed")
	}
}

func TestSnipText(t *testing.T) {
	tests := []struct {
		input  string
		length int
		want   string
	}{
		{"This is a [b long] string", 10, "This is a..."},
		{"[b Bold] text", 4, "Bold..."},
		{"[b Bold] text", 3, "Bol..."},
		{"Short", 10, "Short"},
		{"[i Italic] and [b Bold]", 6, "Italic..."},
		{"[quote Ignored] content", 10, "Ignored co..."},
	}
	for _, tt := range tests {
		if got := SnipText(tt.input, tt.length); got != tt.want {
			t.Errorf("SnipText(%q, %d) = %q, want %q", tt.input, tt.length, got, tt.want)
		}
	}
}
