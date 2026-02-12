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
	//  Bold (5). vis 0-5.
	// [i (vis 5)
	//  Italic (6). vis 5-11.
	// ] (vis 11)
	// ] (vis 11)
	//  plain (space + plain = 6). vis 11-17.
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
	// [b foo] -> "foo" (3)
	// [i bar] -> "bar" (3)
	want := `<strong data-start-pos="0" data-end-pos="3"><span data-start-pos="0" data-end-pos="3">foo</span></strong><i data-start-pos="3" data-end-pos="6"><span data-start-pos="3" data-end-pos="6">bar</span></i>`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestOffsets(t *testing.T) {
	// [code foo]
	// vis 0-3.
	// Inner content: foo.
	input := "[code foo]"
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
	want := `<blockquote class="a4code-block a4code-quote quote-color-0" data-start-pos="0" data-end-pos="12"><div class="quote-body"><span data-start-pos="0" data-end-pos="6">Outer </span><q class="a4code-inline a4code-quote" data-start-pos="6" data-end-pos="12"><span data-start-pos="6" data-end-pos="12">Nested</span></q></div></blockquote>`
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

func TestInlineCode(t *testing.T) {
	input := "text [code inline] text"
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
	input := "[code\nblock\n]"
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
	input := "text [quote inline] text"
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
	input := "[quote \nblock\n]"
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
		{
			name:  "codein with escaped bracket",
			input: `[codein "go" func main() { a := [\]int{} }]`,
			want:  `<pre class="a4code-block a4code-code a4code-language-go" data-start-pos="0" data-end-pos="29"><code class="language-go"><span data-start-pos="0" data-end-pos="29">func main() { a := []int{} }]</span></code></pre>`,
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

func TestCodeWhitespace(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
	}{
		{
			name:      "code with leading newline",
			input:     "[code \nhi]",
			wantValue: "hi",
		},
		{
			name:      "codein with leading newline",
			input:     "[codein \"go\" \nhi]",
			wantValue: "hi",
		},
		{
			name:      "codein with inline newline",
			input:     "[codein \"go\"\nhi]",
			wantValue: "hi",
		},
		{
			name:      "codein with multiple lines",
			input:     "[codein \"go\" \nhi\nhi]",
			wantValue: "hi\nhi",
		},
		{
			name:      "code with multiple lines",
			input:     "[code \nhi\nhi]",
			wantValue: "hi\nhi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
            if len(root.Children) != 1 {
                t.Fatalf("expected 1 child, got %d", len(root.Children))
            }
            node := root.Children[0]
            var got string
            switch n := node.(type) {
            case *ast.Code:
                got = n.Value
            case *ast.CodeIn:
                got = n.Value
            default:
                t.Fatalf("expected Code or CodeIn, got %T", node)
            }

			if got != tt.wantValue {
				t.Errorf("got %q want %q", got, tt.wantValue)
			}
		})
	}
}

func TestCodeInGenerator(t *testing.T) {
    // Need to verify generator output. ToA4Code uses generator.
	tests := []struct {
		name  string
		input *ast.CodeIn
		want  string
	}{
		{
			name:  "inline codein",
			input: &ast.CodeIn{Language: "go", Value: "hi"},
			want:  "[codein \"go\" hi]",
		},
		{
			name:  "multiline codein",
			input: &ast.CodeIn{Language: "go", Value: "hi\nbye"},
			want:  "[codein \"go\"\nhi\nbye]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
            root := &ast.Root{Children: []ast.Node{tt.input}}
			got := ToA4Code(root)
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}
