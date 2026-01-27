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

func TestQuoteHTML(t *testing.T) {
	input := "[quote Outer [quote Nested]]"
	ast, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(ast)
	want := `<blockquote class="a4code-block a4code-quote quote-color-0"><div class="quote-body"> Outer <blockquote class="a4code-block a4code-quote quote-color-1"><div class="quote-body"> Nested</div></blockquote></div></blockquote>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestQuoteOfHTML(t *testing.T) {
	input := `[quoteof "User" Outer [quoteof "User2" Nested]]`
	ast, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(ast)
	want := `<blockquote class="a4code-block a4code-quoteof quote-color-0"><div class="quote-header">Quote of User:</div><div class="quote-body"> Outer <blockquote class="a4code-block a4code-quoteof quote-color-1"><div class="quote-header">Quote of User2:</div><div class="quote-body"> Nested</div></blockquote></div></blockquote>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
