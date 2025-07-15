package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func TestRequireThreadAndTopicTrue(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername",
		}).AddRow(2, 0, 0, 1, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}))

	mock.ExpectQuery("SELECT t.idforumtopic").
		WithArgs(int32(0), int32(1), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "lastposter", "forumcategory_idforumcategory", "title", "description", "threads", "comments", "lastaddition", "LastPosterUsername",
		}).AddRow(1, 0, 0, sql.NullString{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}))

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	ctx := context.WithValue(req.Context(), common.KeyQueries, q)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Context().Value(common.KeyThread) == nil || r.Context().Value(common.KeyTopic) == nil {
			t.Errorf("context values missing")
		}
		w.WriteHeader(http.StatusOK)
	})

	RequireThreadAndTopic(handler).ServeHTTP(rr, req)
	if !called {
		t.Errorf("expected handler call")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("unexpected status %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRequireThreadAndTopicFalse(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername",
		}).AddRow(2, 0, 0, 3, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}))

	mock.ExpectQuery("SELECT t.idforumtopic").
		WithArgs(int32(0), int32(3), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "lastposter", "forumcategory_idforumcategory", "title", "description", "threads", "comments", "lastaddition", "LastPosterUsername",
		}).AddRow(3, 0, 0, sql.NullString{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}))

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	ctx := context.WithValue(req.Context(), common.KeyQueries, q)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	RequireThreadAndTopic(handler).ServeHTTP(rr, req)
	if called {
		t.Errorf("expected handler not called")
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRequireThreadAndTopicError(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	ctx := context.WithValue(req.Context(), common.KeyQueries, q)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	RequireThreadAndTopic(handler).ServeHTTP(rr, req)
	if called {
		t.Errorf("expected handler not called")
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
