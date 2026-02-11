package a4code

import (
	"testing"

	"github.com/arran4/goa4web/a4code/ast"
)

func TestQuote(t *testing.T) {
	got := QuoteText("bob", "hello")
	want := "[quoteof \"bob\" hello]\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	tree, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tree.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := tree.Children[0].(*ast.QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", tree.Children[0])
	}
	if q.Name != "bob" {
		t.Errorf("quote name = %q", q.Name)
	}
	if len(q.Children) < 1 {
		t.Fatalf("expected child text")
	}
	tnode, ok := q.Children[0].(*ast.Text)
	if !ok || tnode.Value != " hello" {
		t.Errorf("quote text = %#v", q.Children[0])
	}
}

func TestQuoteFullParagraphs(t *testing.T) {
	text := "foo\n\nbar"
	got := QuoteText("bob", text, WithParagraphQuote())
	want := "[quoteof \"bob\" foo]\n\n\n\n[quoteof \"bob\" bar]\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	tree, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tree.Children) < 2 {
		t.Fatalf("expected at least 2 nodes, got %d", len(tree.Children))
	}
	q1, _ := tree.Children[0].(*ast.QuoteOf)
	q2, _ := tree.Children[2].(*ast.QuoteOf)
	if q1 == nil || q2 == nil {
		t.Fatalf("parse result types: %T %T", tree.Children[0], tree.Children[2])
	}
	t1 := q1.Children[0].(*ast.Text)
	t2 := q2.Children[0].(*ast.Text)
	if t1.Value != " foo" || t2.Value != " bar" {
		t.Errorf("quote texts = %q %q", t1.Value, t2.Value)
	}
}

func TestQuoteFullEscaping(t *testing.T) {
	text := "see \\[bracket\\]"
	got := QuoteText("bob", text, WithParagraphQuote())
	want := "[quoteof \"bob\" see [bracket]]\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	tree, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tree.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := tree.Children[0].(*ast.QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf, got %T", tree.Children[0])
	}
	if len(q.Children) < 2 {
		t.Fatalf("unexpected children %v", q.Children)
	}
	tnode := q.Children[0].(*ast.Text)
	if tnode.Value != " see " {
		t.Errorf("quote text = %q", tnode.Value)
	}
}

func TestQuoteFullImage(t *testing.T) {
	text := "[img http://example.com/foo.jpg]"
	got := QuoteText("bob", text, WithParagraphQuote())
	want := "[quoteof \"bob\" [img http://example.com/foo.jpg]]\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	tree, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tree.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := tree.Children[0].(*ast.QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf, got %T", tree.Children[0])
	}
	if len(q.Children) < 2 {
		t.Fatalf("unexpected children %v", q.Children)
	}
	if _, ok := q.Children[1].(*ast.Image); !ok {
		t.Fatalf("expected Image node, got %T", q.Children[1])
	}
}

func TestQuoteTrim(t *testing.T) {
	got := QuoteText("bob", " hello \n", WithTrimSpace())
	want := "[quoteof \"bob\" hello]\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestQuoteOfWithSpaces(t *testing.T) {
	input := `[quoteof "Arran on messenger" https://github.com/nao1215/sqly]`
	tree, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tree.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := tree.Children[0].(*ast.QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", tree.Children[0])
	}
	wantName := "Arran on messenger"
	if q.Name != wantName {
		t.Errorf("quote name = %q, want %q", q.Name, wantName)
	}
}

func TestQuoteOfWithSpacesAndEscapedQuote(t *testing.T) {
	input := `[quoteof "Arran \"The Man\" on messenger" https://github.com/nao1215/sqly]`
	tree, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tree.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := tree.Children[0].(*ast.QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", tree.Children[0])
	}
	wantName := `Arran "The Man" on messenger`
	if q.Name != wantName {
		t.Errorf("quote name = %q, want %q", q.Name, wantName)
	}
}

