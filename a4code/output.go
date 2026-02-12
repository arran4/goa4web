package a4code

import (
	"bytes"

	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/a4code/format"
	"github.com/arran4/goa4web/a4code/text"
)

// ToCode converts the AST back to a4code markup string.
func ToCode(n ast.Node) string {
	var buf bytes.Buffer
	if n != nil {
		if err := ast.Generate(&buf, n, format.NewGenerator()); err != nil {
			return "" // Or log error? Current signature returns string.
		}
	}
	return buf.String()
}

// ToCleanText converts the AST to plain text, stripping all markup and prefixes.
func ToCleanText(n ast.Node) string {
	var buf bytes.Buffer
	if err := ast.Generate(&buf, n, text.NewCleanGenerator()); err != nil {
		return ""
	}
	return buf.String()
}

// ToText converts the AST to plain text, stripping all markup.
func ToText(n ast.Node) string {
	var buf bytes.Buffer
	if err := ast.Generate(&buf, n, text.NewGenerator()); err != nil {
		return ""
	}
	return buf.String()
}
