package blogs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestBloggerListPageSearchRedirect(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	rows := sqlmock.NewRows([]string{"username", "count"}).AddRow("arran4", 2)
	mock.ExpectQuery(regexp.QuoteMeta("WITH role_ids")).
		WithArgs(int32(0), "%arran4%", "%arran4%", int32(0), int32(0), nil, int32(16), int32(0)).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/blogs/bloggers?search=arran4", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.ShareSignKey = "secret"
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	BloggerListPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if loc := rr.Result().Header.Get("Location"); loc != "/blogs/blogger/arran4" {
		t.Fatalf("location=%s", loc)
	}
}
