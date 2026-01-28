package a4code

import (
	"strings"
)

func Snip(s string, l int) string {
	if len(s) > l {
		return strings.TrimSpace(s[:l]) + "..."
	}
	return s
}

func SnipText(s string, l int) string {
	if root, err := ParseString(s); err == nil {
		s = ToText(root)
	}
	return Snip(s, l)
}
