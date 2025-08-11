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

	mock.ExpectQuery(regexp.QuoteMeta("SELECT l.id")).
		WithArgs(int32(2), sqlmock.AnyArg(), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "language_id", "author_id", "category_id", "thread_id", "title", "url", "description", "listed", "timezone", "username", "title"}).
			AddRow(1, 1, 2, 1, 1, "t", "http://u", "d", time.Unix(0, 0), time.Local.String(), "bob", "cat"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments")).
		WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), "linker", sql.NullString{String: "link", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "timezone", "deleted_at", "last_index", "posterusername", "is_owner"}))

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
