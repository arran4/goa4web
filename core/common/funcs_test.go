package common_test

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	common "github.com/arran4/goa4web/core/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestTemplateFuncsFirstline(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	first := funcs["firstline"].(func(string) string)
	if got := first("a\nb\n"); got != "a" {
		t.Errorf("firstline=%q", got)
	}
}

func TestTemplateFuncsLeft(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	left := funcs["left"].(func(int, string) string)
	if got := left(3, "hello"); got != "hel" {
		t.Errorf("left short=%q", got)
	}
	if got := left(10, "hi"); got != "hi" {
		t.Errorf("left long=%q", got)
	}
}

func TestLatestNewsRespectsPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"writerName", "writerId", "idsitenews", "forumthread_id", "language_idlanguage",
		"users_idusers", "news", "occurred", "last_index", "comments",
	}).AddRow("w", 1, 1, 0, 1, 1, "a", now, now, 0).AddRow("w", 1, 2, 0, 1, 1, "b", now, now, 0)

	mock.ExpectQuery("SELECT u.username").WithArgs(int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows)

	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT id, site_news_id, active, created_at").WithArgs(int32(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), common.ContextValues("queries"), queries)
	cd := common.NewCoreData(ctx, queries)
	cd.UserID = 1
	cd.SetRoles([]string{"user"})
	ctx = context.WithValue(ctx, common.ContextValues("coreData"), cd)
	req = req.WithContext(ctx)

	funcs := cd.Funcs(req)
	latestFn := funcs["LatestNews"].(func() (any, error))
	res, err := latestFn()
	if err != nil {
		t.Fatalf("LatestNews: %v", err)
	}
	if l := reflect.ValueOf(res).Len(); l != 1 {
		t.Fatalf("expected 1 news post, got %d", l)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
