package news

import (
	"bytes"
	"database/sql"
	"html/template"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

type fakeCD struct{}

func (*fakeCD) SelectedSectionThreadComments() []*db.GetCommentsByThreadIdForUserRow { return nil }
func (*fakeCD) Offset() int                                                          { return 0 }
func (*fakeCD) SelectedThreadCanReply() bool                                         { return false }
func (*fakeCD) SelectedThreadID() int32                                              { return 0 }
func (*fakeCD) ShowEditNews(int32, int32) bool                                       { return false }
func (*fakeCD) IsAdmin() bool                                                        { return false }
func (*fakeCD) IsAdminMode() bool                                                    { return false }
func (*fakeCD) NewsAnnouncement(int32) *db.SiteAnnouncement                          { return nil }
func (*fakeCD) Location() *time.Location                                             { return time.UTC }

func TestNewsPostPageNoDuplicateLabels(t *testing.T) {
	funcMap := template.FuncMap{
		"cd":          func() *fakeCD { return &fakeCD{} },
		"csrfField":   func() template.HTML { return "" },
		"localTimeIn": func(t time.Time, _ string) time.Time { return t },
		"localTime":   func(t time.Time) time.Time { return t },
		"now":         func() time.Time { return time.Unix(0, 0) },
		"a4code2html": func(s string) template.HTML { return template.HTML(s) },
		"NewsLabels": func(int32) []templates.TopicLabel {
			return []templates.TopicLabel{{Name: "foo", Type: "author"}}
		},
		"add":   func(a, b int) int { return a + b },
		"since": func(time.Time, time.Time) string { return "" },
	}

	base := filepath.Join("..", "..", "core", "templates", "site", "news")
	tmpl := template.Must(template.New("root").Funcs(funcMap).ParseFiles(
		filepath.Join(base, "postPage.gohtml"),
		filepath.Join(base, "post.gohtml"),
	))
	tmpl = template.Must(tmpl.Parse(`{{ define "head" }}{{ end }}
{{ define "tail" }}{{ end }}
{{ define "threadComments" }}{{ end }}
{{ define "comment" }}{{ end }}
{{ define "topicLabels" }}{{ end }}
{{ define "languageCombobox" }}{{ end }}`))

	post := &db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow{
		Idsitenews:   1,
		UsersIdusers: 1,
		News:         sql.NullString{String: "body", Valid: true},
		Writername:   sql.NullString{String: "alice", Valid: true},
		Occurred:     sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Timezone:     sql.NullString{String: "UTC", Valid: true},
		Comments:     sql.NullInt32{Int32: 0, Valid: true},
	}

	data := struct {
		Post    *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
		Labels  []templates.TopicLabel
		BackURL string
	}{
		Post:    post,
		Labels:  []templates.TopicLabel{{Name: "foo", Type: "author"}},
		BackURL: "/news/news/1",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "postPage.gohtml", data); err != nil {
		t.Fatalf("render: %v", err)
	}

	out := buf.String()
	if strings.Count(out, "class=\"label-bar\"") != 1 {
		t.Fatalf("expected 1 label bar, got %d: %q", strings.Count(out, "class=\"label-bar\""), out)
	}
}
