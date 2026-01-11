package a4code

import (
	"bytes"
)

// ToCode converts the AST back to a4code markup string.
func ToCode(n Node) string {
	var buf bytes.Buffer
	n.a4code(&buf)
	return buf.String()
}
