package a4code

import "bytes"

// ToA4Code serializes the node tree back into A4code markup.
func ToA4Code(n Node) string {
	var buf bytes.Buffer
	n.a4code(&buf)
	return buf.String()
}
