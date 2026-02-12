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
	// [b (vis 0)
	//  Bold (space + Bold + space = 6). vis 0-6.
	// [i (vis 6)
	//  Italic (space + Italic = 7). vis 6-13.
	// ] (vis 13)
	// ] (vis 13)
	//  plain (space + plain = 6). vis 13-19.
	want := `<strong data-start-pos="0" data-end-pos="13"><span data-start-pos="0" data-end-pos="6"> Bold </span><i data-start-pos="6" data-end-pos="13"><span data-start-pos="6" data-end-pos="13"> Italic</span></i></strong><span data-start-pos="13" data-end-pos="19"> plain</span>`
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
	// [b foo] -> " foo" (4)
	// [i bar] -> " bar" (4)
	want := `<strong data-start-pos="0" data-end-pos="4"><span data-start-pos="0" data-end-pos="4"> foo</span></strong><i data-start-pos="4" data-end-pos="8"><span data-start-pos="4" data-end-pos="8"> bar</span></i>`
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
	// Outer quote starts at 0 (newline/start) and ends at EOF -> Block
	// Inner quote starts after space -> Inline
	want := `<blockquote class="a4code-block a4code-quote quote-color-0" data-start-pos="0" data-end-pos="14"><div class="quote-body"><span data-start-pos="0" data-end-pos="7"> Outer </span><q class="a4code-inline a4code-quote" data-start-pos="7" data-end-pos="14"><span data-start-pos="7" data-end-pos="14"> Nested</span></q></div></blockquote>`
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
	want := `<blockquote class="a4code-block a4code-quoteof quote-color-0" data-start-pos="0" data-end-pos="14"><div class="quote-header">Quote of User:</div><div class="quote-body"><span data-start-pos="0" data-end-pos="7"> Outer </span><blockquote class="a4code-block a4code-quoteof quote-color-1" data-start-pos="7" data-end-pos="14"><div class="quote-header">Quote of User2:</div><div class="quote-body"><span data-start-pos="7" data-end-pos="14"> Nested</span></div></blockquote></div></blockquote>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestInlineCode(t *testing.T) {
	input := "text [code]inline[/code] text"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// Expect <code>
	if !strings.Contains(got, "<code class=\"a4code-inline a4code-code\"") {
		t.Errorf("expected inline code, got %q", got)
	}
}

func TestBlockCode(t *testing.T) {
	input := "[code]\nblock\n[/code]"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// Expect <pre>
	if !strings.Contains(got, "<pre class=\"a4code-block a4code-code\"") {
		t.Errorf("expected block code, got %q", got)
	}
}

func TestInlineCodeWithBrackets(t *testing.T) {
	input := "please use [code [quote]] so I know."
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// Expect <code> and content "[quote]"
	if !strings.Contains(got, "<code class=\"a4code-inline a4code-code\"") {
		t.Errorf("expected inline code, got %q", got)
	}
	if !strings.Contains(got, "[quote]") {
		t.Errorf("expected content [quote], got %q", got)
	}
}

func TestInlineQuote(t *testing.T) {
	input := "text [quote]inline[/quote] text"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// Expect <q>
	if !strings.Contains(got, "<q class=\"a4code-inline a4code-quote\"") {
		t.Errorf("expected inline quote, got %q", got)
	}
}

func TestBlockQuote(t *testing.T) {
	input := "[quote]\nblock\n[/quote]"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// Expect <blockquote>
	if !strings.Contains(got, "<blockquote class=\"a4code-block a4code-quote") {
		t.Errorf("expected block quote, got %q", got)
	}
}

func TestQuoteOfAlwaysBlock(t *testing.T) {
	input := "text [quoteof user text]"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	// Expect <blockquote>
	if !strings.Contains(got, "<blockquote class=\"a4code-block a4code-quoteof") {
		t.Errorf("expected block quoteof, got %q", got)
	}
}
