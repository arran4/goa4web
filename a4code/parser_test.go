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
	want := `<blockquote class="a4code-block a4code-quote quote-color-0" data-start-pos="0" data-end-pos="14"><div class="quote-body"><span data-start-pos="0" data-end-pos="7"> Outer </span><blockquote class="a4code-block a4code-quote quote-color-1" data-start-pos="7" data-end-pos="14"><div class="quote-body"><span data-start-pos="7" data-end-pos="14"> Nested</span></div></blockquote></div></blockquote>`
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

func TestCodeIn(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple codein",
			input: `[codein "go" func main() {}]`,
			want:  `<pre class="a4code-block a4code-code a4code-language-go" data-start-pos="0" data-end-pos="14"><code class="language-go"><span data-start-pos="0" data-end-pos="14">func main() {}</span></code></pre>`,
		},
		{
			name:  "codein unquoted language",
			input: `[codein go func main() {}]`,
			want:  `<pre class="a4code-block a4code-code a4code-language-go" data-start-pos="0" data-end-pos="14"><code class="language-go"><span data-start-pos="0" data-end-pos="14">func main() {}</span></code></pre>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := ToHTML(tree)
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}
