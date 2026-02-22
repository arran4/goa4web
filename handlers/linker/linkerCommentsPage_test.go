package linker

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCommentsPage(t *testing.T) {
	t.Run("Happy Path - Allows Global View Grant", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow = &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
			ID:         1,
			LanguageID: sql.NullInt32{Int32: 1, Valid: true},
			AuthorID:   2,
			CategoryID: sql.NullInt32{Int32: 1, Valid: true},
			ThreadID:   1,
			Title:      sql.NullString{String: "t", Valid: true},
			Url:        sql.NullString{String: "http://u", Valid: true},
			Listed:     sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
			Username:   sql.NullString{String: "bob", Valid: true},
			Title_2:    sql.NullString{String: "cat", Valid: true},
		}
		queries.GetCommentsBySectionThreadIdForUserReturns = []*db.GetCommentsBySectionThreadIdForUserRow{}
		queries.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Comments:               sql.NullInt32{Int32: 0, Valid: true},
			Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Locked:                 sql.NullBool{Bool: false, Valid: true},
		}
		queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{
			{Name: "administrator", IsAdmin: true},
		}
		store := sessions.NewCookieStore([]byte("t"))
		core.Store = store
		core.SessionName = "test-session"

		req := httptest.NewRequest("GET", "/linker/comments/1", nil)
		req = mux.SetURLVars(req, map[string]string{"link": "1"})
		w := httptest.NewRecorder()
		sess := testhelpers.Must(store.Get(req, core.SessionName))
		sess.Values["UID"] = int32(2)
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles([]string{"administrator"}))
		cd.UserID = 2
		cd.AdminMode = true
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		CommentsPage(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func newCommentsPageRequest(t *testing.T, queries db.Querier, roles []string, userID int32) (*httptest.ResponseRecorder, *http.Request, *common.CoreData) {
	t.Helper()

	store := sessions.NewCookieStore([]byte("t"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/linker/comments/1", nil)
	req = mux.SetURLVars(req, map[string]string{"link": "1"})
	w := httptest.NewRecorder()
	sess := testhelpers.Must(store.Get(req, core.SessionName))
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

func writeTempCommentsTemplate(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	siteDir := filepath.Join(dir, "site")
	if err := os.Mkdir(siteDir, 0o755); err != nil {
		t.Fatalf("create site dir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(siteDir, "linker"), 0o755); err != nil {
		t.Fatalf("create site/linker dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "linker", "commentsPage.gohtml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	return dir
}

func TestCommentsPageEditControls(t *testing.T) {
	t.Run("Use Edit Grant", commentsPageEditControlsUseEditGrant)
	t.Run("Require Grant Not Role", commentsPageEditControlsRequireGrantNotRole)
	t.Run("Allow Admin Mode", commentsPageEditControlsAllowAdminMode)
}

func commentsPageEditControlsUseEditGrant(t *testing.T) {
	dir := writeTempCommentsTemplate(t, "{{ if .CanEdit }}EDIT_CONTROLS{{ end }}")

	queries := testhelpers.NewQuerierStub()
	queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow = &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:       1,
		ThreadID: 1,
		Title:    sql.NullString{String: "Link Title", Valid: true},
	}
	queries.GetCommentsBySectionThreadIdForUserReturns = []*db.GetCommentsBySectionThreadIdForUserRow{}
	queries.GetThreadBySectionThreadIDForReplierReturn = &db.Forumthread{
		Idforumthread: 1,
	}
	queries.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
		Idforumthread: 1,
	}

	// Map to track grant checks
	grantChecks := make(map[string]bool)

	queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		key := fmt.Sprintf("%s:%s:%s", arg.Section, arg.Item.String, arg.Action)
		grantChecks[key] = true

		// Allow 'view' so page renders
		if arg.Section == "linker" && arg.Item.String == "link" && arg.Action == "view" {
			return 1, nil
		}
		// Allow 'edit' to test the specific case
		if arg.Section == "linker" && arg.Item.String == "link" && arg.Action == "edit" {
			return 1, nil
		}

		return 0, sql.ErrNoRows
	}

	w, req, cd := newCommentsPageRequest(t, queries, []string{"user"}, 2)
	cd.Config.TemplatesDir = dir

	CommentsPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}

	body := w.Body.String()
	if body != "EDIT_CONTROLS" {
		t.Errorf("Expected 'EDIT_CONTROLS' in body, got: %q", body)
	}

	// Verify that the edit grant was checked
	if !grantChecks["linker:link:edit"] {
		t.Error("Expected check for 'edit' grant")
	}
}

func commentsPageEditControlsRequireGrantNotRole(t *testing.T) {
	dir := writeTempCommentsTemplate(t, `CanEdit: {{.CanEdit}}`)

	queries := testhelpers.NewQuerierStub()
	queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow = &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:         1,
		LanguageID: sql.NullInt32{Int32: 1, Valid: true},
		AuthorID:   2,
		CategoryID: sql.NullInt32{Int32: 1, Valid: true},
		ThreadID:   1,
		Title:      sql.NullString{String: "t", Valid: true},
		Url:        sql.NullString{String: "http://u", Valid: true},
		Listed:     sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
		Username:   sql.NullString{String: "bob", Valid: true},
		Title_2:    sql.NullString{String: "cat", Valid: true},
	}
	queries.GetCommentsBySectionThreadIdForUserReturns = []*db.GetCommentsBySectionThreadIdForUserRow{}
	queries.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
		Idforumthread:          1,
		Firstpost:              1,
		Lastposter:             1,
		ForumtopicIdforumtopic: 1,
		Comments:               sql.NullInt32{Int32: 0, Valid: true},
		Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Locked:                 sql.NullBool{Bool: false, Valid: true},
	}
	queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{
		{Name: "user", IsAdmin: false},
	}
	queries.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.Action == "view" {
			return 1, nil
		}
		if p.Action == "edit" {
			return 1, nil
		}
		return 0, sql.ErrNoRows
	}

	w, req, cd := newCommentsPageRequest(t, queries, []string{"user"}, 2)
	cd.Config.TemplatesDir = dir

	CommentsPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "CanEdit: true" {
		t.Errorf("Expected CanEdit: true, got %q", body)
	}

	// Verify that if we have the role but NOT the grant, it is false.
	// We reuse everything but modify SystemCheckGrantFn
	queries.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
		if p.Action == "view" {
			return 1, nil
		}
		// Deny edit
		return 0, sql.ErrNoRows
	}

	w, req, cd = newCommentsPageRequest(t, queries, []string{"user"}, 2)
	cd.Config.TemplatesDir = dir
	CommentsPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	body = w.Body.String()
	if body != "CanEdit: false" {
		t.Errorf("Expected CanEdit: false (since grant is denied), got %q", body)
	}
}

