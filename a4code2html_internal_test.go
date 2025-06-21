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

func TestGetNextSpecialChars(t *testing.T) {
	specials := []byte{'*', '/', '_'}
	for _, ch := range specials {
		c := NewA4Code2HTML()
		c.input = string(ch) + "]"
		if got := c.getNext(true); got != string(ch) || c.input != "" {
			t.Fatalf("char %q got %q remaining %q", ch, got, c.input)
		}

		c = NewA4Code2HTML()
		c.input = "\\" + string(ch) + "]"
		if got := c.getNext(false); got != string(ch) || c.input != "" {
			t.Fatalf("escape %q got %q remaining %q", ch, got, c.input)
		}
	}
}

func TestProcessSpecialCommands(t *testing.T) {
	tests := []struct {
		cmd  string
		want string
	}{
		{"*", "<strong>text</strong>"},
		{"/", "<i>text</i>"},
		{"_", "<u>text</u>"},
	}
	for _, tt := range tests {
		c := NewA4Code2HTML()
		c.input = "[" + tt.cmd + " text]"
		c.Process()
		if got := c.Output(); got != tt.want {
			t.Fatalf("cmd %q got %q", tt.cmd, got)
		}
	}
}

func TestProcessEscapedSpecialChars(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "start \\* mid \\/ end \\_"
	c.Process()
	if got := c.Output(); got != "start * mid / end _" {
		t.Fatalf("got %q", got)
	}
}
