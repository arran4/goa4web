package templates

import (
	_ "embed"
	"regexp"
	"strings"
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

func TestThreadPageUsesThreadMarkReadTask(t *testing.T) {
	if strings.Contains(threadPageTemplate, "Mark Topic Read") {
		t.Fatalf("template uses deprecated Mark Topic Read task")
	}
	if c := strings.Count(threadPageTemplate, "Mark Thread Read"); c != 3 {
		t.Fatalf("expected 3 Mark Thread Read tasks, got %d", c)
	}
	if strings.Contains(threadPageTemplate, "/topic/{{.Topic.Idforumtopic}}/labels") {
		t.Fatalf("mark-read form posts to topic labels endpoint")
	}
}
