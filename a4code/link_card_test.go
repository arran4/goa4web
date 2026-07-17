package a4code

import (
	"testing"
	"github.com/arran4/goa4web/a4code/ast"
)

func TestLinkCardAfterQuote(t *testing.T) {
	input := "[quote foo]\n[link bar]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse nodes error: %v", err)
	}
	linkNode, ok := nodes[len(nodes)-1].(*ast.Link)
	if !ok {
		t.Fatalf("expected link node")
	}
	if !linkNode.IsBlock {
		t.Errorf("expected link after quote\\n to be block")
	}

	input2 := "[quote foo][link bar]"
	nodes2, err := ParseNodes(input2)
	if err != nil {
		t.Fatalf("parse nodes error: %v", err)
	}
	linkNode2, ok := nodes2[1].(*ast.Link)
	if !ok {
		t.Fatalf("expected link node")
	}
	if !linkNode2.IsBlock {
		t.Errorf("expected link right after quote ending to be block")
	}
}

func TestLinkCardBeforeQuote(t *testing.T) {
	input := "[link bar][quote foo]"
	nodes, err := ParseNodes(input)
	if err != nil {
		t.Fatalf("parse nodes error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	linkNode, ok := nodes[0].(*ast.Link)
	if !ok {
		t.Fatalf("expected link node, got %T", nodes[0])
	}
	if !linkNode.IsBlock {
		t.Errorf("expected link before quote to be block")
	}
}
