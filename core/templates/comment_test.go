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

type fakeCD struct{
	UserID int32
}

func (*fakeCD) CommentEditing(*db.GetCommentsByThreadIdForUserRow) bool       { return false }
func (*fakeCD) SelectedThreadCanReply() bool                                  { return false }
func (*fakeCD) CanEditComment(*db.GetCommentsByThreadIdForUserRow) bool       { return false }
func (*fakeCD) CommentEditURL(*db.GetCommentsByThreadIdForUserRow) string     { return "" }
func (*fakeCD) CommentEditSaveURL(*db.GetCommentsByThreadIdForUserRow) string { return "" }
func (*fakeCD) CommentAdminURL(*db.GetCommentsByThreadIdForUserRow) string    { return "" }
func (*fakeCD) Location() *time.Location                                      { return time.UTC }
func (*fakeCD) LocalTime(t time.Time) time.Time                               { return t }
func (*fakeCD) LocalTimeIn(t time.Time, _ string) time.Time                   { return t }

func TestCommentTimestampSelfLink(t *testing.T) {
	funcMap := template.FuncMap{
		"cd":          func() *fakeCD { return &fakeCD{} },
		"a4code2html": func(s string) template.HTML { return template.HTML(s) },
		"csrfField":   func() template.HTML { return "" },
		"since":       func(time.Time, time.Time) string { return "" },
		"timeAgo":     func(time.Time) string { return "" },
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
	data := map[string]any{"Comment": cmt, "Number": 1, "Prev": 0}
	if err := tmpl.ExecuteTemplate(&buf, "comment", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "href=\"#c1\">#1</a>") {
		t.Fatalf("missing comment self-link: %s", out)
	}
}

func TestCommentUsernameBold(t *testing.T) {
	tests := []struct {
		name        string
		currentUser int32
		commentUser int32
		username    string
		wantBold    bool
	}{
		{
			name:        "SameUser",
			currentUser: 123,
			commentUser: 123,
			username:    "alice",
			wantBold:    false,
		},
		{
			name:        "DifferentUser",
			currentUser: 123,
			commentUser: 456,
			username:    "bob",
			wantBold:    true,
		},
		{
			name:        "GuestViewingUser",
			currentUser: 0,
			commentUser: 456,
			username:    "charlie",
			wantBold:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcMap := template.FuncMap{
				"cd":          func() *fakeCD { return &fakeCD{UserID: tt.currentUser} },
				"a4code2html": func(s string) template.HTML { return template.HTML(s) },
				"csrfField":   func() template.HTML { return "" },
				"since":       func(time.Time, time.Time) string { return "" },
				"timeAgo":     func(time.Time) string { return "" },
			}
			tmpl := template.Must(template.New("root").Funcs(funcMap).ParseFiles("site/comment.gohtml", "site/languageCombobox.gohtml"))
			var buf bytes.Buffer
			cmt := &db.GetCommentsByThreadIdForUserRow{
				Idcomments:     1,
				UsersIdusers:   tt.commentUser,
				Written:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
				Text:           sql.NullString{String: "hi", Valid: true},
				Posterusername: sql.NullString{String: tt.username, Valid: true},
				IsOwner:        tt.currentUser == tt.commentUser,
			}
			data := map[string]any{"Comment": cmt, "Number": 1, "Prev": 0}
			if err := tmpl.ExecuteTemplate(&buf, "comment", data); err != nil {
				t.Fatalf("execute template: %v", err)
			}
			out := buf.String()

			if tt.wantBold {
				if !strings.Contains(out, "<strong>"+tt.username+"</strong>") {
					t.Errorf("expected bold username for %s, got: %s", tt.username, out)
				}
			} else {
				if strings.Contains(out, "<strong>"+tt.username+"</strong>") {
					t.Errorf("expected non-bold username for %s, got: %s", tt.username, out)
				}
				if !strings.Contains(out, "<div class=\"username\">"+tt.username+"</div>") {
					t.Errorf("expected normal username for %s, got: %s", tt.username, out)
				}
			}
		})
	}
}
