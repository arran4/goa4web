package main

import "testing"

func TestA4code2htmlEscape(t *testing.T) {
	c := NewA4Code2HTML()
	if got := c.Escape('&'); got != "&amp;" {
		t.Errorf("amp %q", got)
	}
	if got := c.Escape('<'); got != "&lt;" {
		t.Errorf("lt %q", got)
	}
	if got := c.Escape('>'); got != "&gt;" {
		t.Errorf("gt %q", got)
	}
	if got := c.Escape('\n'); got != "<br />\n" {
		t.Errorf("newline %q", got)
	}
	c.codeType = ct_tagstrip
	if got := c.Escape('\n'); got != "\n" {
		t.Errorf("tagstrip %q", got)
	}
	c.codeType = ct_wordsonly
	if got := c.Escape('&'); got != " " {
		t.Errorf("wordsonly %q", got)
	}
}
