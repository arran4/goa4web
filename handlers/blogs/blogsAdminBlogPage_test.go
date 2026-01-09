package blogs

import (
	"context"
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

func TestAdminBlogPage_UsesURLParam(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	blogID := 7
	rows := sqlmock.NewRows([]string{"idblogs", "forumthread_id", "users_idusers", "language_id", "blog", "written", "timezone", "username", "comments", "is_owner", "title"}).
		AddRow(blogID, nil, 1, 1, "body", time.Now(), time.Local.String(), "user", 0, true, "body")
	mock.ExpectQuery("SELECT b.idblogs").WillReturnRows(rows)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "private_labels", "public_profile_allowed_at"}))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "role_id", "section", "item", "rule_type", "item_id", "item_rule", "action", "extra", "active"}))

	req := httptest.NewRequest("GET", "/admin/blogs/blog/"+strconv.Itoa(blogID), nil)
	req = mux.SetURLVars(req, map[string]string{"blog": strconv.Itoa(blogID)})
	cfg := config.NewRuntimeConfig()
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()

	AdminBlogPage(rr, req.WithContext(ctx))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
