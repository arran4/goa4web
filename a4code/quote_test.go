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
	if q.Name != "\"bob\"" {
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
	got := QuoteText("bob", text, WithFullQuote())
	want := "[quoteof \"bob\" foo]\n[quoteof \"bob\" bar]\n"
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
	got := QuoteText("bob", text, WithFullQuote())
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
	got := QuoteText("bob", text, WithFullQuote())
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
