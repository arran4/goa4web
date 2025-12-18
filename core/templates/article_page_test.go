package templates

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
)

//go:embed site/writings/articlePage.gohtml
var articlePageTemplate string

func TestArticlePageLabelFormIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*id="label-form"[^>]*>.*{{ csrfField }}`)
	if !re.MatchString(articlePageTemplate) {
		t.Fatalf("label form missing csrfField")
	}
}

func TestArticlePageReplyFormIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*>.*?</form>`)
	forms := re.FindAllString(articlePageTemplate, -1)
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

func TestArticlePageDoesNotContainInlineMarkRead(t *testing.T) {
	re := regexp.MustCompile(`class="mark-read"`)
	if re.MatchString(articlePageTemplate) {
		t.Fatalf("mark-read actions should be provided by the custom index, not inline forms")
	}
}
