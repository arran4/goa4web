package templates_test

import (
	htemplate "html/template"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

type emailData struct {
	URL            string
	SubjectPrefix  string
	UnsubscribeUrl string
	SignOff        string
	SignOffHTML    htemplate.HTML
	Item           any
	Recipient      any
	Notifications  []any
	BaseURL        string
}

func sampleEmailData() emailData {
	item := map[string]any{
		"Action":       "action",
		"Author":       "author",
		"Body":         "body",
		"ExpiresAt":    time.Now().Add(24 * time.Hour),
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
		"Thread": map[string]any{
			"Lastposterusername": map[string]any{"String": "poster"},
			"Comments":           map[string]any{"Int32": 1},
		},
		"ThreadID":           1,
		"ThreadURL":          "https://example.com/thread",
		"Time":               time.Now(),
		"Title":              map[string]any{"String": "title"},
		"TopicTitle":         "topic title",
		"UnsubURL":           "https://example.com/unsub",
		"UserPermissionsURL": "https://example.com/perm",
		"UserURL":            "https://example.com/user",
		"Username":           "username",
		"BaseURL":            "https://example.com",
		"Notifications": []any{
			map[string]any{
				"Link":      map[string]any{"Valid": true, "String": "/link"},
				"Message":   map[string]any{"String": "message"},
				"CreatedAt": time.Now(),
			},
		},
		"LastSentAt": map[string]any{"Valid": true, "Time": time.Now()},

		"DigestTitle": "Daily Digest",
	}
	return emailData{
		URL:            "https://example.com",
		BaseURL:        "https://example.com",
		SubjectPrefix:  "prefix",
		UnsubscribeUrl: "https://example.com/unsubscribe",
		SignOff:        "signoff",
		SignOffHTML:    htemplate.HTML("signoff"),
		Item:           item,
		Recipient: map[string]any{
			"Username": map[string]any{"String": "recipient"},
		},
		Notifications: []any{
			map[string]any{
				"Link":      map[string]any{"Valid": true, "String": "/link"},
				"Message":   map[string]any{"String": "message"},
				"CreatedAt": time.Now(),
			},
		},
	}
}

func TestEmailTemplatesExecute(t *testing.T) {
	funcs := common.GetTemplateFuncs()
	htmlT := templates.GetCompiledEmailHtmlTemplates(funcs, templates.WithSilence(true))
	textT := templates.GetCompiledEmailTextTemplates(funcs, templates.WithSilence(true))
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
