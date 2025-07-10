package writings

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
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

	cd := corecommon.NewCoreData(req.Context(), q, corecommon.WithSession(sess))
	cd.SecurityLevel = "writer"
	ctx := context.WithValue(req.Context(), hcommon.KeyCoreData, cd)
	ctx = context.WithValue(ctx, hcommon.KeyQueries, q)
	req = req.WithContext(ctx)

	rows := sqlmock.NewRows([]string{
		"idwriting", "users_idusers", "forumthread_idforumthread", "language_idlanguage", "writing_category_id", "title", "published", "writing", "abstract", "private", "deleted_at", "WriterId", "WriterUsername",
	}).AddRow(2, 1, 0, 1, 1, sql.NullString{}, sql.NullTime{}, sql.NullString{}, sql.NullString{}, sql.NullBool{}, sql.NullTime{}, 1, sql.NullString{})
	mock.ExpectQuery("SELECT w.idwriting").
		WithArgs(int32(1), int32(2), int32(1)).
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
