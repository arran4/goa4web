package a4code

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/a4code/ast"
)

func TestUpdateBlockStatus(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		checkLink   func(*testing.T, *ast.Root)
	}{
		{
			name:  "Root: Standalone link",
			input: "[link=url]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected standalone link in root to be block")
				}
			},
		},
		{
			name:  "Root: Link surrounded by newlines",
			input: "\n[link=url]\n",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link surrounded by newlines to be block")
				}
			},
		},
		{
			name:  "Root: Link after text no newline",
			input: "foo[link=url]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if l.IsBlock {
					t.Error("Expected link after text to be inline")
				}
			},
		},
		{
			name:  "Root: Link before text no newline",
			input: "[link=url]foo",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if l.IsBlock {
					t.Error("Expected link before text to be inline")
				}
			},
		},
		{
			name:  "Quote: Standalone link",
			input: "[quote][link=url][/quote]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in quote to be block")
				}
			},
		},
		{
			name:  "Quote: Link with newlines",
			input: "[quote]\n[link=url]\n[/quote]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in quote with newlines to be block")
				}
			},
		},
		{
			name:  "Bold: Standalone link (Inline context)",
			input: "[b][link=url][/b]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if l.IsBlock {
					t.Error("Expected link in bold (inline context) to be inline")
				}
			},
		},
		{
			name:  "QuoteOf: Standalone link",
			input: "[quoteof user][link=url][/quoteof]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in quoteof to be block")
				}
			},
		},
		{
			name:  "Spoiler: Standalone link",
			input: "[spoiler][link=url][/spoiler]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in spoiler to be block")
				}
			},
		},
		{
			name:  "Indent: Standalone link",
			input: "[indent][link=url][/indent]",
			checkLink: func(t *testing.T, root *ast.Root) {
				l := findFirstLink(root)
				if !l.IsBlock {
					t.Error("Expected link in indent to be block")
				}
			},
		},
		{
			name:  "Multiple Block Links",
			input: "[quote][link=1]\n[link=2][/quote]",
			checkLink: func(t *testing.T, root *ast.Root) {
				var links []*ast.Link
				ast.Walk(root, func(n ast.Node) error {
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
			input: "[quote]foo [link=1]\n[link=2][/quote]",
			checkLink: func(t *testing.T, root *ast.Root) {
				var links []*ast.Link
				ast.Walk(root, func(n ast.Node) error {
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
					t.Error("Expected second link (after newline) to be block")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := Parse(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			tc.checkLink(t, root)
		})
	}
}

func findFirstLink(n ast.Node) *ast.Link {
	var found *ast.Link
	ast.Walk(n, func(node ast.Node) error {
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
