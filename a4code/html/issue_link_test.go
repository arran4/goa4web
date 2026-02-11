package html

import (
	"bytes"
	"testing"

	"github.com/arran4/goa4web/a4code/ast"
	"github.com/google/go-cmp/cmp"
)

func TestLinkWithWhitespaceChildren(t *testing.T) {
	tests := []struct {
		name string
		node *ast.Link
		want string
	}{
		{
			name: "Empty children",
			node: &ast.Link{Href: "http://example.com/"},
			want: `<a href="http://example.com/" target="_BLANK" data-start-pos="0" data-end-pos="0">http://example.com/</a>`,
		},
		{
			name: "Whitespace children",
			node: &ast.Link{
				Href: "http://example.com/",
				Children: []ast.Node{
					&ast.Text{Value: " "},
				},
			},
			want: `<a href="http://example.com/" target="_BLANK" data-start-pos="0" data-end-pos="0">http://example.com/</a>`,
		},
		{
			name: "Non-whitespace children",
			node: &ast.Link{
				Href: "http://example.com/",
				Children: []ast.Node{
					&ast.Text{Value: " Link "},
				},
			},
			want: `<a href="http://example.com/" target="_BLANK" data-start-pos="0" data-end-pos="0"><span data-start-pos="0" data-end-pos="0"> Link </span></a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			gen := NewGenerator()
			if err := gen.Link(&buf, tt.node); err != nil {
				t.Fatalf("Generate error: %v", err)
			}
			got := buf.String()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("%s: diff\n%s", tt.name, diff)
			}
		})
	}
}
