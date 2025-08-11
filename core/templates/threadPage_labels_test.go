package templates

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func csrfField() template.HTML { return "" }

// TestThreadPageShowsDefaultPrivateLabels ensures that the thread page template
// renders special private labels like "new" and "unread".
func TestThreadPageShowsDefaultPrivateLabels(t *testing.T) {
	tmpl := template.New("test").Funcs(template.FuncMap{"csrfField": csrfField})

	// Provide stub templates used by threadPage.gohtml.
	if _, err := tmpl.Parse(`{{define "head"}}{{end}}{{define "tail"}}{{end}}{{define "threadComments"}}{{end}}{{define "forumReply"}}{{end}}`); err != nil {
		t.Fatalf("parse stubs: %v", err)
	}
	if _, err := tmpl.ParseFiles("site/forum/topicLabels.gohtml", "site/forum/threadPage.gohtml"); err != nil {
		t.Fatalf("parse templates: %v", err)
	}

	data := struct {
		Topic    struct{ Idforumtopic int32 }
		Thread   struct{ Idforumthread int32 }
		Labels   []struct{ Text, Type string }
		BasePath string
		BackURL  string
	}{}
	data.Topic.Idforumtopic = 1
	data.Thread.Idforumthread = 3
	data.Labels = []struct{ Text, Type string }{
		{Text: "new", Type: "private"},
		{Text: "unread", Type: "private"},
	}
	data.BasePath = "/forum"
	data.BackURL = "/forum/topic/1/thread/1"

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "threadPage.gohtml", data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `value="new"`) || !strings.Contains(out, `value="unread"`) {
		t.Fatalf("expected new and unread labels in output: %s", out)
	}
}
