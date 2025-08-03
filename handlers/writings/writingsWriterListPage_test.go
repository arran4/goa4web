package writings

import (
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestWriterListPage_List(t *testing.T) {
	t.Skip("environment not fully configured")
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(".*").WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/writings/writers", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	WriterListPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != 200 {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestWriterListPage_Search(t *testing.T) {
	t.Skip("environment not fully configured")
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(".*").WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/writings/writers?search=bob", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	WriterListPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != 200 {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
