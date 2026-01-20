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
	funcMap := template.FuncMap{
		"csrfField": csrfField,
		"assetHash": func(s string) string { return s },
		"dict": func(values ...any) map[string]any {
			m := make(map[string]any)
			for i := 0; i+1 < len(values); i += 2 {
				k, _ := values[i].(string)
				m[k] = values[i+1]
			}
			return m
		},
	}
	tmpl := template.New("test").Funcs(funcMap)

	// Provide stub templates used by threadPage.gohtml.
	if _, err := tmpl.Parse(`{{define "head"}}{{end}}{{define "tail"}}{{end}}{{define "threadComments"}}{{end}}{{define "forumReply"}}{{end}}{{define "_share.gohtml"}}{{end}}`); err != nil {
		t.Fatalf("parse stubs: %v", err)
	}
	if _, err := tmpl.ParseFiles("site/forum/topicLabels.gohtml", "site/forum/threadPage.gohtml"); err != nil {
		t.Fatalf("parse templates: %v", err)
	}

	data := struct {
		Topic    struct{ Idforumtopic int32 }
		Thread   struct{ Idforumthread int32 }
		Labels   []TopicLabel
		BasePath string
		BackURL  string
	}{}
	data.Topic.Idforumtopic = 1
	data.Thread.Idforumthread = 3
	data.Labels = []TopicLabel{{Name: "new", Type: "private"}, {Name: "unread", Type: "private"}}
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