func commentsPageEditControlsAllowAdminMode(t *testing.T) {
	dir := writeTempCommentsTemplate(t, `{{ range .Comments }}{{ call $.AdminURL . }}{{ end }}`)

	t.Logf("Templates in site: %v", templates.ListSiteTemplateNames())

	linkID := 1
	threadID := 1
	userID := int32(2)
	commentID := int32(100)

	queries := testhelpers.NewQuerierStub(testhelpers.WithGrantResult(true))
	queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow = &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:         int32(linkID),
		LanguageID: sql.NullInt32{Int32: 1, Valid: true},
		AuthorID:   userID,
		CategoryID: sql.NullInt32{Int32: 1, Valid: true},
		ThreadID:   int32(threadID),
		Title:      sql.NullString{String: "t", Valid: true},
		Url:        sql.NullString{String: "http://u", Valid: true},
		Listed:     sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Timezone:   sql.NullString{String: time.Local.String(), Valid: true},
		Username:   sql.NullString{String: "bob", Valid: true},
		Title_2:    sql.NullString{String: "cat", Valid: true},
	}
	queries.GetCommentsBySectionThreadIdForUserReturns = []*db.GetCommentsBySectionThreadIdForUserRow{
		{
			Idcomments:    int32(commentID),
			ForumthreadID: int32(threadID),
			Text:          sql.NullString{String: "some comment", Valid: true},
			IsOwner:       false,
		},
	}
	queries.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
		Idforumthread:          int32(threadID),
		Firstpost:              1,
		Lastposter:             1,
		ForumtopicIdforumtopic: 1,
		Comments:               sql.NullInt32{Int32: 0, Valid: true},
		Lastaddition:           sql.NullTime{Time: time.Unix(0, 0), Valid: true},
		Locked:                 sql.NullBool{Bool: false, Valid: true},
	}
	queries.GetPermissionsByUserIDReturns = []*db.GetPermissionsByUserIDRow{
		{Name: "administrator", IsAdmin: true},
	}

	w, req, cd := newCommentsPageRequest(t, queries, []string{"administrator"}, userID)
	cd.AdminMode = true
	cd.Config.TemplatesDir = dir

	CommentsPage(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	expectedAdminURL := "/admin/comment/100"
	if !strings.Contains(body, expectedAdminURL) {
		t.Errorf("expected admin URL %q in body, got: %q", expectedAdminURL, body)
	}
}
