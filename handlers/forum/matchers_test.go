package forum

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func TestGetThreadAndTopicTrue(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername", "seelevel", "level",
		}).AddRow(2, 0, 0, 1, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}))

	mock.ExpectQuery("SELECT t.idforumtopic").
		WithArgs(int32(0), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "lastposter", "forumcategory_idforumcategory", "title", "description", "threads", "comments", "lastaddition", "LastPosterUsername", "seelevel", "level",
		}).AddRow(1, 0, 0, sql.NullString{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}))

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	ctx := context.WithValue(req.Context(), common.KeyQueries, q)
	req = req.WithContext(ctx)

	if !GetThreadAndTopic()(req, &mux.RouteMatch{}) {
		t.Errorf("expected match")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetThreadAndTopicFalse(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumthread", "firstpost", "lastposter", "forumtopic_idforumtopic", "comments", "lastaddition", "locked", "LastPosterUsername", "seelevel", "level",
		}).AddRow(2, 0, 0, 3, sql.NullInt32{}, sql.NullTime{}, sql.NullBool{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}))

	mock.ExpectQuery("SELECT t.idforumtopic").
		WithArgs(int32(0), int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{
			"idforumtopic", "lastposter", "forumcategory_idforumcategory", "title", "description", "threads", "comments", "lastaddition", "LastPosterUsername", "seelevel", "level",
		}).AddRow(3, 0, 0, sql.NullString{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}, sql.NullInt32{}, sql.NullInt32{}))

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	ctx := context.WithValue(req.Context(), common.KeyQueries, q)
	req = req.WithContext(ctx)

	if GetThreadAndTopic()(req, &mux.RouteMatch{}) {
		t.Errorf("expected no match")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetThreadAndTopicError(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	q := db.New(sqldb)

	mock.ExpectQuery("SELECT th.idforumthread").
		WithArgs(int32(0), int32(2)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/forum/topic/1/thread/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "2"})
	ctx := context.WithValue(req.Context(), common.KeyQueries, q)
	req = req.WithContext(ctx)

	if GetThreadAndTopic()(req, &mux.RouteMatch{}) {
		t.Errorf("expected no match on error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
