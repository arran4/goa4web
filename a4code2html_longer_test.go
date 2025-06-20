package main

import "testing"

func TestA4code2htmlComplex(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[b Bold [i Italic]] plain [link http://x example]"
	c.Process()
	want := "<strong>Bold <i>Italic</i></strong> plain <a href=\"http://x\" target=\"_BLANK\">example</a>"
	if got := c.Output(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestA4code2htmlUnclosed(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[b bold"
	c.Process()
	want := "<strong>bold</strong>"
	if got := c.Output(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
