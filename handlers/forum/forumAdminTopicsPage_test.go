package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminTopicsPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(`).WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler"}).
		AddRow(1, 0, 0, 0, "t", "d", 0, 0, time.Now(), "")
	mock.ExpectQuery("SELECT t.idforumtopic").WillReturnRows(rows)

	categoryRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery("SELECT f\\.\\* FROM forumcategory").WillReturnRows(categoryRows)

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	r := httptest.NewRequest("GET", "/admin/forum/topics", nil)
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
	w := httptest.NewRecorder()

	AdminTopicsPage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
