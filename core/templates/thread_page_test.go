package templates

import (
	_ "embed"
	"regexp"
	"testing"
)

//go:embed site/forum/threadPage.gohtml
var threadPageTemplate string

func TestThreadPageLabelFormIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*id="label-form"[^>]*>.*{{ csrfField }}`)
	if !re.MatchString(threadPageTemplate) {
		t.Fatalf("label form missing csrfField")
	}
}

func TestThreadPageDoesNotContainInlineMarkRead(t *testing.T) {
	re := regexp.MustCompile(`class="mark-read"`)
	if re.MatchString(threadPageTemplate) {
		t.Fatalf("mark-read actions should be provided by the custom index, not inline forms")
	}
}
