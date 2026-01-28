package a4code

import (
	"testing"
)

func TestQuote(t *testing.T) {
	got := QuoteText("bob", "hello")
	want := "[quoteof \"bob\" hello]\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	ast, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ast.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := ast.Children[0].(*QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", ast.Children[0])
	}
	if q.Name != "bob" {
		t.Errorf("quote name = %q", q.Name)
	}
	if len(q.Children) < 1 {
		t.Fatalf("expected child text")
	}
	tnode, ok := q.Children[0].(*Text)
	if !ok || tnode.Value != " hello" {
		t.Errorf("quote text = %#v", q.Children[0])
	}
}

func TestQuoteFullParagraphs(t *testing.T) {
	text := "foo\n\nbar"
	got := QuoteText("bob", text, WithParagraphQuote())
	want := "[quoteof \"bob\" foo]\n\n\n[quoteof \"bob\" bar]\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	ast, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ast.Children) < 2 {
		t.Fatalf("expected at least 2 nodes, got %d", len(ast.Children))
	}
	q1, _ := ast.Children[0].(*QuoteOf)
	q2, _ := ast.Children[2].(*QuoteOf)
	if q1 == nil || q2 == nil {
		t.Fatalf("parse result types: %T %T", ast.Children[0], ast.Children[2])
	}
	t1 := q1.Children[0].(*Text)
	t2 := q2.Children[0].(*Text)
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
	ast, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ast.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := ast.Children[0].(*QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf, got %T", ast.Children[0])
	}
	if len(q.Children) < 2 {
		t.Fatalf("unexpected children %v", q.Children)
	}
	tnode := q.Children[0].(*Text)
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
	ast, err := ParseString(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ast.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := ast.Children[0].(*QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf, got %T", ast.Children[0])
	}
	if len(q.Children) < 2 {
		t.Fatalf("unexpected children %v", q.Children)
	}
	if _, ok := q.Children[1].(*Image); !ok {
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
	ast, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ast.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := ast.Children[0].(*QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", ast.Children[0])
	}
	wantName := "Arran on messenger"
	if q.Name != wantName {
		t.Errorf("quote name = %q, want %q", q.Name, wantName)
	}
}

func TestQuoteOfWithSpacesAndEscapedQuote(t *testing.T) {
	input := `[quoteof "Arran \"The Man\" on messenger" https://github.com/nao1215/sqly]`
	ast, err := ParseString(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ast.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := ast.Children[0].(*QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", ast.Children[0])
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
	ast, err := ParseString(encoded)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ast.Children) < 1 {
		t.Fatalf("no nodes parsed")
	}
	q, ok := ast.Children[0].(*QuoteOf)
	if !ok {
		t.Fatalf("expected QuoteOf node, got %T", ast.Children[0])
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
			s:     "hello [b]world[/b]",
			start: 2,
			end:   8,
			want:  "llo [b]wo[/b]",
		},
		{
			name:  "Partial Bold",
			s:     "hello [b]world[/b]",
			start: 7,
			end:   10,
			want:  "[b]orl[/b]",
		},
		{
			name:  "Across Bold",
			s:     "he[b]llo[/b] world",
			start: 1,
			end:   10,
			want:  "e[b]llo[/b] wor",
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

func TestQuoteRepro(t *testing.T) {
	// Case 1: Filter out quotes of quotes
	t.Run("FilterNestedQuotes", func(t *testing.T) {
		input := "Para 1\n\n[quoteof \"other\" inner]\n\nPara 2"
		got := QuoteText("user", input, WithParagraphQuote())

		// Expected:
		// [quoteof "user" Para 1]
		// \n\n\n (triple line break added after Para 1)
		// (Inner quote filtered out)
		// [quoteof "user" Para 2]
		// \n

		want := "[quoteof \"user\" Para 1]\n\n\n[quoteof \"user\" Para 2]\n"
		if got != want {
			t.Errorf("Got:\n%q\nWant:\n%q", got, want)
		}
	})

	// Case 2: Triple line breaks
	t.Run("TripleLineBreaks", func(t *testing.T) {
		input := "Para 1\n\nPara 2"
		got := QuoteText("user", input, WithParagraphQuote())
		want := "[quoteof \"user\" Para 1]\n\n\n[quoteof \"user\" Para 2]\n"
		if got != want {
			t.Errorf("Got:\n%q\nWant:\n%q", got, want)
		}
	})

	// Case 3: Bug fix for starting bracket
	t.Run("ParagraphStartingWithBracket", func(t *testing.T) {
		input := "Para 1\n\n[b]bold[/b]"
		got := QuoteText("user", input, WithParagraphQuote())
		// Before fix: ...[quoteof "user" [\nbold...]
		// After fix: ...[quoteof "user" [b]bold[/b]]

		want := "[quoteof \"user\" Para 1]\n\n\n[quoteof \"user\" [b]bold[/b]]\n"
		if got != want {
			t.Errorf("Got:\n%q\nWant:\n%q", got, want)
		}
	})

	// Case 4: Not filtering if mixed content
	t.Run("NotFilterMixedQuotes", func(t *testing.T) {
		input := "[quoteof \"other\" inner] and more"
		got := QuoteText("user", input, WithParagraphQuote())
		want := "[quoteof \"user\" [quoteof \"other\" inner] and more]\n"
		if got != want {
			t.Errorf("Got:\n%q\nWant:\n%q", got, want)
		}
	})
}
