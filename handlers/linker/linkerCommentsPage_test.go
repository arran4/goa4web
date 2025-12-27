package linker

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
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

	mock.ExpectQuery(regexp.QuoteMeta("SELECT l.id")).
		WithArgs(int32(2), sqlmock.AnyArg(), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "language_id", "author_id", "category_id", "thread_id", "title", "url", "description", "listed", "timezone", "username", "title"}).
			AddRow(1, 1, 2, 1, 1, "t", "http://u", "d", time.Unix(0, 0), time.Local.String(), "bob", "cat"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments")).
		WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), "linker", sql.NullString{String: "link", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "posterusername", "is_owner"}))

	threadRows := sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername"}).
		AddRow(1, 1, 1, 1, 0, time.Unix(0, 0), false, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT th.idforumthread")).
		WithArgs(int32(2), int32(1), int32(2), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(threadRows)

	rr := httptest.NewRecorder()
	CommentsPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
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

func expectGrantCheck(mock sqlmock.Sqlmock, viewerID, itemID int32, action string, err error) {
	expect := mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM grants g")).WithArgs(
		viewerID,
		"linker",
		sql.NullString{String: "link", Valid: true},
		action,
		sql.NullInt32{Int32: itemID, Valid: true},
		sql.NullInt32{Int32: viewerID, Valid: viewerID != 0},
	)
	if err != nil {
		expect.WillReturnError(err)
		return
	}
	expect.WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	writeTempCommentsTemplate(t, `{{if .CanEdit}}can-edit{{else}}no-edit{{end}}`)

	queries := db.New(conn)
	w, req, cd := newCommentsPageRequest(t, queries, nil, 2)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT l.id")).WithArgs(int32(2), sql.NullInt32{Int32: 2, Valid: true}, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "language_id", "author_id", "category_id", "thread_id", "title", "url", "description", "listed", "timezone", "username", "title"}).
			AddRow(1, 1, 2, 1, 1, "t", "http://u", "d", time.Unix(0, 0), time.Local.String(), "bob", "cat"))

	expectGrantCheck(mock, 2, 1, "reply", nil)
	expectGrantCheck(mock, 2, 1, "view", nil)
	expectGrantCheck(mock, 2, 1, "edit-any", sql.ErrNoRows)
	expectGrantCheck(mock, 2, 1, "edit", nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments")).WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), "linker", sql.NullString{String: "link", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "posterusername", "is_owner"}).
			AddRow(5, 1, 2, sql.NullInt32{}, sql.NullTime{Time: time.Unix(0, 0), Valid: true}, sql.NullString{String: "text", Valid: true}, sql.NullString{String: time.Local.String(), Valid: true}, sql.NullTime{}, sql.NullTime{}, sql.NullString{String: "bob", Valid: true}, true))

	threadRows := sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername"}).
		AddRow(1, 1, 1, 1, 0, time.Unix(0, 0), false, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT th.idforumthread")).WithArgs(int32(2), int32(1), int32(2), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).WillReturnRows(threadRows)

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "can-edit" {
		t.Fatalf("expected edit controls, got %q", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	// Prevent unused warning in case the handler changes.
	_ = cd
}

func TestCommentsPageEditControlsRequireGrantNotRole(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	writeTempCommentsTemplate(t, `{{if .CanEdit}}can-edit{{else}}no-edit{{end}}`)

	queries := db.New(conn)
	w, req, cd := newCommentsPageRequest(t, queries, []string{"administrator"}, 3)
	cd.AdminMode = false

	mock.ExpectQuery(regexp.QuoteMeta("SELECT l.id")).WithArgs(int32(3), sql.NullInt32{Int32: 3, Valid: true}, int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "language_id", "author_id", "category_id", "thread_id", "title", "url", "description", "listed", "timezone", "username", "title"}).
			AddRow(1, 1, 2, 1, 1, "t", "http://u", "d", time.Unix(0, 0), time.Local.String(), "bob", "cat"))

	expectGrantCheck(mock, 3, 1, "reply", sql.ErrNoRows)
	expectGrantCheck(mock, 3, 1, "view", nil)
	expectGrantCheck(mock, 3, 1, "edit-any", sql.ErrNoRows)
	expectGrantCheck(mock, 3, 1, "edit", sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments")).WithArgs(int32(3), int32(3), int32(1), int32(3), int32(3), "linker", sql.NullString{String: "link", Valid: true}, sql.NullInt32{Int32: 3, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "posterusername", "is_owner"}))

	threadRows := sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername"}).
		AddRow(1, 1, 1, 1, 0, time.Unix(0, 0), false, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT th.idforumthread")).WithArgs(int32(3), int32(1), int32(3), int32(3), int32(3), sql.NullInt32{Int32: 3, Valid: true}).WillReturnRows(threadRows)

	CommentsPage(w, req)

	if got := strings.TrimSpace(w.Body.String()); got != "no-edit" {
		t.Fatalf("expected edit controls to be hidden without grants, got %q", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	_ = cd
}
