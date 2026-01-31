package handlertest

import (
	"strings"
)

// GetTemplateFuncs returns a map of template functions required for compiling
// notification templates in tests.
func GetTemplateFuncs() map[string]any {
	return map[string]any{
		"a4code2string": func(s string) string {
			return s
		},
		"truncateWords": func(i int, s string) string {
			words := strings.Fields(s)
			if len(words) > i {
				return strings.Join(words[:i], " ") + "..."
			}
			return s
		},
		"lower": strings.ToLower,
	}
}
