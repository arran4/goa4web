package templates

import (
	"bytes"
	"database/sql"
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

type fakeCD struct{}

func (*fakeCD) CommentEditing(*db.GetCommentsByThreadIdForUserRow) bool       { return false }
func (*fakeCD) SelectedThreadCanReply() bool                                  { return false }
func (*fakeCD) CanEditComment(*db.GetCommentsByThreadIdForUserRow) bool       { return false }
func (*fakeCD) CommentEditURL(*db.GetCommentsByThreadIdForUserRow) string     { return "" }
func (*fakeCD) CommentEditSaveURL(*db.GetCommentsByThreadIdForUserRow) string { return "" }
func (*fakeCD) CommentAdminURL(*db.GetCommentsByThreadIdForUserRow) string    { return "" }

func TestCommentTimestampSelfLink(t *testing.T) {
	funcMap := template.FuncMap{
		"cd":          func() *fakeCD { return &fakeCD{} },
		"localTime":   func(t time.Time) time.Time { return t },
		"a4code2html": func(s string) template.HTML { return template.HTML(s) },
		"csrfField":   func() template.HTML { return "" },
	}
	tmpl := template.Must(template.New("root").Funcs(funcMap).ParseFiles("site/comment.gohtml", "site/languageCombobox.gohtml"))
	var buf bytes.Buffer
	cmt := &db.GetCommentsByThreadIdForUserRow{
		Idcomments:     1,
		Written:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Text:           sql.NullString{String: "hi", Valid: true},
		Posterusername: sql.NullString{String: "alice", Valid: true},
		IsOwner:        true,
	}
	if err := tmpl.ExecuteTemplate(&buf, "comment", cmt); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "id=\"comment-1\" href=\"#comment-1\"") {
		t.Fatalf("missing comment self-link: %s", out)
	}
}
