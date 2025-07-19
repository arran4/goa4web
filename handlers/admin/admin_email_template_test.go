package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	db "github.com/arran4/goa4web/internal/db"
	logProv "github.com/arran4/goa4web/internal/email/log"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func init() { logProv.Register() }

func TestAdminEmailTemplateTestAction_NoProvider(t *testing.T) {
	config.AppRuntimeConfig.EmailProvider = ""

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &common.CoreData{UserID: 1})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	TestTemplateTask.Action(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	want := url.QueryEscape(userhandlers.ErrMailNotConfigured.Error())
	if req.URL.RawQuery != "error="+want {
		t.Fatalf("query=%q", req.URL.RawQuery)
	}
	if !strings.Contains(rr.Body.String(), "<a href=") {
		t.Fatalf("body=%q", rr.Body.String())
	}
}

func TestAdminEmailTemplateTestAction_WithProvider(t *testing.T) {
	config.AppRuntimeConfig.EmailProvider = "log"

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := db.New(sqldb)
	rows := sqlmock.NewRows([]string{"idusers", "email", "username"}).
		AddRow(1, "u@example.com", "u")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username FROM users u LEFT JOIN user_emails ue ON ue.id = ( SELECT id FROM user_emails ue2 WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1 ) WHERE u.idusers = ?")).
		WithArgs(int32(1)).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO pending_emails").
		WithArgs(int32(1), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	ctx := context.WithValue(req.Context(), common.KeyCoreData, &common.CoreData{UserID: 1})
	ctx = context.WithValue(ctx, common.KeyQueries, q)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	TestTemplateTask.Action(rr, req)

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
	rows := sqlmock.NewRows([]string{"id", "to_user_id", "body", "error_count", "created_at"}).AddRow(1, 2, "b", 0, time.Now())
	mock.ExpectQuery("SELECT id, to_user_id, body, error_count, created_at FROM pending_emails WHERE sent_at IS NULL ORDER BY id").WillReturnRows(rows)
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

func (r *recordAdminMail) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	r.to = append(r.to, to.Address)
	return nil
}

func TestNotifyAdminsEnv(t *testing.T) {
	cfgOrig := config.AppRuntimeConfig
	config.AppRuntimeConfig.AdminEmails = "a@test.com,b@test.com"
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	t.Cleanup(func() { config.AppRuntimeConfig = cfgOrig })

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q := db.New(sqldb)

	emails := []string{"a@test.com", "b@test.com"}
	for _, e := range emails {
		mock.ExpectQuery("UserByEmail").
			WithArgs(sql.NullString{String: e, Valid: true}).
			WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, e, "u"))
		mock.ExpectExec("INSERT INTO pending_emails").
			WithArgs(int32(1), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	os.Setenv(config.EnvAdminEmails, "a@test.com,b@test.com")
	defer os.Unsetenv(config.EnvAdminEmails)

	rec := &recordAdminMail{}
	n := notif.New(q, rec)
	n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})

	if len(rec.to) != 0 {
		t.Fatalf("expected no direct mail got %d", len(rec.to))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestNotifyAdminsDisabled(t *testing.T) {
	cfgOrig := config.AppRuntimeConfig
	config.AppRuntimeConfig.AdminEmails = "a@test.com"
	config.AppRuntimeConfig.AdminNotify = false
	config.AppRuntimeConfig.EmailEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = cfgOrig })
	orig := config.AppRuntimeConfig
	defer func() { config.AppRuntimeConfig = orig }()
	config.AppRuntimeConfig.AdminEmails = "a@test.com"
	os.Setenv(config.EnvAdminNotify, "false")
	config.AppRuntimeConfig.AdminEmails = "a@test.com"
	defer os.Unsetenv(config.EnvAdminEmails)
	defer os.Unsetenv(config.EnvAdminNotify)
	origEmails := config.AppRuntimeConfig.AdminEmails
	config.AppRuntimeConfig.AdminEmails = "a@test.com"
	defer func() { config.AppRuntimeConfig.AdminEmails = origEmails }()
	rec := &recordAdminMail{}
	n := notif.New(nil, rec)
	n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})
	if len(rec.to) != 0 {
		t.Fatalf("expected 0 mails, got %d", len(rec.to))
	}
}
