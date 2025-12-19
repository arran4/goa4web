package common_test

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestTemplateFuncsFirstline(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	first := funcs["firstline"].(func(string) string)
	if got := first("a\nb\n"); got != "a" {
		t.Errorf("firstline=%q", got)
	}
}

func TestTemplateFuncsLeft(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	left := funcs["left"].(func(int, string) string)
	if got := left(3, "hello"); got != "hel" {
		t.Errorf("left short=%q", got)
	}
	if got := left(10, "hi"); got != "hi" {
		t.Errorf("left long=%q", got)
	}
}

func TestTemplateFuncsCSRFToken(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	if _, ok := funcs["csrfToken"]; !ok {
		t.Errorf("csrfToken func missing")
	}
	if _, ok := funcs["csrf"]; ok {
		t.Errorf("csrf func should not be present")
	}
}

func TestLatestNewsRespectsPermissions(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"writerName", "writerId", "idsitenews", "forumthread_id", "language_id",
		"users_idusers", "news", "occurred", "timezone", "comments",
	}).AddRow("w", 1, 1, 0, 1, 1, "a", now, time.Local.String(), 0).AddRow("w", 1, 2, 0, 1, 1, "b", now, time.Local.String(), 0)

	mock.ExpectQuery("SELECT u.username").WithArgs(int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows)

	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectQuery("SELECT 1 FROM grants g JOIN roles").WithArgs("user", "administrator").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	_ = req.WithContext(ctx)

	res, err := cd.LatestNews()
	if err != nil {
		t.Fatalf("LatestNews: %v", err)
	}
	if l := len(res); l != 1 {
		t.Fatalf("expected 1 news post, got %d", l)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAddmodeSkipsAdminLinks(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{AdminMode: true}
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	funcs := cd.Funcs(req)
	addmode := funcs["addmode"].(func(string) string)

	tests := []struct {
		in   string
		want string
	}{
		{"/admin", "/admin"},
		{"/admin/tools", "/admin/tools"},
		{"/admin/tools?flag=1", "/admin/tools?flag=1"},
		{"http://example.com/admin", "http://example.com/admin"},
		{"/administrator", "/administrator?mode=admin"},
		{"/user", "/user?mode=admin"},
		{"/user?id=1", "/user?id=1&mode=admin"},
	}

	for _, tt := range tests {
		if got := addmode(tt.in); got != tt.want {
			t.Errorf("addmode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
