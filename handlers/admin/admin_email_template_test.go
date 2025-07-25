package admin

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	logProv "github.com/arran4/goa4web/internal/email/log"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func newEmailReg() *email.Registry {
	r := email.NewRegistry()
	logProv.Register(r)
	return r
}

func TestAdminEmailTemplateTestAction_NoProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = ""

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	reg := newEmailReg()
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithEmailProvider(reg.ProviderFromConfig(cfg)))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(testTemplateTask)(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAdminEmailTemplateTestAction_WithProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = "log"

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	rows := sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "u@example.com", "u")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username FROM users u LEFT JOIN user_emails ue ON ue.id = ( SELECT id FROM user_emails ue2 WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1 ) WHERE u.idusers = ?")).
		WithArgs(int32(1)).WillReturnRows(rows)
	mock.ExpectQuery("SELECT body FROM template_overrides WHERE name = ?").
		WithArgs("updateEmail.gotxt").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(sql.NullInt32{Int32: 1, Valid: true}, sqlmock.AnyArg(), false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest("POST", "/admin/email/template", nil)
	q := db.New(sqldb)
	reg := newEmailReg()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithEmailProvider(reg.ProviderFromConfig(cfg)))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(testTemplateTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
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
	rows := sqlmock.NewRows([]string{"id", "to_user_id", "body", "error_count", "created_at", "direct_email"}).AddRow(1, 2, "b", 0, time.Now(), false)
	mock.ExpectQuery("SELECT id, to_user_id, body, error_count, created_at, direct_email FROM pending_emails WHERE sent_at IS NULL ORDER BY id").WillReturnRows(rows)
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
	cfg := config.NewRuntimeConfig()
	cfg.AdminEmails = "a@test.com,b@test.com"
	cfg.AdminNotify = true
	cfg.EmailEnabled = true
	cfg.EmailFrom = "from@example.com"

	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	var q *db.Queries
	emails := []string{"a@test.com", "b@test.com"}
	for _, e := range emails {
		mock.ExpectQuery("UserByEmail").
			WithArgs(sql.NullString{String: e, Valid: true}).
			WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, e, "u"))
		mock.ExpectExec("INSERT INTO pending_emails").
			WithArgs(sql.NullInt32{Int32: 1, Valid: true}, sqlmock.AnyArg(), false).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	os.Setenv(config.EnvAdminEmails, "a@test.com,b@test.com")
	defer os.Unsetenv(config.EnvAdminEmails)
	cfg = config.NewRuntimeConfig()
	origEmails := cfg.AdminEmails
	cfg.AdminEmails = "a@test.com,b@test.com"
	defer func() { cfg.AdminEmails = origEmails }()
	sqldb, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqldb.Close()
	q = db.New(sqldb)
	rows := sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test.com", "a")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username FROM users u JOIN user_emails ue ON ue.user_id = u.idusers WHERE ue.email = ? LIMIT 1")).WithArgs("a@test.com").WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(sql.NullInt32{Int32: 1, Valid: true}, sqlmock.AnyArg(), false).WillReturnResult(sqlmock.NewResult(1, 1))
	rows2 := sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(2, "b@test.com", "b")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username FROM users u JOIN user_emails ue ON ue.user_id = u.idusers WHERE ue.email = ? LIMIT 1")).WithArgs("b@test.com").WillReturnRows(rows2)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(sql.NullInt32{Int32: 2, Valid: true}, sqlmock.AnyArg(), false).WillReturnResult(sqlmock.NewResult(1, 1))

	rec := &recordAdminMail{}
	n := notif.New(notif.WithQueries(q), notif.WithEmailProvider(rec), notif.WithConfig(cfg))
	n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})
	if len(rec.to) != 0 {
		t.Fatalf("expected 0 direct mails, got %d", len(rec.to))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestNotifyAdminsDisabled(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.AdminEmails = "a@test.com"
	cfg.AdminNotify = false
	cfg.EmailEnabled = true
	cfg.AdminEmails = "a@test.com"
	os.Setenv(config.EnvAdminNotify, "false")
	cfg.AdminEmails = "a@test.com"
	defer os.Unsetenv(config.EnvAdminEmails)
	defer os.Unsetenv(config.EnvAdminNotify)
	cfg.AdminEmails = "a@test.com"
	rec := &recordAdminMail{}
	n := notif.New(notif.WithEmailProvider(rec), notif.WithConfig(cfg))
	n.NotifyAdmins(context.Background(), &notif.EmailTemplates{}, notif.EmailData{})
	if len(rec.to) != 0 {
		t.Fatalf("expected 0 mails, got %d", len(rec.to))
	}
}
