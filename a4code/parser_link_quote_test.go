package a4code

import (
	"testing"

	"github.com/arran4/goa4web/a4code/ast"
)

func TestLinkBlockWithQuote(t *testing.T) {
	input := "[quote test]\n[link test]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	quote := nodes[0].(*ast.Quote)
	if !quote.IsBlock {
		t.Errorf("expected quote to be block")
	}

	link := nodes[2].(*ast.Link)
	if !link.IsBlock {
		t.Errorf("expected link to be block because it follows a quote block")
	}
}

func TestLinkBlockBeforeQuote(t *testing.T) {
	input := "[link test]\n[quote test]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	link := nodes[0].(*ast.Link)
	if !link.IsBlock {
		t.Errorf("expected link to be block because it precedes a quote block")
	}

	quote := nodes[2].(*ast.Quote)
	if !quote.IsBlock {
		t.Errorf("expected quote to be block")
	}
}

func TestInlineQuoteFollowedByLink(t *testing.T) {
	input := "text [quote inline] [link http://example.com]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	quote := nodes[1].(*ast.Quote)
	if quote.IsBlock {
		t.Errorf("expected quote to be inline")
	}

	link := nodes[3].(*ast.Link)
	if link.IsBlock {
		t.Errorf("expected link following inline quote to be inline")
	}
}

func TestLinkBlockWithQuoteOf(t *testing.T) {
	input := "[qo name test][link test]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	quote := nodes[0].(*ast.QuoteOf)
	if !quote.IsBlock {
		t.Errorf("expected quote to be block")
	}

	link := nodes[1].(*ast.Link)
	if !link.IsBlock {
		t.Errorf("expected link to be block because it follows a quote block")
	}
}

func TestLinkBlockBeforeQuoteOf(t *testing.T) {
	input := "[link test][qo name test]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	link := nodes[0].(*ast.Link)
	if !link.IsBlock {
		t.Errorf("expected link to be block because it precedes a quote block")
	}

	quote := nodes[1].(*ast.QuoteOf)
	if !quote.IsBlock {
		t.Errorf("expected quote to be block")
	}
}
