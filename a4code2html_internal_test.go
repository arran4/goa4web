package main

import "testing"

func TestGetNext(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "abc]"
	got := c.getNext(false)
	if got != "abc" {
		t.Fatalf("got %q", got)
	}
	if c.input != "" {
		t.Fatalf("remaining %q", c.input)
	}
}

func TestDirectOutput(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "fooENDbar"
	c.directOutput("END")
	if got := c.output.String(); got != "foo" {
		t.Fatalf("got %q", got)
	}
	if c.input != "bar" {
		t.Fatalf("input %q", c.input)
	}
}

func TestNextcommSimple(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[*]"
	c.Process()
	if got := c.Output(); got != "<strong></strong>" {
		t.Fatalf("output %q", got)
	}
}
