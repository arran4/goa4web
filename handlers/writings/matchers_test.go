package writings

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
)

func TestRequireWritingAuthorArticleVar(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/writings/article/2/edit", nil)
	req = mux.SetURLVars(req, map[string]string{"article": "2"})

	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)

	cd := common.NewCoreData(req.Context(), q, common.WithSession(sess))
	cd.SetRoles([]string{"content writer"})
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	ctx = ctx
	req = req.WithContext(ctx)

	rows := sqlmock.NewRows([]string{
		"idwriting", "users_idusers", "forumthread_id", "language_idlanguage", "writing_category_id", "title", "published", "writing", "abstract", "private", "deleted_at", "last_index", "WriterId", "WriterUsername",
	}).AddRow(2, 1, 0, 1, 1, sql.NullString{}, sql.NullTime{}, sql.NullString{}, sql.NullString{}, sql.NullBool{}, sql.NullTime{}, sql.NullTime{}, 1, sql.NullString{})
	mock.ExpectQuery("SELECT w.idwriting").
		WithArgs(int32(1), int32(2), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(rows)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
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
