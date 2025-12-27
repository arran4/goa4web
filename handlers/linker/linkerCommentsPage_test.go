package linker

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

func TestCommentsPageAllowsGlobalViewGrant(t *testing.T) {
	writeTempCommentsTemplate(t, `{{.Link.Title.String}}|{{(index .Comments 0).Posterusername.String}}|replyable={{.IsReplyable}}`)

	stub := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
			ID:         1,
			LanguageID: sql.NullInt32{Int32: 1, Valid: true},
			AuthorID:   2,
			CategoryID: sql.NullInt32{Int32: 1, Valid: true},
			ThreadID:   1,
			Title:      sql.NullString{String: "t", Valid: true},
			Url:        sql.NullString{String: "http://u", Valid: true},
			Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
			Username:   sql.NullString{String: "bob", Valid: true},
			Title_2:    sql.NullString{String: "cat", Valid: true},
		},
		GetCommentsBySectionThreadIdForUserRows: []*db.GetCommentsBySectionThreadIdForUserRow{{
			Idcomments:     5,
			ForumthreadID:  1,
			UsersIdusers:   2,
			Written:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Text:           sql.NullString{String: "comment", Valid: true},
			Posterusername: sql.NullString{String: "bob", Valid: true},
			IsOwner:        true,
		}},
		GetThreadLastPosterAndPermsRow: &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Lastposterusername:     sql.NullString{String: "bob", Valid: true},
		},
	}

	w, req, cd := newCommentsPageRequest(t, stub, []string{"administrator"}, 2)
	cd.AdminMode = true

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "t|bob|replyable=true" {
		t.Fatalf("expected stubbed data to render, got %q", got)
	}
	if len(stub.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls) != 1 {
		t.Fatalf("expected link fetch to be called once, got %d", len(stub.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls))
	}
	if len(stub.GetCommentsBySectionThreadIdForUserCalls) != 1 {
		t.Fatalf("expected comments fetch to be called once, got %d", len(stub.GetCommentsBySectionThreadIdForUserCalls))
	}
	if len(stub.GetThreadLastPosterAndPermsCalls) != 1 {
		t.Fatalf("expected thread fetch to be called once, got %d", len(stub.GetThreadLastPosterAndPermsCalls))
	}
}

func newCommentsPageRequest(t *testing.T, queries db.Querier, roles []string, userID int32) (*httptest.ResponseRecorder, *http.Request, *common.CoreData) {
	t.Helper()

	store := sessions.NewCookieStore([]byte("t"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/linker/comments/1", nil)
	req = mux.SetURLVars(req, map[string]string{"link": "1"})
	w := httptest.NewRecorder()
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = userID
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles(roles))
	cd.UserID = userID
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return w, req, cd
}

func writeTempCommentsTemplate(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	siteDir := filepath.Join(dir, "site")
	if err := os.Mkdir(siteDir, 0o755); err != nil {
		t.Fatalf("create site dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "commentsPage.gohtml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	templates.SetDir(dir)
	t.Cleanup(func() { templates.SetDir("") })
}

func TestCommentsPageEditControlsUseEditGrant(t *testing.T) {
	writeTempCommentsTemplate(t, `{{if .CanEdit}}can-edit{{else}}no-edit{{end}}`)

	stub := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
			ID:         1,
			LanguageID: sql.NullInt32{Int32: 1, Valid: true},
			AuthorID:   2,
			CategoryID: sql.NullInt32{Int32: 1, Valid: true},
			ThreadID:   1,
			Title:      sql.NullString{String: "t", Valid: true},
			Url:        sql.NullString{String: "http://u", Valid: true},
			Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
			Username:   sql.NullString{String: "bob", Valid: true},
			Title_2:    sql.NullString{String: "cat", Valid: true},
		},
		GetCommentsBySectionThreadIdForUserRows: []*db.GetCommentsBySectionThreadIdForUserRow{{
			Idcomments:     5,
			ForumthreadID:  1,
			UsersIdusers:   2,
			Written:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Text:           sql.NullString{String: "text", Valid: true},
			Posterusername: sql.NullString{String: "bob", Valid: true},
			IsOwner:        true,
		}},
		GetThreadLastPosterAndPermsRow: &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Lastposterusername:     sql.NullString{String: "bob", Valid: true},
		},
	}
	stub.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.Action == "view" || p.Action == "reply" || p.Action == "edit" {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}

	w, req, _ := newCommentsPageRequest(t, stub, nil, 2)

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "can-edit" {
		t.Fatalf("expected edit controls, got %q", got)
	}
	if len(stub.SystemCheckGrantCalls) != 4 {
		t.Fatalf("expected 4 grant checks, got %d", len(stub.SystemCheckGrantCalls))
	}
}

func TestCommentsPageEditControlsRequireGrantNotRole(t *testing.T) {
	writeTempCommentsTemplate(t, `{{if .CanEdit}}can-edit{{else}}no-edit{{end}}`)

	stub := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
			ID:         1,
			LanguageID: sql.NullInt32{Int32: 1, Valid: true},
			AuthorID:   2,
			CategoryID: sql.NullInt32{Int32: 1, Valid: true},
			ThreadID:   1,
			Title:      sql.NullString{String: "t", Valid: true},
			Url:        sql.NullString{String: "http://u", Valid: true},
			Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
			Username:   sql.NullString{String: "bob", Valid: true},
			Title_2:    sql.NullString{String: "cat", Valid: true},
		},
		GetThreadLastPosterAndPermsRow: &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Lastposterusername:     sql.NullString{String: "bob", Valid: true},
		},
	}
	stub.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		switch p.Action {
		case "view":
			return 1, nil
		default:
			return 0, sql.ErrNoRows
		}
	}

	w, req, cd := newCommentsPageRequest(t, stub, []string{"administrator"}, 3)
	cd.AdminMode = false

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "no-edit" {
		t.Fatalf("expected edit controls to be hidden without grants, got %q", got)
	}
	if got := len(stub.SystemCheckGrantCalls); got != 4 {
		t.Fatalf("expected 4 grant checks, got %d", got)
	}

}

func TestCommentsPageEditControlsAllowAdminMode(t *testing.T) {
	writeTempCommentsTemplate(t, `{{if .CanEdit}}can-edit{{else}}no-edit{{end}}`)

	stub := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
			ID:         1,
			LanguageID: sql.NullInt32{Int32: 1, Valid: true},
			AuthorID:   2,
			CategoryID: sql.NullInt32{Int32: 1, Valid: true},
			ThreadID:   1,
			Title:      sql.NullString{String: "t", Valid: true},
			Url:        sql.NullString{String: "http://u", Valid: true},
			Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
			Username:   sql.NullString{String: "bob", Valid: true},
			Title_2:    sql.NullString{String: "cat", Valid: true},
		},
		GetCommentsBySectionThreadIdForUserRows: []*db.GetCommentsBySectionThreadIdForUserRow{{
			Idcomments:     9,
			ForumthreadID:  1,
			UsersIdusers:   2,
			Written:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Text:           sql.NullString{String: "text", Valid: true},
			Posterusername: sql.NullString{String: "bob", Valid: true},
			IsOwner:        false,
		}},
		GetThreadLastPosterAndPermsRow: &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Lastposterusername:     sql.NullString{String: "bob", Valid: true},
		},
	}

	w, req, cd := newCommentsPageRequest(t, stub, []string{"administrator"}, 4)
	cd.AdminMode = true

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "can-edit" {
		t.Fatalf("expected admin mode to allow edit controls, got %q", got)
	}
}
