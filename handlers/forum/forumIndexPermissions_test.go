package forum

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	corecommon "github.com/arran4/goa4web/core/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func TestCustomForumIndexWriteReply(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := dbpkg.New(sqldb)
	ctx := context.WithValue(req.Context(), corecorecommon.KeyQueries, q)
	cd := corecommon.NewCoreData(ctx, q)

	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "reply", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Write Reply") {
		t.Errorf("expected write reply item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexWriteReplyDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2/thread/3", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "thread": "3"})

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := dbpkg.New(sqldb)
	ctx := context.WithValue(req.Context(), corecorecommon.KeyQueries, q)
	cd := corecommon.NewCoreData(ctx, q)

	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "reply", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	CustomForumIndex(cd, req.WithContext(ctx))
	if corecommon.ContainsItem(cd.CustomIndexItems, "Write Reply") {
		t.Errorf("unexpected write reply item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexCreateThread(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := dbpkg.New(sqldb)
	ctx := context.WithValue(req.Context(), corecorecommon.KeyQueries, q)
	cd := corecommon.NewCoreData(ctx, q)

	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	CustomForumIndex(cd, req.WithContext(ctx))
	if !corecommon.ContainsItem(cd.CustomIndexItems, "Create Thread") {
		t.Errorf("expected create thread item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCustomForumIndexCreateThreadDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/forum/topic/2", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "2", "category": "1"})

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := dbpkg.New(sqldb)
	ctx := context.WithValue(req.Context(), corecorecommon.KeyQueries, q)
	cd := corecommon.NewCoreData(ctx, q)

	mock.ExpectQuery("SELECT 1 FROM grants").
		WithArgs(sqlmock.AnyArg(), "forum", sqlmock.AnyArg(), "post", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	CustomForumIndex(cd, req.WithContext(ctx))
	if corecommon.ContainsItem(cd.CustomIndexItems, "Create Thread") {
		t.Errorf("unexpected create thread item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
