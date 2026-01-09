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
