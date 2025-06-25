package goa4web

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
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestAdminEmailTemplateTestAction_NoProvider(t *testing.T) {
	os.Unsetenv("EMAIL_PROVIDER")
	runtimeconfig.AppRuntimeConfig.EmailProvider = ""

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{UserID: 1})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminEmailTemplateTestActionPage(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status=%d", rr.Code)
	}
	want := url.QueryEscape(errMailNotConfigured)
	if loc := rr.Header().Get("Location"); !strings.Contains(loc, want) {
		t.Fatalf("location=%q", loc)
	}
}

func TestAdminEmailTemplateTestAction_WithProvider(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "log")
	runtimeconfig.AppRuntimeConfig.EmailProvider = "log"
	defer os.Unsetenv("EMAIL_PROVIDER")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	rows := sqlmock.NewRows([]string{"idusers", "email", "passwd", "passwd_algorithm", "username"}).
		AddRow(1, "u@example.com", "", "", "u")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idusers, email, passwd, passwd_algorithm, username\nFROM users\nWHERE idusers = ?")).
		WithArgs(int32(1)).WillReturnRows(rows)

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), ContextValues("coreData"), &CoreData{UserID: 1})
	ctx = context.WithValue(ctx, ContextValues("queries"), q)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	adminEmailTemplateTestActionPage(rr, req)

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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
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
	os.Setenv("ADMIN_EMAILS", "a@test.com,b@test.com")
	defer os.Unsetenv("ADMIN_EMAILS")
	rec := &recordAdminMail{}
	notifyAdmins(context.Background(), rec, nil, "page")
	if len(rec.to) != 2 {
		t.Fatalf("expected 2 mails, got %d", len(rec.to))
	}
}

func TestNotifyAdminsDisabled(t *testing.T) {
	os.Setenv("ADMIN_EMAILS", "a@test.com")
	os.Setenv("ADMIN_NOTIFY", "false")
	defer os.Unsetenv("ADMIN_EMAILS")
	defer os.Unsetenv("ADMIN_NOTIFY")
	rec := &recordAdminMail{}
	notifyAdmins(context.Background(), rec, nil, "page")
	if len(rec.to) != 0 {
		t.Fatalf("expected 0 mails, got %d", len(rec.to))
	}
}
