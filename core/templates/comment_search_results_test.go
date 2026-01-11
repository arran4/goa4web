package templates_test

import (
	"bytes"
	"database/sql"
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

type fakeSearchCD struct {
	comments []*db.GetCommentsByIdsForUserWithThreadInfoRow
}

func (f *fakeSearchCD) SearchComments() []*db.GetCommentsByIdsForUserWithThreadInfoRow {
	return f.comments
}
func (*fakeSearchCD) SearchCommentsNoResults() bool               { return false }
func (*fakeSearchCD) SearchCommentsEmptyWords() bool              { return false }
func (*fakeSearchCD) LocalTimeIn(t time.Time, _ string) time.Time { return t }

func TestCommentSearchResultsHighlightsAndEscapes(t *testing.T) {
	cd := &fakeSearchCD{
		comments: []*db.GetCommentsByIdsForUserWithThreadInfoRow{{
			Idforumcategory:    sql.NullInt32{Int32: 1, Valid: true},
			ForumcategoryTitle: sql.NullString{String: "Category", Valid: true},
			ForumtopicTitle:    sql.NullString{String: "Topic", Valid: true},
			Idforumtopic:       sql.NullInt32{Int32: 2, Valid: true},
			Idforumthread:      sql.NullInt32{Int32: 3, Valid: true},
			ForumthreadID:      3,
			Posterusername:     sql.NullString{String: "poster", Valid: true},
			Written:            sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Timezone:           sql.NullString{String: "UTC", Valid: true},
			Text:               sql.NullString{String: "<b>match</b> term", Valid: true},
		}},
	}
	funcs := template.FuncMap{
		"cd": func() *fakeSearchCD { return cd },
		"topicTitleOrDefault": func(title string) string {
			return title
		},
		"highlightSearch": func(s string) template.HTML {
			return common.HighlightSearchTerms(s, []string{"match"})
		},
	}
	tmpl := template.Must(template.New("root").Funcs(funcs).ParseFiles("site/commentSearchReslts.gohtml"))

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "commentSearchResults", nil); err != nil {
		t.Fatalf("execute template: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "<mark>match</mark>") {
		t.Fatalf("expected highlighted search term, got: %s", out)
	}
	if strings.Contains(out, "<b>") {
		t.Fatalf("expected escaped HTML in comment text, got raw tag: %s", out)
	}
	if !strings.Contains(out, "&lt;b&gt;") {
		t.Fatalf("expected escaped HTML tag, got: %s", out)
	}
}
