package templates

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// dict is a helper for building maps in templates.
func dict(values ...any) map[string]any {
	m := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, _ := values[i].(string)
		m[key] = values[i+1]
	}
	return m
}

func csrfField() template.HTML { return "" }

// TestThreadPageShowsDefaultPrivateLabels ensures that the thread page template
// renders special private labels like "new" and "unread".
func TestThreadPageShowsDefaultPrivateLabels(t *testing.T) {
	funcMap := template.FuncMap{
		"dict":      dict,
		"csrfField": csrfField,
	}
	tmpl := template.New("test").Funcs(funcMap)

	// Provide stub templates used by threadPage.gohtml.
	if _, err := tmpl.Parse(`{{define "head"}}{{end}}{{define "tail"}}{{end}}{{define "threadComments"}}{{end}}{{define "forumReply"}}{{end}}`); err != nil {
		t.Fatalf("parse stubs: %v", err)
	}
	if _, err := tmpl.ParseFiles("site/forum/topicLabels.gohtml", "site/forum/threadPage.gohtml"); err != nil {
		t.Fatalf("parse templates: %v", err)
	}

	data := struct {
		Topic         struct{ Idforumtopic int32 }
		Thread        struct{ Idforumthread int32 }
		PublicLabels  []string
		AuthorLabels  []string
		PrivateLabels []string
		BasePath      string
	}{}
	data.Topic.Idforumtopic = 1
	data.Thread.Idforumthread = 3
	data.PrivateLabels = []string{"new", "unread"}
	data.BasePath = "/forum"

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "threadPage.gohtml", data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `value="new"`) || !strings.Contains(out, `value="unread"`) {
		t.Fatalf("expected new and unread labels in output: %s", out)
	}
}
