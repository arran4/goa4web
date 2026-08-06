package a4code

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/a4code/ast"
	"github.com/stretchr/testify/assert"
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
	want := `<img src="image.jpg" data-start-pos="0" data-end-pos="1" />`
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
	input := "please use [code [quote\\]] so I know."
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
			want:  `<pre class="a4code-block a4code-code a4code-language-go" data-start-pos="0" data-end-pos="28"><code class="language-go"><span data-start-pos="0" data-end-pos="28">func main() { a := []int{} }</span></code></pre>`,
		},
		{
			name:  "codein with balanced brackets",
			input: `[codein "go" func main() { a := [\]int{} }]`,
			want:  `<pre class="a4code-block a4code-code a4code-language-go" data-start-pos="0" data-end-pos="28"><code class="language-go"><span data-start-pos="0" data-end-pos="28">func main() { a := []int{} }</span></code></pre>`,
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

func TestCodeWithNestedQuote(t *testing.T) {
	input := "[code[quote\\]]"
	tree, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := ToHTML(tree)
	if !strings.Contains(got, "[quote]") {
		t.Errorf("expected content [quote], got %q", got)
	}

	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree.Children))
	}
	codeNode, ok := tree.Children[0].(*ast.Code)
	if !ok {
		t.Fatalf("expected Code node, got %T", tree.Children[0])
	}
	if codeNode.Value != "[quote]" {
		t.Errorf("expected value [quote], got %q", codeNode.Value)
	}
}

func TestQOMarkup(t *testing.T) {
	input := "[qo user text]"
	tree, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree.Children))
	}
	q, ok := tree.Children[0].(*ast.QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", tree.Children[0])
	}
	if q.Name != "user" {
		t.Errorf("expected Name %q, got %q", "user", q.Name)
	}
}

func TestInvalidTags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[invalid]", `<span data-start-pos="0" data-end-pos="0">[invalid]</span>`},
		{"[invalid text]", `<span data-start-pos="0" data-end-pos="4">[invalid<span data-start-pos="0" data-end-pos="4">text</span>]</span>`},
		{"[invalid [b bold]]", `<span data-start-pos="0" data-end-pos="4">[invalid<strong data-start-pos="0" data-end-pos="4"><span data-start-pos="0" data-end-pos="4">bold</span></strong>]</span>`},
		{"[foo=bar]", `<span data-start-pos="0" data-end-pos="3">[foo<span data-start-pos="0" data-end-pos="3">bar</span>]</span>`},
	}

	for _, tc := range tests {
		node, err := ParseString(tc.input)
		if err != nil {
			t.Fatalf("ParseString(%q) error: %v", tc.input, err)
		}
		output := ToHTML(node)
		if output != tc.expected {
			t.Errorf("Input: %q\nExpected: %q\nGot:      %q", tc.input, tc.expected, output)
		}
	}
}

