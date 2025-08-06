package linker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

func TestAdminLinkPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	rows := sqlmock.NewRows([]string{"idlinker", "language_idlanguage", "users_idusers", "linker_category_id", "forumthread_id", "title", "url", "description", "listed", "username", "title_2"}).
		AddRow(1, 1, 1, 1, 1, "t", "u", "d", time.Now(), "poster", "cat")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title FROM linker l JOIN users u ON l.users_idusers = u.idusers JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory WHERE l.idlinker = ?")).
		WithArgs(1).WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/admin/linker/links/link/1", nil)
	req = mux.SetURLVars(req, map[string]string{"link": "1"})
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminLinkPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}
