package common

import (
	"html"
	"html/template"
	"strings"

	searchdomain "github.com/arran4/goa4web/internal/search"
)

// HighlightSearchTerms returns HTML with searchWords highlighted using <mark>.
// The returned HTML escapes all other content to avoid injection.
func HighlightSearchTerms(text string, searchWords []string) template.HTML {
	if text == "" {
		return template.HTML("")
	}
	normalized := map[string]struct{}{}
	for _, word := range searchWords {
		if trimmed := strings.TrimSpace(word); trimmed != "" {
			normalized[strings.ToLower(trimmed)] = struct{}{}
		}
	}
	if len(normalized) == 0 {
		return template.HTML(html.EscapeString(text))
	}

	var b strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); {
		if !searchdomain.IsTokenRune(runes[i]) {
			j := i + 1
			for j < len(runes) && !searchdomain.IsTokenRune(runes[j]) {
				j++
			}
			b.WriteString(html.EscapeString(string(runes[i:j])))
			i = j
			continue
		}
		j := i + 1
		for j < len(runes) && searchdomain.IsTokenRune(runes[j]) {
			j++
		}
		segment := string(runes[i:j])
		if _, ok := normalized[strings.ToLower(segment)]; ok {
			b.WriteString("<mark>")
			b.WriteString(html.EscapeString(segment))
			b.WriteString("</mark>")
		} else {
			b.WriteString(html.EscapeString(segment))
		}
		i = j
	}
	return template.HTML(b.String())
}
