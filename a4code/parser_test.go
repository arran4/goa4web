package a4code

import (
	"strings"
	"testing"
)

func TestParseToHTML(t *testing.T) {
	input := "[b Bold [i Italic]] plain"
	ast, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(ast)
	want := "<strong> Bold <i> Italic</i></strong> plain"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestParseImage(t *testing.T) {
	ast, err := Parse(strings.NewReader("[img=image.jpg]"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(ast)
	want := "<img src=\"image.jpg\" />"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	input := "[b Hello [i world]]"
	root, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	markup := ToA4Code(root)
	root2, err := Parse(strings.NewReader(markup))
	if err != nil {
		t.Fatalf("reparse error: %v", err)
	}
	html1 := ToHTML(root)
	html2 := ToHTML(root2)
	if html1 != html2 {
		t.Errorf("round trip mismatch: %q vs %q", html1, html2)
	}
}

func TestParseNodes(t *testing.T) {
	input := "[b foo][i bar]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse nodes error: %v", err)
	}
	root := &Root{Children: nodes}
	got := ToHTML(root)
	want := "<strong> foo</strong><i> bar</i>"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
