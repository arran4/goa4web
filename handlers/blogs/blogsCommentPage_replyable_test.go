package blogs

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
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
	"github.com/arran4/goa4web/internal/db"
)

func setupCommentRequest(t *testing.T, queries db.Querier, store *sessions.CookieStore, blogID int) (*http.Request, *sessions.Session) {
	req := httptest.NewRequest("GET", "/blogs/blog/1/comments", nil)
	req = mux.SetURLVars(req, map[string]string{"blog": "1"})
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
	cd.LoadSelectionsFromRequest(req)
	cd.SetCurrentSection("blogs")
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	return req, sess
}

func TestCommentPageLockedThreadDisablesReply(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	queries := db.New(dbconn)
	store := sessions.NewCookieStore([]byte("t"))
	core.Store = store
	core.SessionName = "test-session"

	req, _ := setupCommentRequest(t, queries, store, 1)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof FROM language")).
		WillReturnRows(sqlmock.NewRows([]string{"idlanguage", "nameof"}))
	cd := req.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if _, err := cd.Languages(); err != nil {
		t.Fatalf("languages: %v", err)
	}

	blogRows := sqlmock.NewRows([]string{"idblogs", "forumthread_id", "users_idusers", "language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)", "is_owner"}).
		AddRow(1, 1, 2, 1, "hi", time.Unix(0, 0), "bob", 0, false)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).WillReturnRows(blogRows)

	threadRows := sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername"}).
		AddRow(1, 1, 1, 1, 0, time.Unix(0, 0), true, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT th.idforumthread")).WithArgs(int32(2), int32(1), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).WillReturnRows(threadRows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments")).
		WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), "blogs", sql.NullString{String: "entry", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "posterusername"}))

	rr := httptest.NewRecorder()
	CommentPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if regexp.MustCompile(`Reply:`).FindString(rr.Body.String()) != "" {
		t.Fatalf("reply form should be hidden")
	}
}

func TestCommentPageUnlockedThreadShowsReply(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	queries := db.New(dbconn)
	store := sessions.NewCookieStore([]byte("t"))
	core.Store = store
	core.SessionName = "test-session"

	req, _ := setupCommentRequest(t, queries, store, 1)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof FROM language")).
		WillReturnRows(sqlmock.NewRows([]string{"idlanguage", "nameof"}))
	cd := req.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if _, err := cd.Languages(); err != nil {
		t.Fatalf("languages: %v", err)
	}

	blogRows := sqlmock.NewRows([]string{"idblogs", "forumthread_id", "users_idusers", "language_idlanguage", "blog", "written", "username", "coalesce(th.comments, 0)", "is_owner"}).
		AddRow(1, 1, 2, 1, "hi", time.Unix(0, 0), "bob", 0, false)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT b.idblogs")).WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).WillReturnRows(blogRows)

	threadRows := sqlmock.NewRows([]string{"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername"}).
		AddRow(1, 1, 1, 1, 0, time.Unix(0, 0), false, "bob")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT th.idforumthread")).WithArgs(int32(2), int32(1), int32(2), int32(2), sql.NullInt32{Int32: 2, Valid: true}).WillReturnRows(threadRows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.idcomments")).
		WithArgs(int32(2), int32(2), int32(1), int32(2), int32(2), "blogs", sql.NullString{String: "entry", Valid: true}, sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "posterusername"}))

	rr := httptest.NewRecorder()
	CommentPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if !regexp.MustCompile(`Reply:`).MatchString(rr.Body.String()) {
		t.Fatalf("reply form should be shown")
	}
}
