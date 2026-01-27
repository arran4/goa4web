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
	s := "This is a [b]long[/b] string"
	if SnipText(s, 10) != "This is a..." {
		t.Errorf("SnipText failed: %s", SnipText(s, 10))
	}
	if SnipText("[b]"+s+"[/b]", 10) != "This is a..." {
		t.Errorf("SnipText failed: %s", SnipText("[b]"+s+"[/b]", 10))
	}
}
