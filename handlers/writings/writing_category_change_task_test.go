package writings

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestWritingCategoryChangeTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)

	mock.ExpectExec("UPDATE writing_category").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), int32(0), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	form := url.Values{"name": {"A"}, "desc": {"B"}, "pcid": {"0"}, "cid": {"1"}}
	req := httptest.NewRequest("POST", "/admin/writings/categories", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if v := writingCategoryChangeTask.Action(nil, req); v != nil {
		t.Fatalf("action returned %v", v)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWritingCategoryWouldLoop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)

	rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "").
		AddRow(2, 1, "b", "")
	mock.ExpectQuery("SELECT wc.idwritingcategory").WillReturnRows(rows)

	loop, err := writingCategoryWouldLoop(context.Background(), queries, 1, 2)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWritingCategoryWouldLoopSelfRef(t *testing.T) {
	loop, err := writingCategoryWouldLoop(context.Background(), nil, 3, 3)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
}

func TestWritingCategoryWouldLoopHeadToTail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)

	rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "").
		AddRow(2, 1, "b", "").
		AddRow(3, 2, "c", "").
		AddRow(4, 3, "d", "")
	mock.ExpectQuery("SELECT wc.idwritingcategory").WillReturnRows(rows)

	loop, err := writingCategoryWouldLoop(context.Background(), queries, 1, 4)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWritingCategoryWouldLoopAfterNode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)

	rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 0, "a", "").
		AddRow(2, 3, "b", "").
		AddRow(3, 2, "c", "")
	mock.ExpectQuery("SELECT wc.idwritingcategory").WillReturnRows(rows)

	loop, err := writingCategoryWouldLoop(context.Background(), queries, 1, 2)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestWritingCategoryChangeTaskLoop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	queries := dbpkg.New(db)

	rows := sqlmock.NewRows([]string{"idwritingcategory", "writing_category_id", "title", "description"}).
		AddRow(1, 2, "a", "").
		AddRow(2, 1, "b", "")
	mock.ExpectQuery("SELECT wc.idwritingcategory").WillReturnRows(rows)

	form := url.Values{"name": {"A"}, "desc": {"B"}, "pcid": {"2"}, "cid": {"1"}}
	req := httptest.NewRequest("POST", "/admin/writings/categories", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if v := writingCategoryChangeTask.Action(nil, req); v == nil {
		t.Fatalf("expected error")
	} else if ue, ok := v.(common.UserError); !ok {
		t.Fatalf("expected user error got %T", v)
	} else if ue.UserErrorMessage() != "invalid parent category" {
		t.Fatalf("unexpected error message %q", ue.UserErrorMessage())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
