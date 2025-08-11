package templates

import (
	_ "embed"
	"regexp"
	"testing"
)

//go:embed site/forum/threadPage.gohtml
var threadPageTemplate string

func TestThreadPageMarkReadIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*class="mark-read"[^>]*>.*{{ csrfField }}`)
	if !re.MatchString(threadPageTemplate) {
		t.Fatalf("mark-read form missing csrfField")
	}
}