func TestQuoteRoundTripComplexName(t *testing.T) {
	name := `Foo "Bar" Baz \ Quux`
	text := "some content"
	// Generate
	encoded := QuoteText(name, text)

	// Parse back
	tree, err := ParseString(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tree.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := tree.Children[0].(*ast.QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", tree.Children[0])
	}

	if q.Name != name {
		t.Errorf("Round trip name mismatch.\nGot: %q\nWant: %q", q.Name, name)
	}
}

func TestSubstring(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		start int
		end   int
		want  string
	}{
		{
			name:  "Simple",
			s:     "hello world",
			start: 2,
			end:   8,
			want:  "llo wo",
		},
		{
			name:  "With Bold",
			s:     "hello [b world]",
			start: 2,
			end:   8,
			want:  "llo [b wo]",
		},
		{
			name:  "Partial Bold",
			s:     "hello [b world]",
			start: 7,
			end:   10,
			want:  "[b orl]",
		},
		{
			name:  "Across Bold",
			s:     "he[b llo] world",
			start: 1,
			end:   10,
			want:  "e[b llo] wor",
		},
		{
			name:  "Code block",
			s:     "[code]hello[/code]",
			start: 0,
			end:   5,
			want:  "[code]hello[/code]",
		},
		{
			name:  "Partial code block",
			s:     "[code]hello[/code]",
			start: 1,
			end:   3,
			want:  "[code]el[/code]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Substring(tt.s, tt.start, tt.end); got != tt.want {
				t.Errorf("Substring() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuoteDepthOptions(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		restrictedQuoteDepth *int
		truncatedQuoteDepth  *int
		want                 string
	}{
		{
			name:                 "RestrictedDepth0_FilterNested",
			input:                "[quoteof \"A\" [quoteof \"B\" content]]",
			restrictedQuoteDepth: intPtr(0),
			want:                 "",
		},
		{
			name:                 "RestrictedDepth0_AllowDepth0",
			input:                "[quoteof \"A\" content]",
			restrictedQuoteDepth: intPtr(0),
			want:                 "[quoteof \"user\" [quoteof \"A\" content]]\n",
		},
		{
			name:                 "RestrictedDepth1_AllowDepth1",
			input:                "[quoteof \"A\" [quoteof \"B\" content]]",
			restrictedQuoteDepth: intPtr(1),
			want:                 "[quoteof \"user\" [quoteof \"A\" [quoteof \"B\" content]]]\n",
		},
		{
			name:                 "RestrictedDepth1_FilterDepth2",
			input:                "[quoteof \"A\" [quoteof \"B\" [quoteof \"C\" content]]]",
			restrictedQuoteDepth: intPtr(1),
			want:                 "",
		},
		{
			name:                 "RestrictedDepth0_PureQuoteCheck_TextAllowed",
			input:                "[quoteof \"A\" content] and more text",
			restrictedQuoteDepth: intPtr(0),
			want:                 "[quoteof \"user\" [quoteof \"A\" content] and more text]\n",
		},
		{
			name:                "TruncatedDepth0_TruncateInner",
			input:               "[quoteof \"A\" content]",
			truncatedQuoteDepth: intPtr(0),
			want:                "[quoteof \"user\" [quoteof \"A\"]]\n",
		},
		{
			name:                 "TruncatedDepth1_TruncateDepth2",
			input:                "[quoteof \"A\" [quoteof \"B\" content]]",
			truncatedQuoteDepth:  intPtr(1),
			restrictedQuoteDepth: intPtr(2),
			want:                 "[quoteof \"user\" [quoteof \"A\" [quoteof \"B\"]]]\n",
		},
		{
			name:                "TruncatedDepth1_AllowDepth1",
			input:               "[quoteof \"A\" content]",
			truncatedQuoteDepth: intPtr(1),
			want:                "[quoteof \"user\" [quoteof \"A\" content]]\n",
		},
		{
			name:                 "RestrictedAndTruncated_Mix",
			input:                "[quoteof \"A\" [quoteof \"B\" [quoteof \"C\" content]]]",
			restrictedQuoteDepth: intPtr(2),
			truncatedQuoteDepth:  intPtr(1),
			want:                 "[quoteof \"user\" [quoteof \"A\" [quoteof \"B\"]]]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []QuoteOption
			opts = append(opts, WithParagraphQuote())
			if tt.restrictedQuoteDepth != nil {
				opts = append(opts, WithRestrictedQuoteDepth(*tt.restrictedQuoteDepth))
			}
			if tt.truncatedQuoteDepth != nil {
				opts = append(opts, WithTruncatedQuoteDepth(*tt.truncatedQuoteDepth))
			}

			got := QuoteText("user", tt.input, opts...)
			if got != tt.want {
				t.Errorf("QuoteText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestQuoteRepro(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "AllowSingleLevelQuote",
			input: "Para 1\n\n[quoteof \"other\" inner]\n\nPara 2",
			want:  "[quoteof \"user\" Para 1]\n\n\n\n[quoteof \"user\" [quoteof \"other\" inner]]\n\n\n\n[quoteof \"user\" Para 2]\n",
		},
		{
			name:  "FilterDeeplyNestedQuote",
			input: "Para 1\n\n[quoteof \"other\" [quoteof \"inner\" content]]\n\nPara 2",
			want:  "[quoteof \"user\" Para 1]\n\n\n\n[quoteof \"user\" Para 2]\n",
		},
		{
			name:  "TripleLineBreaks",
			input: "Para 1\n\nPara 2",
			want:  "[quoteof \"user\" Para 1]\n\n\n\n[quoteof \"user\" Para 2]\n",
		},
		{
			name:  "ParagraphStartingWithBracket",
			input: "Para 1\n\n[b bold]",
			want:  "[quoteof \"user\" Para 1]\n\n\n\n[quoteof \"user\" [b bold]]\n",
		},
		{
			name:  "NotFilterMixedQuotes",
			input: "[quoteof \"other\" inner] and more",
			want:  "[quoteof \"user\" [quoteof \"other\" inner] and more]\n",
		},
		{
			name:  "EmptyInput",
			input: "",
			want:  "",
		},
		{
			name:  "WhitespaceOnly",
			input: "   \t   ",
			want:  "",
		},
		{
			name:  "NewlinesOnly",
			input: "\n\n\n",
			want:  "",
		},
		{
			name:  "MixedContentPreQuote",
			input: "Prefix [quoteof \"other\" inner]",
			want:  "[quoteof \"user\" Prefix [quoteof \"other\" inner]]\n",
		},
		{
			name:  "MultipleNestedQuotesSameParagraph",
			input: "[quoteof \"A\" 1] [quoteof \"B\" 2]",
			want:  "[quoteof \"user\" [quoteof \"A\" 1] [quoteof \"B\" 2]]\n",
		},
		{
			name:  "CaseInsensitiveFilter",
			input: "[QUOTEOF \"other\" inner]",
			want:  "[quoteof \"user\" [QUOTEOF \"other\" inner]]\n",
		},
		{
			name:  "ParagraphStartingWithCloseBracket",
			input: "]\n\nNext",
			want:  "[quoteof \"user\" ]]\n\n\n\n[quoteof \"user\" Next]\n",
		},
		{
			name:  "ParagraphStartingWithBackslash",
			input: "\\[\n\nNext",
			want:  "[quoteof \"user\" []\n\n\n\n[quoteof \"user\" Next]\n",
		},
		{
			name:  "UserExampleNestedQuote",
			input: "Para 1\n\n[quoteof \"M\" [quoteof \"a\" asdf]]\n\nPara 2",
			want:  "[quoteof \"user\" Para 1]\n\n\n\n[quoteof \"user\" Para 2]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuoteText("user", tt.input, WithParagraphQuote())
			if got != tt.want {
				t.Errorf("QuoteText() = %q, want %q", got, tt.want)
			}
		})
	}
}
