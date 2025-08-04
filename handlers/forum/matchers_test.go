package forum

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestRequireThreadAndTopicTrue(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2), int32(0), int32(0), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername",
		}).AddRow(2, 0, 0, 1, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}))

	mock.ExpectQuery("SELECT t.idforumtopic").
		WithArgs(int32(0), int32(1), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "lastposter", "forumcategory_idforumcategory", "title", "description", "threads", "comments", "lastaddition", "LastPosterUsername",
		}).AddRow(1, 0, 0, sql.NullString{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}))

		// handler will trigger another load
	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2), int32(0), int32(0), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername",
		}).AddRow(2, 0, 0, 1, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}))

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, common.WithConfig(config.NewRuntimeConfig()))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, err := cd.CurrentThread(); err != nil {
			t.Errorf("CurrentThread: %v", err)
		}
		if _, err := cd.CurrentTopic(); err != nil {
			t.Errorf("CurrentTopic: %v", err)
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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2), int32(0), int32(0), sql.NullInt32{Int32: 0, Valid: false}).
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
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, common.WithConfig(config.NewRuntimeConfig()))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
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
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2), int32(0), int32(0), sql.NullInt32{Int32: 0, Valid: false}).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, common.WithConfig(config.NewRuntimeConfig()))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
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
