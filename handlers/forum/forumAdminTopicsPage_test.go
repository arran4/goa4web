package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminTopicsPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(`).WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler"}).
		AddRow(1, 0, 0, 0, "t", "d", 0, 0, time.Now(), "")
	mock.ExpectQuery("SELECT t.idforumtopic").WillReturnRows(rows)

	categoryRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT f.idforumcategory, f.forumcategory_idforumcategory, f.language_id, f.title, f.description FROM forumcategory f")).WithArgs(int32(0), int32(0)).WillReturnRows(categoryRows)

	grantsRows := sqlmock.NewRows([]string{"section", "role_id", "role_name", "user_id", "username"}).
		AddRow("forum", sql.NullInt32{}, sql.NullString{}, sql.NullInt32{}, sql.NullString{})
	mock.ExpectQuery("SELECT g.section").WithArgs(sqlmock.AnyArg()).WillReturnRows(grantsRows)

	origStore := core.Store
	origName := core.SessionName
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	defer func() {
		core.Store = origStore
		core.SessionName = origName
	}()

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	r := httptest.NewRequest("GET", "/admin/forum/topics", nil)
	sess, _ := core.Store.New(r, core.SessionName)
	ctx := context.WithValue(r.Context(), core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	AdminTopicsPage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
