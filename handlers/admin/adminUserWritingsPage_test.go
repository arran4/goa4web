package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminUserWritingsPage(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	userRows := sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).
		AddRow(1, "u@test", "user", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at FROM users u LEFT JOIN user_emails ue ON ue.id = ( SELECT id FROM user_emails ue2 WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1 ) WHERE u.idusers = ?")).
		WithArgs(int32(1)).
		WillReturnRows(userRows)

	writingRows := sqlmock.NewRows([]string{"idwriting", "users_idusers", "forumthread_id", "language_id", "writing_category_id", "title", "published", "timezone", "writing", "abstract", "private", "deleted_at", "last_index", "username", "comments"}).
		AddRow(1, 1, 0, 0, 2, "Title", time.Now(), time.Local.String(), "", "", false, nil, nil, "user", 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT w.idwriting, w.users_idusers, w.forumthread_id, w.language_id, w.writing_category_id, w.title, w.published, w.timezone, w.writing, w.abstract, w.private, w.deleted_at, w.last_index, u.username,\n    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id IS NOT NULL) AS Comments\nFROM writing w\nLEFT JOIN users u ON w.users_idusers = u.idusers\nWHERE w.users_idusers = ?\nORDER BY w.published DESC")).
		WithArgs(int32(1)).
		WillReturnRows(writingRows)

	req := httptest.NewRequest("GET", "/admin/user/1/writings", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	cd.SetCurrentProfileUserID(1)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	adminUserWritingsPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `<a href="/admin/writings/article/1">1</a>`) {
		t.Fatalf("missing admin link: %s", body)
	}
	if !strings.Contains(body, "Title") {
		t.Fatalf("missing title: %s", body)
	}
}
