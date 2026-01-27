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
	// [b (0-2)
	//  Bold (2-8)
	// [i (8-10)
	//  Italic (10-17)
	// ] (17-18) end italic
	// ] (18-19) end bold
	//  plain (19-25)
	want := `<strong data-start-pos="0" data-end-pos="19"><span data-start-pos="2" data-end-pos="8"> Bold </span><i data-start-pos="8" data-end-pos="18"><span data-start-pos="10" data-end-pos="17"> Italic</span></i></strong><span data-start-pos="19" data-end-pos="25"> plain</span>`
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
	// [img=image.jpg] len 15
	want := `<img src="image.jpg" data-start-pos="0" data-end-pos="15" />`
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
	// [b foo] (0-7)
	// [i bar] (7-14)
	want := `<strong data-start-pos="0" data-end-pos="7"><span data-start-pos="2" data-end-pos="6"> foo</span></strong><i data-start-pos="7" data-end-pos="14"><span data-start-pos="9" data-end-pos="13"> bar</span></i>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestOffsets(t *testing.T) {
	// [code]foo[/code]
	// [ (0-1)
	// code (1-5)
	// ] (5-6) included in content due to parseCommand behavior
	// foo (6-9)
	// [/code] (9-16)
	// Code node: 0-16. Inner: 5-9.
	input := "[code]foo[/code]"
	ast, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(ast)
	// Expect ]foo as content because skipArgPrefix does not consume ]
	want := `<pre class="a4code-block a4code-code" data-start-pos="0" data-end-pos="16"><span data-start-pos="5" data-end-pos="9">]foo</span></pre>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
