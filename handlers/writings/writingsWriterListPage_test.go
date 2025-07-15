package writings

import (
	"context"
	"net/http/httptest"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func TestWriterListPage(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()

	queries := db.New(sqldb)
	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.username, COUNT(w.idwriting) AS count")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/writings/writers", nil)
	ctx := context.WithValue(req.Context(), hcommon.KeyQueries, queries)
	ctx = context.WithValue(ctx, hcommon.KeyCoreData, &corecommon.CoreData{})
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
