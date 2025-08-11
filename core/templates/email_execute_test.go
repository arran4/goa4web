package templates_test

import (
	htemplate "html/template"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/templates"
)

type emailData struct {
	URL            string
	SubjectPrefix  string
	UnsubscribeUrl string
	SignOff        string
	SignOffHTML    htemplate.HTML
	Item           interface{}
}

func sampleEmailData() emailData {
	item := map[string]interface{}{
		"Action":       "action",
		"Author":       "author",
		"BlogURL":      "https://example.com/blog",
		"BoardURL":     "https://example.com/board",
		"Code":         "code",
		"CommentURL":   "https://example.com/comment",
		"IP":           "127.0.0.1",
		"LanguageID":   "en",
		"LanguageName": "English",
		"LinkURL":      "https://example.com/link",
		"Message":      "message",
		"Moderator":    "mod",
		"NewsURL":      "https://example.com/news",
		"Path":         "/path",
		"Permission":   "perm",
		"PostURL":      "https://example.com/post",
		"Question":     "question",
		"Reason":       "reason",
		"ResetURL":     "https://example.com/reset",
		"Role":         "role",
		"Thread": map[string]interface{}{
			"Lastposterusername": map[string]interface{}{"String": "poster"},
			"Comments":           map[string]interface{}{"Int32": 1},
		},
		"ThreadID":           1,
		"ThreadURL":          "https://example.com/thread",
		"Time":               time.Now(),
		"Title":              map[string]interface{}{"String": "title"},
		"TopicTitle":         "topic title",
		"UnsubURL":           "https://example.com/unsub",
		"UserPermissionsURL": "https://example.com/perm",
		"UserURL":            "https://example.com/user",
		"Username":           "username",
	}
	return emailData{
		URL:            "https://example.com",
		SubjectPrefix:  "prefix",
		UnsubscribeUrl: "https://example.com/unsubscribe",
		SignOff:        "signoff",
		SignOffHTML:    htemplate.HTML("signoff"),
		Item:           item,
	}
}

func TestEmailTemplatesExecute(t *testing.T) {
	htmlT := templates.GetCompiledEmailHtmlTemplates(nil)
	textT := templates.GetCompiledEmailTextTemplates(nil)
	data := sampleEmailData()

	for _, tmpl := range htmlT.Templates() {
		name := tmpl.Name()
		if !strings.HasSuffix(name, ".gohtml") {
			continue
		}
		t.Run("html/"+name, func(t *testing.T) {
			if err := tmpl.Execute(io.Discard, data); err != nil {
				t.Errorf("execute %s: %v", name, err)
			}
		})
	}

	for _, tmpl := range textT.Templates() {
		name := tmpl.Name()
		if !strings.HasSuffix(name, ".gotxt") {
			continue
		}
		t.Run("text/"+name, func(t *testing.T) {
			if err := tmpl.Execute(io.Discard, data); err != nil {
				t.Errorf("execute %s: %v", name, err)
			}
		})
	}
}
