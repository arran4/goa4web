package templates

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
)

//go:embed site/writings/articlePage.gohtml
var articlePageTemplate string

func TestArticlePageMarkReadIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*class="mark-read"[^>]*>.*{{ csrfField }}`)
	if !re.MatchString(articlePageTemplate) {
		t.Fatalf("mark-read form missing csrfField")
	}
}

func TestArticlePageUsesThreadMarkReadTask(t *testing.T) {
	if strings.Contains(articlePageTemplate, "Mark Topic Read") {
		t.Fatalf("template uses deprecated Mark Topic Read task")
	}
	if c := strings.Count(articlePageTemplate, "Mark Thread Read"); c != 4 {
		t.Fatalf("expected 4 Mark Thread Read tasks, got %d", c)
	}
}
