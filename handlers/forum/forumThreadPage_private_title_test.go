package forum

import (
	"context"
	"database/sql"
	"net/http/httptest"
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

func TestThreadPagePrivateSetsTitle(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	mock.ExpectQuery("SELECT th.idforumthread").
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername", "firstpostuserid",
		}).AddRow(int32(1), int32(1), int32(1), int32(1), sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}, sql.NullInt32{}))

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "LastPosterUsername",
		}).AddRow(int32(1), int32(1), int32(1), int32(1), sql.NullString{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, "private", sql.NullString{}))

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "username"}).AddRow(int32(2), sql.NullString{String: "Bob", Valid: true}))

	origStore := core.Store
	origName := core.SessionName
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	defer func() {
		core.Store = origStore
		core.SessionName = origName
	}()

	req := httptest.NewRequest("GET", "/private/topic/1/thread/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	// Inject an invalid session to force GetSessionOrFail to fail before rendering.
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), "bad")

	q := db.New(dbconn)
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, q, cfg)
	cd.SetCurrentSection("privateforum")
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ThreadPageWithBasePath(rr, req, "/private")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if cd.PageTitle == "" {
		t.Fatalf("page title not set")
	}
}
