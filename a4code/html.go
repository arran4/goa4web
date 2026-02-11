package a4code

import (
	"bytes"

	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/a4code/html"
)

// ToHTML converts a node tree to HTML.
func ToHTML(n ast.Node) string {
	var buf bytes.Buffer
	if err := ast.Generate(&buf, n, html.NewGenerator()); err != nil {
		return ""
	}
	return buf.String()
}
