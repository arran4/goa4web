package linker

import (
	"context"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminLinkViewPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	req := httptest.NewRequest("GET", "/admin/linker/links/link/1", nil)
	req = mux.SetURLVars(req, map[string]string{"link": "1"})
	w := httptest.NewRecorder()

	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT l.idlinker")).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idlinker", "language_id", "users_idusers", "linker_category_id", "forumthread_id", "title", "url", "description", "listed", "username", "title_2"}).
			AddRow(1, 1, 2, 1, 0, "t", "http://u", "d", time.Unix(0, 0), "bob", "cat"))

	adminLinkViewPage(w, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