func TestUpdateBlockStatus(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkLink func(*testing.T, *ast.Root)
	}{
		{
			name:  "Root: Standalone link",
			input: "[link url]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected standalone link in root to be block")
				}
			},
		},
		{
			name:  "Root: Link surrounded by newlines",
			input: "\n[link url]\n",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link surrounded by new lines to be block")
				}
			},
		},
		{
			name:  "Root: Link after text no newline",
			input: "foo[link url]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if l.IsBlock {
					t.Error("Expected link after text to be inline")
				}
			},
		},
		{
			name:  "Root: Link before text no newline",
			input: "[link url]foo",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if l.IsBlock {
					t.Error("Expected link before text to be inline")
				}
			},
		},
		{
			name:  "Quote: Standalone link",
			input: "[quote [link url]]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in quote to be block")
				}
			},
		},
		{
			name:  "Quote: Link with newlines",
			input: "[quote \n[link url]\n]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in quote with new lines to be block")
				}
			},
		},
		{
			name:  "Bold: Standalone link (Inline context)",
			input: "[b [link url]]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if l.IsBlock {
					t.Error("Expected link in bold (inline context) to be inline")
				}
			},
		},
		{
			name:  "QuoteOf: Standalone link",
			input: "[quoteof user [link url]]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in quoteof to be block")
				}
			},
		},
		{
			name:  "Spoiler: Standalone link",
			input: "[spoiler [link url]]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in spoiler to be block")
				}
			},
		},
		{
			name:  "Indent: Standalone link",
			input: "[indent [link url]]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in indent to be block")
				}
			},
		},
		{
			name:  "Multiple Block Links",
			input: "[quote [link 1]\n[link 2]]",
			checkLink: func(t *testing.T, root *ast.Root) {
				var links []*ast.Link
				_ = ast.Walk(root, func(n ast.Node) error {
					if l, ok := n.(*ast.Link); ok {
						links = append(links, l)
					}
					return nil
				})
				if len(links) != 2 {
					t.Fatalf("Expected 2 links, got %d", len(links))
				}
				if !links[0].IsBlock {
					t.Error("Expected first link to be block")
				}
				if !links[1].IsBlock {
					t.Error("Expected second link to be block")
				}
			},
		},
		{
			name:  "Mixed Inline/Block Links",
			input: "[quote foo [link 1]\n[link 2]]",
			checkLink: func(t *testing.T, root *ast.Root) {
				var links []*ast.Link
				_ = ast.Walk(root, func(n ast.Node) error {
					if l, ok := n.(*ast.Link); ok {
						links = append(links, l)
					}
					return nil
				})
				if len(links) != 2 {
					t.Fatalf("Expected 2 links, got %d", len(links))
				}
				if links[0].IsBlock {
					t.Error("Expected first link (after text) to be inline")
				}
				if !links[1].IsBlock {
					t.Error("Expected second link (after new line) to be block")
				}
			},
		},
		{
			name:  "Lisp Style: Link with Title",
			input: "[link url Title]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if l.Href != "url" {
					t.Errorf("Expected Href='url', got %q", l.Href)
				}
				// Title should be a child text node
				if len(l.Children) != 1 {
					t.Errorf("Expected 1 child, got %d", len(l.Children))
					return
				}
				if txt, ok := l.Children[0].(*ast.Text); ok {
					if strings.TrimSpace(txt.Value) != "Title" {
						t.Errorf("Expected child text 'Title', got %q", txt.Value)
					}
				} else {
					t.Errorf("Expected child to be Text, got %T", l.Children[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}
			tt.checkLink(t, root)
		})
	}
}

func TestQuoteAdjacentLinkBoundaries(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantLinkBlock  bool
		wantQuoteBlock *bool
	}{
		{
			name:          "link remains block before trailing whitespace and a newline",
			input:         "[link url] \n",
			wantLinkBlock: true,
		},
		{
			name:          "link remains block before whitespace and a following quote",
			input:         "[link url] \n[quote text]",
			wantLinkBlock: true,
		},
		{
			name:           "quote remains block before whitespace and a following link",
			input:          "[quote text] \n[link url]",
			wantLinkBlock:  true,
			wantQuoteBlock: new(true),
		},
		{
			name:           "inline quote and link remain inline",
			input:          "text [quote text] [link url]",
			wantLinkBlock:  false,
			wantQuoteBlock: new(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			link := findFirstLink(root)
			if link == nil {
				t.Fatal("expected link")
			}
			if link.IsBlock != tt.wantLinkBlock {
				t.Errorf("link IsBlock = %t, want %t", link.IsBlock, tt.wantLinkBlock)
			}

			if tt.wantQuoteBlock != nil {
				quote := findFirstQuote(root)
				if quote == nil {
					t.Fatal("expected quote")
				}
				if quote.IsBlock != *tt.wantQuoteBlock {
					t.Errorf("quote IsBlock = %t, want %t", quote.IsBlock, *tt.wantQuoteBlock)
				}
			}
		})
	}
}

func TestCodeBlockEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Nested escaped bracket",
			input:    `[code [quote test\]]`,
			expected: `[quote test]`,
		},
		{
			name:     "Escaped bracket prevents termination (with space)",
			input:    `[code C:\]path ]`,
			expected: `C:]path `,
		},
		{
			name:     "Escaped bracket prevents termination (EOF case)",
			input:    `[code C:\]path]`,
			expected: `C:]path`, // Captures C:]path, last bracket terminates block
		},
		{
			name:     "Standard block content (not balanced anymore)",
			input:    `[code [b]bold[/b]]`,
			expected: `[b`, // Balancing is disabled, stops at first ]
		},
		{
			name:     "Standard block content (fully escaped)",
			input:    `[code [b\]bold[/b\]]`,
			expected: `[b]bold[/b]`, // Now requires escaping all closing brackets
		},
		{
			name:     "Literal bracket at end",
			input:    `[code smile :-\] ]`,
			expected: `smile :-] `,
		},
		{
			name:     "Multiple nested escaped brackets",
			input:    `[code [ [ \] \] ]`,
			expected: `[ [ ] ] `,
		},
		{
			name:     "Escaped open bracket literal",
			input:    `[code \[literal]`,
			expected: `[literal`,
		},
		{
			name:     "Escaped open bracket literal closed",
			input:    `[code \[literal\]]`,
			expected: `[literal]`,
		},
		{
			name:     "New line handling",
			input:    "[code \nline1\nline2\n]",
			expected: "line1\nline2\n", // Leading newline consumed by parser
		},
		{
			name:     "Comment case 1",
			input:    "[code [b]",
			expected: `[b`,
		},
		{
			name:     "Comment case 2",
			input:    "[code [ [ ] ]",
			expected: `[ [ `, // Stops at first ]
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := ParseString(tc.input)
			assert.NoError(t, err)
			assert.NotEmpty(t, root.Children)
			codeNode, ok := root.Children[0].(*ast.Code)
			assert.True(t, ok, "Expected ast.Code node")
			assert.Equal(t, tc.expected, codeNode.Value)
		})
	}
}

func TestLinkIsImmediateClose(t *testing.T) {
	t.Run("NoContent", func(t *testing.T) {
		// a4code syntax [link url] (no content)
		input := "[link http://example.com]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		link := findFirstLink(root)
		assert.NotNil(t, link)
		assert.Equal(t, "http://example.com", link.Href)
		assert.True(t, link.IsImmediateClose(), "Expected IsImmediateClose() to be true for [link url]")
		assert.Empty(t, link.Children)
	})

	t.Run("WithContent", func(t *testing.T) {
		// a4code syntax [link url Content]
		input := "[link http://example.com Content]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		link := findFirstLink(root)
		assert.NotNil(t, link)
		assert.Equal(t, "http://example.com", link.Href)
		assert.False(t, link.IsImmediateClose(), "Expected IsImmediateClose() to be false for [link url Content]")
		assert.NotEmpty(t, link.Children)
	})

	t.Run("WithNestedContent", func(t *testing.T) {
		// a4code syntax [link url [b Bold]]
		input := "[link http://example.com [b Bold]]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		link := findFirstLink(root)
		assert.NotNil(t, link)
		assert.Equal(t, "http://example.com", link.Href)
		assert.False(t, link.IsImmediateClose(), "Expected IsImmediateClose() to be false for [link url [b Bold]]")
		assert.NotEmpty(t, link.Children)
	})
}

func TestToText_Code(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Inline code",
			input: "Start [code func main() {}] End",
			want:  "Start func main() {} End",
		},
		{
			name:  "Block code",
			input: "[code\nfunc main() {}\n]",
			want:  "func main() {}\n",
		},
		{
			name:  "CodeIn",
			input: "[codein \"go\" func main() {}]",
			want:  "func main() {}",
		},
		{
			name:  "Code with brackets",
			input: "[code [b\\]bold[/b\\]]",
			want:  "[b]bold[/b]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString error: %v", err)
			}
			got := ToText(root)
			if got != tt.want {
				t.Errorf("ToText() = %q, want %q", got, tt.want)
			}

			// Also check ToCleanText
			clean := ToCleanText(root)
			if clean != tt.want {
				t.Errorf("ToCleanText() = %q, want %q", clean, tt.want)
			}
		})
	}
}

func findFirstLink(n ast.Node) *ast.Link {
	var found *ast.Link
	_ = ast.Walk(n, func(node ast.Node) error {
		if found != nil {
			return nil
		}
		if l, ok := node.(*ast.Link); ok {
			found = l
		}
		return nil
	})
	return found
}

func findFirstQuote(n ast.Node) *ast.Quote {
	var found *ast.Quote
	_ = ast.Walk(n, func(node ast.Node) error {
		if found != nil {
			return nil
		}
		if quote, ok := node.(*ast.Quote); ok {
			found = quote
		}
		return nil
	})
	return found
}
