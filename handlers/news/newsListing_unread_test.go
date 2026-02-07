package news

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

type mockCD struct {
	labels []templates.TopicLabel
}

func (m *mockCD) LatestNews() []*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow {
	return []*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow{
		{
			Idsitenews:   123,
			UsersIdusers: 456,
			Writername:   sql.NullString{String: "writer", Valid: true},
			News:         sql.NullString{String: "Some news content", Valid: true},
			Occurred:     sql.NullTime{Time: time.Now(), Valid: true},
			Timezone:     sql.NullString{String: "UTC", Valid: true},
			Comments:     sql.NullInt32{Int32: 5, Valid: true},
		},
	}
}

func (m *mockCD) NewsLabels(newsID int32, userID int32) []templates.TopicLabel {
	return m.labels
}

func (m *mockCD) LocalTimeIn(t time.Time, _ string) time.Time { return t }
func (m *mockCD) LocalTime(t time.Time) time.Time             { return t }
func (m *mockCD) SelectedThreadID() int32                     { return 0 }
func (m *mockCD) ShowEditNews(int32, int32) bool              { return false }
func (m *mockCD) IsAdmin() bool                               { return false }
func (m *mockCD) IsAdminMode() bool                           { return false }
func (m *mockCD) NewsAnnouncement(int32) *db.SiteAnnouncement { return nil }
func (m *mockCD) SelectedThreadCanReply() bool                { return false } // Added to match previous test dependencies if needed

func TestNewsListingDismissLink(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		funcMap := template.FuncMap{
			"cd":          func() *mockCD { return &mockCD{labels: []templates.TopicLabel{{Name: "unread", Type: "private"}}} },
			"csrfField":   func() template.HTML { return "" },
			"now":         func() time.Time { return time.Unix(1000, 0) },
			"a4code2html": func(s string) template.HTML { return template.HTML(s) },
			"add":         func(a, b int) int { return a + b },
			"since":       func(time.Time, time.Time) string { return "" },
			"assetHash":   func(s string) string { return s },
			"dict": func(values ...interface{}) (map[string]interface{}, error) {
				if len(values)%2 != 0 {
					return nil, errors.New("invalid dict call")
				}
				dict := make(map[string]interface{}, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, errors.New("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
			"printf": fmt.Sprintf,
		}

		base := filepath.Join("..", "..", "core", "templates", "site")
		tmpl := template.Must(template.New("root").Funcs(funcMap).ParseFiles(
			filepath.Join(base, "news", "page.gohtml"),
			filepath.Join(base, "news", "post.gohtml"),
			filepath.Join(base, "forum", "topicLabels.gohtml"), // Needed until we replace it
			filepath.Join(base, "_share.gohtml"),
		))

		// Define stubs for other templates used in page.gohtml
		tmpl = template.Must(tmpl.Parse(`{{ define "head" }}{{ end }}
{{ define "tail" }}{{ end }}
{{ define "threadComments" }}{{ end }}
{{ define "comment" }}{{ end }}
{{ define "languageCombobox" }}{{ end }}`))

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "page.gohtml", nil); err != nil {
			t.Fatalf("render: %v", err)
		}

		out := buf.String()

		// We expect the dismiss link to be present
		expectedLink := "/news/123/labels?task=Mark+Thread+Read"
		if !strings.Contains(out, expectedLink) {
			t.Errorf("Expected dismiss link %q not found in output:\n%s", expectedLink, out)
		}

		expectedOnclick := "replaceContent(event, this.href, 'news-labels-123')"
		if !strings.Contains(out, expectedOnclick) {
			t.Errorf("Expected onclick %q not found in output:\n%s", expectedOnclick, out)
		}

		if !strings.Contains(out, "(x)") {
			t.Errorf("Expected dismiss text '(x)' not found in output")
		}

		if !strings.Contains(out, `data-timestamp="1000"`) {
			t.Errorf("Expected data-timestamp=\"1000\" not found")
		}

		if !strings.Contains(out, `id="news-labels-123"`) {
			t.Errorf("Expected id=\"news-labels-123\" not found")
		}
	})
}
