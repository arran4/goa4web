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

// ToText converts the AST to plain text, stripping all markup.
func ToText(n Node) string {
	var buf bytes.Buffer
	Walk(n, func(node Node) error {
		if t, ok := node.(*Text); ok {
			buf.WriteString(t.Value)
		}
		return nil
	})
	return buf.String()
}
