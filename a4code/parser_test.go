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
	ast, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(ast)
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
	root := &Root{Children: nodes}
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
	ast, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(ast)
	want := `<pre class="a4code-block a4code-code" data-start-pos="0" data-end-pos="3"><span data-start-pos="0" data-end-pos="3">foo</span></pre>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
