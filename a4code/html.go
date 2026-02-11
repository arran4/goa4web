package a4code

import "bytes"

// ToHTML converts a node tree to HTML.
func ToHTML(n Node) string {
	var buf bytes.Buffer
	n.html(&buf, 0)
	return buf.String()
}
