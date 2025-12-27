package writings

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

var grantQuery = regexp.QuoteMeta(`WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = ?
)
SELECT 1 FROM grants g
WHERE g.section = ?
  AND (g.item = ? OR g.item IS NULL)
  AND g.action = ?
  AND g.active = 1
  AND (g.item_id = ? OR g.item_id IS NULL)
  AND (g.user_id = ? OR g.user_id IS NULL)
  AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
LIMIT 1
`)

func TestRequireWritingAuthorWritingVar(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/writings/article/2/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"writing": "2"})

	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)

	cd := common.NewCoreData(
		req.Context(),
		q,
		config.NewRuntimeConfig(),
		common.WithSession(sess),
		common.WithUserRoles([]string{"content writer"}),
	)
	cd.LoadSelectionsFromRequest(req)
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rows := sqlmock.NewRows([]string{
		"idwriting", "users_idusers", "forumthread_id", "language_id", "writing_category_id", "title", "published", "timezone", "writing", "abstract", "private", "deleted_at", "last_index", "WriterId", "WriterUsername",
	}).AddRow(2, 1, 0, 1, 1, sql.NullString{}, sql.NullTime{}, sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullBool{}, sql.NullTime{}, sql.NullTime{}, 1, sql.NullString{})
	mock.ExpectQuery("SELECT w.idwriting").
		WithArgs(int32(1), int32(2), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(rows)
	mock.ExpectQuery(grantQuery).
		WithArgs(int32(1), "writing", sql.NullString{String: "article", Valid: true}, "edit", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if cd.CurrentWritingLoaded() == nil {
			t.Errorf("writing not cached")
		}
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	RequireWritingAuthor(handler).ServeHTTP(rr, req)

	if !called {
		t.Errorf("expected handler call")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestMatchCanEditWritingArticleUsesGrant(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	req := httptest.NewRequest("GET", "/writings/article/2/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"writing": "2"})

	cd := common.NewCoreData(req.Context(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 7
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mock.ExpectQuery(grantQuery).
		WithArgs(int32(7), "writing", sql.NullString{String: "article", Valid: true}, "edit", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 7, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	if !MatchCanEditWritingArticle(req, &mux.RouteMatch{}) {
		t.Fatalf("expected match to allow edit")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestMatchCanPostWritingDeniesWithoutGrant(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	req := httptest.NewRequest("GET", "/writings/category/3/add", nil)
	req = mux.SetURLVars(req, map[string]string{"category": "3"})

	cd := common.NewCoreData(req.Context(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 4
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mock.ExpectQuery(grantQuery).
		WithArgs(int32(4), "writing", sql.NullString{String: "category", Valid: true}, "post", sql.NullInt32{Int32: 3, Valid: true}, sql.NullInt32{Int32: 4, Valid: true}).
		WillReturnError(sql.ErrNoRows)

	if MatchCanPostWriting(req, &mux.RouteMatch{}) {
		t.Fatalf("expected match to deny posting")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
