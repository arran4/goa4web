package a4code

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/a4code/ast"
)

func TestParseToHTML(t *testing.T) {
	input := "[b Bold [i Italic]] plain"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// New behavior: consume space after tag.
	// [b Bold [i Italic]] -> Bold node ("Bold "), Italic node ("Italic").
	// plain -> " plain".
	want := `<strong data-start-pos="0" data-end-pos="11"><span data-start-pos="0" data-end-pos="5">Bold </span><i data-start-pos="5" data-end-pos="11"><span data-start-pos="5" data-end-pos="11">Italic</span></i></strong><span data-start-pos="11" data-end-pos="17"> plain</span>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestParseImage(t *testing.T) {
	input := "[img=image.jpg]"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// [img] is 0-width in visible space.
	want := `<img src="image.jpg" data-start-pos="0" data-end-pos="0" />`
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
	root := &ast.Root{Children: nodes}
	got := ToHTML(root)
	// [b foo] -> "foo"
	// [i bar] -> "bar"
	want := `<strong data-start-pos="0" data-end-pos="3"><span data-start-pos="0" data-end-pos="3">foo</span></strong><i data-start-pos="3" data-end-pos="6"><span data-start-pos="3" data-end-pos="6">bar</span></i>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestOffsets(t *testing.T) {
	// [code]foo[/code]
	// vis 0-3.
	// Inner content: foo.
	input := "[code]foo[/code]"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	want := `<pre class="a4code-block a4code-code" data-start-pos="0" data-end-pos="3"><span data-start-pos="0" data-end-pos="3">foo</span></pre>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestQuoteHTML(t *testing.T) {
	input := "[quote Outer [quote Nested]]"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	want := `<blockquote class="a4code-block a4code-quote quote-color-0" data-start-pos="0" data-end-pos="12"><div class="quote-body"><span data-start-pos="0" data-end-pos="6">Outer </span><blockquote class="a4code-block a4code-quote quote-color-1" data-start-pos="6" data-end-pos="12"><div class="quote-body"><span data-start-pos="6" data-end-pos="12">Nested</span></div></blockquote></div></blockquote>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestQuoteOfHTML(t *testing.T) {
	input := `[quoteof "User" Outer [quoteof "User2" Nested]]`
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	want := `<blockquote class="a4code-block a4code-quoteof quote-color-0" data-start-pos="0" data-end-pos="12"><div class="quote-header">Quote of User:</div><div class="quote-body"><span data-start-pos="0" data-end-pos="6">Outer </span><blockquote class="a4code-block a4code-quoteof quote-color-1" data-start-pos="6" data-end-pos="12"><div class="quote-header">Quote of User2:</div><div class="quote-body"><span data-start-pos="6" data-end-pos="12">Nested</span></div></blockquote></div></blockquote>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
