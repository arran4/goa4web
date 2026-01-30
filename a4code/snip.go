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

func SnipWords(s string, count int) string {
	words := strings.Fields(s)
	if len(words) > count {
		return strings.Join(words[:count], " ") + "..."
	}
	return strings.Join(words, " ")
}

func SnipTextWords(s string, count int) string {
	if root, err := ParseString(s); err == nil {
		s = ToText(root)
	}
	return SnipWords(s, count)
}
