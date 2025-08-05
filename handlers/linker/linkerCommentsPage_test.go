package linker

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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
	cd.UserID = 2
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof FROM language")).
		WillReturnRows(sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en"))

	linkQuery := "SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title FROM linker l JOIN users u ON l.users_idusers = u.idusers JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory WHERE l.idlinker = ? AND l.listed IS NOT NULL AND l.deleted_at IS NULL AND EXISTS ( SELECT 1 FROM grants g WHERE g.section='linker' AND (g.item='link' OR g.item IS NULL) AND g.action IN ('view','comment','reply') AND g.active=1 AND (g.item_id = l.idlinker OR g.item_id IS NULL) AND (g.user_id = ? OR g.user_id IS NULL) AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids)) ) LIMIT 1"
	mock.ExpectQuery(regexp.QuoteMeta(linkQuery)).
		WithArgs(int32(2), int32(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"idlinker", "language_idlanguage", "users_idusers", "linker_category_id", "forumthread_id", "title", "url", "description", "listed", "username", "title_2"}).
			AddRow(1, 1, 2, 1, 1, "t", "http://u", "d", time.Unix(0, 0), "bob", "cat"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments")).
		WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "last_index", "posterusername", "is_owner"}))

	threadRows := sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername"}).
		AddRow(1, 1, 1, 1, 0, time.Unix(0, 0), false, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT th.idforumthread")).
		WithArgs(int32(2), int32(1), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(threadRows)

	rr := httptest.NewRecorder()
	CommentsPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
