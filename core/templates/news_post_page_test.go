package templates

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"
)

//go:embed site/news/postPage.gohtml
var newsPostPageTemplate string

func TestNewsPostPageMarkReadIncludesCSRF(t *testing.T) {
	re := regexp.MustCompile(`(?s)<form[^>]*class="mark-read"[^>]*>.*?{{ csrfField }}`)
	matches := re.FindAllString(newsPostPageTemplate, -1)
	if len(matches) != 4 {
		t.Fatalf("expected 4 mark-read forms with csrfField, got %d", len(matches))
	}
}

func TestNewsPostPageUsesThreadMarkReadTask(t *testing.T) {
	if c := strings.Count(newsPostPageTemplate, "Mark Thread Read"); c != 4 {
		t.Fatalf("expected 4 Mark Thread Read tasks, got %d", c)
	}
}
