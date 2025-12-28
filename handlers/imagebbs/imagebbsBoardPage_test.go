package imagebbs

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

type fakeBoardPageCD struct {
	boardID int32
	boards  []*db.Imageboard
	posts   []*db.ListImagePostsByBoardForListerRow
	t       *testing.T
}

func (f *fakeBoardPageCD) SelectedBoardID() int32 { return f.boardID }
func (f *fakeBoardPageCD) SubImageBoards(parentID int32) ([]*db.Imageboard, error) {
	f.t.Helper()
	if parentID != f.boardID {
		f.t.Fatalf("unexpected board id %d (want %d)", parentID, f.boardID)
	}
	return f.boards, nil
}
func (f *fakeBoardPageCD) SelectedBoardPosts() ([]*db.ListImagePostsByBoardForListerRow, error) {
	return f.posts, nil
}
func (*fakeBoardPageCD) LocalTimeIn(t time.Time, _ string) time.Time { return t }

func TestBoardPageRendersSubBoards(t *testing.T) {
	t.Parallel()

	cd := &fakeBoardPageCD{
		boardID: 3,
		boards: []*db.Imageboard{
			{
				Idimageboard:           4,
				ImageboardIdimageboard: sql.NullInt32{Int32: 3, Valid: true},
				Title:                  sql.NullString{String: "child", Valid: true},
				Description:            sql.NullString{String: "sub", Valid: true},
			},
		},
		posts: []*db.ListImagePostsByBoardForListerRow{
			{
				Idimagepost:            1,
				ForumthreadID:          1,
				UsersIdusers:           1,
				ImageboardIdimageboard: sql.NullInt32{Int32: 3, Valid: true},
				Posted:                 sql.NullTime{Time: time.Unix(0, 0), Valid: true},
				Timezone:               sql.NullString{String: time.Local.String(), Valid: true},
				Description:            sql.NullString{String: "desc", Valid: true},
				Thumbnail:              sql.NullString{String: "/t", Valid: true},
				Fullimage:              sql.NullString{String: "/f", Valid: true},
				FileSize:               10,
				Approved:               true,
				Comments:               sql.NullInt32{Int32: 0, Valid: true},
				Username:               sql.NullString{String: "alice", Valid: true},
			},
		},
		t: t,
	}

	funcs := template.FuncMap{
		"cd":        func() *fakeBoardPageCD { return cd },
		"csrfField": func() template.HTML { return "" },
	}

	tmpl, err := templates.LoadSiteTemplates(funcs, filepath.Join("..", "..", "core", "templates"))
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, ImagebbsBoardPageTmpl, struct{}{}); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	body := out.String()
	if !strings.Contains(body, "Sub-Boards") {
		t.Fatalf("expected sub boards in output: %s", body)
	}
	if !strings.Contains(body, "child") || !strings.Contains(body, "sub") {
		t.Fatalf("expected board details in output: %s", body)
	}
	if !strings.Contains(body, "Pictures:") {
		t.Fatalf("expected pictures in output: %s", body)
	}
	if !strings.Contains(body, "desc") || !strings.Contains(body, "alice") || !strings.Contains(body, "/imagebbs/board/3/thread/1") {
		t.Fatalf("expected post details in output: %s", body)
	}
}
