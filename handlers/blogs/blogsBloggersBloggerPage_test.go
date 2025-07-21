package blogs

import (
	"context"
	"github.com/arran4/goa4web/core/consts"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
)

func TestBloggersBloggerPage(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := db.New(sqldb)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("bob", 2)
	mock.ExpectQuery(regexp.QuoteMeta("WITH RECURSIVE role_ids")).
		WithArgs(int32(0), int32(0), int32(0), nil, int32(1000), int32(0)).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/blogs/bloggers/blogger", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	BloggersBloggerPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != 200 {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
