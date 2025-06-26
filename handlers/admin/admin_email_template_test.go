package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/common"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestAdminEmailTemplateTestAction_NoProvider(t *testing.T) {
	os.Unsetenv(config.EnvEmailProvider)
	runtimeconfig.AppRuntimeConfig.EmailProvider = ""

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &CoreData{UserID: 1})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminEmailTemplateTestActionPage(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Code)
	}
	want := url.QueryEscape(userhandlers.ErrMailNotConfigured)
	if loc := rr.Header().Get("Location"); !strings.Contains(loc, want) {
		t.Fatalf("location=%q", loc)
	}
}

func TestAdminEmailTemplateTestAction_WithProvider(t *testing.T) {
	os.Setenv(config.EnvEmailProvider, "log")
	runtimeconfig.AppRuntimeConfig.EmailProvider = "log"
	defer os.Unsetenv(config.EnvEmailProvider)

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := db.New(sqldb)
	rows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "u@example.com", "", "", "u")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, email, passwd, passwd_algorithm, username\nFROM users\nWHERE idusers = ?")).
		WithArgs(int32(1)).WillReturnRows(rows)

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &CoreData{UserID: 1})
	ctx = context.WithValue(ctx, common.KeyQueries, q)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminEmailTemplateTestActionPage(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/admin/email/template" {
		t.Fatalf("location=%q", loc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestListUnsentPendingEmails(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := db.New(sqldb)
	rows := sqlmock.NewRows([]string{"id", "to_email", "subject", "body", "created_at"}).AddRow(1, "a@test", "s", "b", time.Now())
	mock.ExpectQuery("SELECT id, to_email, subject, body, created_at FROM pending_emails WHERE sent_at IS NULL ORDER BY id").WillReturnRows(rows)
	if _, err := q.ListUnsentPendingEmails(context.Background()); err != nil {
		t.Fatalf("list: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestRecentNotifications(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := db.New(sqldb)
	rows := sqlmock.NewRows([]string{"id", "users_idusers", "link", "message", "created_at", "read_at"}).AddRow(1, 1, "/l", "m", time.Now(), nil)
	mock.ExpectQuery("SELECT id, users_idusers, link, message, created_at, read_at FROM notifications ORDER BY id DESC LIMIT ?").WithArgs(int32(5)).WillReturnRows(rows)
	if _, err := q.RecentNotifications(context.Background(), 5); err != nil {
		t.Fatalf("recent: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

type recordAdminMail struct{ to []string }

func (r *recordAdminMail) Send(ctx context.Context, to, subject, body string) error {
	r.to = append(r.to, to)
	return nil
}

func TestNotifyAdminsEnv(t *testing.T) {
	os.Setenv(config.EnvAdminEmails, "a@test.com,b@test.com")
	defer os.Unsetenv(config.EnvAdminEmails)
	rec := &recordAdminMail{}
	notifyAdmins(context.Background(), rec, nil, "page")
	if len(rec.to) != 2 {
		t.Fatalf("expected 2 mails, got %d", len(rec.to))
	}
}

func TestNotifyAdminsDisabled(t *testing.T) {
	os.Setenv(config.EnvAdminEmails, "a@test.com")
	os.Setenv(config.EnvAdminNotify, "false")
	defer os.Unsetenv(config.EnvAdminEmails)
	defer os.Unsetenv(config.EnvAdminNotify)
	rec := &recordAdminMail{}
	notifyAdmins(context.Background(), rec, nil, "page")
	if len(rec.to) != 0 {
		t.Fatalf("expected 0 mails, got %d", len(rec.to))
	}
}
