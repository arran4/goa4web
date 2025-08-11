package blogs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminBlogCommentsPage_UsesURLParam(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	blogID := 9
	rows := sqlmock.NewRows([]string{"idblogs", "forumthread_id", "users_idusers", "language_idlanguage", "blog", "written", "timezone", "username", "comments", "is_owner"}).
		AddRow(blogID, sql.NullInt32{Int32: 1, Valid: true}, 1, 1, "body", time.Now(), time.Local.String(), "user", 0, true)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/admin/blogs/blog/"+strconv.Itoa(blogID)+"/comments", nil)
	req = mux.SetURLVars(req, map[string]string{"blog": strconv.Itoa(blogID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()

	AdminBlogCommentsPage(rr, req.WithContext(ctx))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
