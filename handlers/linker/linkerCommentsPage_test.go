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

func linkerItemForUser(linkID, authorID, threadID int32) *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow {
	return &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:          linkID,
		LanguageID:  sql.NullInt32{Int32: 1, Valid: true},
		AuthorID:    authorID,
		CategoryID:  sql.NullInt32{Int32: 1, Valid: true},
		ThreadID:    threadID,
		Title:       sql.NullString{String: "t", Valid: true},
		Url:         sql.NullString{String: "http://u", Valid: true},
		Description: sql.NullString{String: "d", Valid: true},
		Listed:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Timezone:    sql.NullString{String: time.Local.String(), Valid: true},
		Username:    sql.NullString{String: "bob", Valid: true},
		Title_2:     sql.NullString{String: "cat", Valid: true},
	}
}

func threadRow(threadID int32) *db.GetThreadLastPosterAndPermsRow {
	return &db.GetThreadLastPosterAndPermsRow{
		Idforumthread:          threadID,
		Firstpost:              1,
		Lastposter:             1,
		ForumtopicIdforumtopic: 1,
		Comments:               sql.NullInt32{Int32: 0, Valid: true},
		Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Locked:                 sql.NullBool{Bool: false, Valid: true},
		Lastposterusername:     sql.NullString{String: "bob", Valid: true},
	}
}

func commentRow(commentID, userID, threadID int32, isOwner bool) *db.GetCommentsBySectionThreadIdForUserRow {
	return &db.GetCommentsBySectionThreadIdForUserRow{
		Idcomments:     commentID,
		ForumthreadID:  threadID,
		UsersIdusers:   userID,
		Written:        sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Text:           sql.NullString{String: "text", Valid: true},
		Timezone:       sql.NullString{String: time.Local.String(), Valid: true},
		DeletedAt:      sql.NullTime{},
		LastIndex:      sql.NullTime{},
		Posterusername: sql.NullString{String: "bob", Valid: true},
		IsOwner:        isOwner,
	}
}

func TestCommentsPageAllowsGlobalViewGrant(t *testing.T) {
	queries := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: linkerItemForUser(1, 2, 1),
		GetThreadLastPosterAndPermsRow:                                          threadRow(1),
	}
	writeTempCommentsTemplate(t, `ok`)
	store := sessions.NewCookieStore([]byte("t"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/linker/comments/1", nil)
	req = mux.SetURLVars(req, map[string]string{"link": "1"})
	w := httptest.NewRecorder()
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(2)
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true
	cd.UserID = 2
	cd.AdminMode = true
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	CommentsPage(rr, req)

	if len(queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls) != 1 {
		t.Fatalf("expected linker item lookup, got %d", len(queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls))
	}
	call := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserCalls[0]
	if call.ID != 1 || call.ViewerID != 2 {
		t.Fatalf("unexpected linker item params %+v", call)
	}
	if len(queries.GetCommentsBySectionThreadIdForUserCalls) != 1 {
		t.Fatalf("expected one comments query, got %d", len(queries.GetCommentsBySectionThreadIdForUserCalls))
	}
	if len(queries.GetThreadLastPosterAndPermsCalls) != 1 {
		t.Fatalf("expected thread permissions lookup, got %d", len(queries.GetThreadLastPosterAndPermsCalls))
	}
	if len(queries.SystemCheckGrantCalls) != 0 {
		t.Fatalf("expected admin mode to skip grant checks, got %d", len(queries.SystemCheckGrantCalls))
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

func setGrantResponses(stub *db.QuerierStub, responses map[string]error) {
	stub.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if err, ok := responses[p.Action]; ok {
			if err != nil {
				return 0, err
			}
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}
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

	queries := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: linkerItemForUser(1, 2, 1),
		GetCommentsBySectionThreadIdForUserRows: []*db.GetCommentsBySectionThreadIdForUserRow{
			commentRow(5, 2, 1, true),
		},
		GetThreadLastPosterAndPermsRow: threadRow(1),
	}
	w, req, cd := newCommentsPageRequest(t, queries, nil, 2)

	setGrantResponses(queries, map[string]error{
		"reply":    nil,
		"view":     nil,
		"edit-any": sql.ErrNoRows,
		"edit":     nil,
	})

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "can-edit" {
		t.Fatalf("expected edit controls, got %q", got)
	}
	if len(queries.SystemCheckGrantCalls) != 4 {
		t.Fatalf("expected four grant checks, got %d", len(queries.SystemCheckGrantCalls))
	}
	actions := []string{
		queries.SystemCheckGrantCalls[0].Action,
		queries.SystemCheckGrantCalls[1].Action,
		queries.SystemCheckGrantCalls[2].Action,
		queries.SystemCheckGrantCalls[3].Action,
	}
	expectedActions := []string{"reply", "view", "edit-any", "edit"}
	for i, act := range expectedActions {
		if actions[i] != act {
			t.Fatalf("expected action %q at %d, got %q", act, i, actions[i])
		}
	}
	if len(queries.GetCommentsBySectionThreadIdForUserCalls) != 1 {
		t.Fatalf("expected one comments query, got %d", len(queries.GetCommentsBySectionThreadIdForUserCalls))
	}

	// Prevent unused warning in case the handler changes.
	_ = cd
}

func TestCommentsPageEditControlsRequireGrantNotRole(t *testing.T) {
	writeTempCommentsTemplate(t, `{{if .CanEdit}}can-edit{{else}}no-edit{{end}}`)

	queries := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: linkerItemForUser(1, 2, 1),
		GetCommentsBySectionThreadIdForUserRows:                                 []*db.GetCommentsBySectionThreadIdForUserRow{},
		GetThreadLastPosterAndPermsRow:                                          threadRow(1),
	}
	w, req, cd := newCommentsPageRequest(t, queries, []string{"administrator"}, 3)
	cd.AdminMode = false

	setGrantResponses(queries, map[string]error{
		"reply":    sql.ErrNoRows,
		"view":     nil,
		"edit-any": sql.ErrNoRows,
		"edit":     sql.ErrNoRows,
	})

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "no-edit" {
		t.Fatalf("expected edit controls to be hidden without grants, got %q", got)
	}
	if len(queries.SystemCheckGrantCalls) != 4 {
		t.Fatalf("expected grant checks for non-admin, got %d", len(queries.SystemCheckGrantCalls))
	}

	_ = cd
}

func TestCommentsPageEditControlsAllowAdminMode(t *testing.T) {
	writeTempCommentsTemplate(t, `{{if .CanEdit}}can-edit{{else}}no-edit{{end}}`)

	queries := &db.QuerierStub{
		GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow: linkerItemForUser(1, 2, 1),
		GetCommentsBySectionThreadIdForUserRows: []*db.GetCommentsBySectionThreadIdForUserRow{
			commentRow(9, 2, 1, false),
		},
		GetThreadLastPosterAndPermsRow: threadRow(1),
	}
	w, req, cd := newCommentsPageRequest(t, queries, []string{"administrator"}, 4)
	cd.AdminMode = true

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "can-edit" {
		t.Fatalf("expected admin mode to allow edit controls, got %q", got)
	}
	if len(queries.SystemCheckGrantCalls) != 0 {
		t.Fatalf("expected admin mode to skip grant checks, got %d", len(queries.SystemCheckGrantCalls))
	}

	_ = cd
}
