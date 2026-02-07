package a4code

import "github.com/arran4/goa4web/a4code/ast"

// ToA4Code serializes the node tree back into A4code markup.
func ToA4Code(n ast.Node) string {
	return ToCode(n)
}
