package templates

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
)

//go:embed site/news/postPage.gohtml
var newsPostPageTemplate string

func TestNewsPostPageLabelFormIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*id="label-form"[^>]*>.*{{ csrfField }}`)
	if !re.MatchString(newsPostPageTemplate) {
		t.Fatalf("label form missing csrfField")
	}
}

func TestNewsPostPageReplyFormIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*>.*?</form>`)
	forms := re.FindAllString(newsPostPageTemplate, -1)
	found := false
	for _, f := range forms {
		if strings.Contains(f, `name="replytext"`) && strings.Contains(f, `{{ csrfField }}`) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("reply form missing csrfField")
	}
}

func TestNewsPostPageDoesNotContainInlineMarkRead(t *testing.T) {
	re := regexp.MustCompile(`class="mark-read"`)
	if re.MatchString(newsPostPageTemplate) {
		t.Fatalf("mark-read actions should be provided by the custom index, not inline forms")
	}
}
