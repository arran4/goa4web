package forum

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestThreadPageQuotePrefillsReply(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	mock.ExpectQuery(".*").
		WithArgs(int32(1), int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername",
		}).AddRow(int32(1), int32(1), int32(1), int32(1), sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}))

	mock.ExpectQuery(".*").
		WithArgs(int32(1), int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_idlanguage", "title", "description", "threads", "comments", "lastaddition", "handler", "LastPosterUsername",
		}).AddRow(int32(1), int32(1), int32(1), int32(1), sql.NullString{String: "topic", Valid: true}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, "normal", sql.NullString{}))

	mock.ExpectQuery(".*").
		WithArgs(int32(1), int32(1), int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "last_index", "posterusername", "is_owner",
		}))

	mock.ExpectQuery(".*").
		WithArgs(int32(1), int32(1), int32(2), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idcomments", "forumthread_id", "users_idusers", "language_idlanguage", "written", "text", "deleted_at", "last_index", "username", "is_owner",
		}).AddRow(int32(2), int32(1), int32(1), int32(1), sql.NullTime{}, sql.NullString{String: "hi", Valid: true}, sql.NullTime{}, sql.NullTime{}, sql.NullString{String: "alice", Valid: true}, false))

	mock.ExpectQuery(".*").
		WithArgs(int32(1), "forum", sql.NullString{String: "topic", Valid: true}, "reply", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/1?quote=2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	w := httptest.NewRecorder()
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	q := db.New(dbconn)
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(sess), common.WithUserRoles([]string{"user"}))
	cd.SetCurrentSection("forum")
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ThreadPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

	if !strings.Contains(rr.Body.String(), "[quoteof") {
		t.Fatalf("reply not quoted: %s", rr.Body.String())
	}
}
