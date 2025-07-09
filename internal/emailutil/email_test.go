package emailutil_test

import (
	"context"
	"flag"
	"fmt"
	"net/mail"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	mockdlq "github.com/arran4/goa4web/internal/dlq/mock"
	"github.com/arran4/goa4web/internal/email"
	jmapProv "github.com/arran4/goa4web/internal/email/jmap"
	localProv "github.com/arran4/goa4web/internal/email/local"
	logProv "github.com/arran4/goa4web/internal/email/log"
	mockemail "github.com/arran4/goa4web/internal/email/mock"
	sesProv "github.com/arran4/goa4web/internal/email/ses"
	smtpProv "github.com/arran4/goa4web/internal/email/smtp"
	"github.com/arran4/goa4web/internal/emailutil"
	"github.com/arran4/goa4web/runtimeconfig"
	"strings"
)

func init() {
	logProv.Register()
	smtpProv.Register()
	localProv.Register()
	jmapProv.Register()
	sesProv.Register()
}

func TestGetEmailProviderLog(t *testing.T) {
	cfg := runtimeconfig.RuntimeConfig{EmailProvider: "log"}
	p := email.ProviderFromConfig(cfg)
	if _, ok := p.(logProv.Provider); !ok {
		t.Errorf("expected LogProvider, got %#v", p)
	}
}

func TestGetEmailProviderUnknown(t *testing.T) {
	cfg := runtimeconfig.RuntimeConfig{EmailProvider: "unknown"}
	if p := email.ProviderFromConfig(cfg); p != nil {
		t.Errorf("expected nil for unknown provider, got %#v", p)
	}
}

func TestEmailConfigPrecedence(t *testing.T) {
	env := map[string]string{
		config.EnvEmailProvider: "ses",
		config.EnvSMTPHost:      "env",
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("email-provider", "smtp", "")
	fs.String("smtp-port", "25", "")
	vals := map[string]string{
		config.EnvEmailProvider: "log",
		config.EnvSMTPHost:      "file",
	}
	_ = fs.Parse([]string{"--email-provider=smtp", "--smtp-port=25"})
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals, func(k string) string { return env[k] })
	if cfg.EmailProvider != "smtp" || cfg.EmailSMTPHost != "file" || cfg.EmailSMTPPort != "25" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadEmailConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvEmailProvider: "log",
	}
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals, func(string) string { return "" })
	if cfg.EmailProvider != "log" {
		t.Fatalf("want log got %q", cfg.EmailProvider)
	}
}

func TestNotifyChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(int32(2), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	ctx := context.WithValue(context.Background(), common.KeyQueries, q)
	rec := &mockemail.Provider{}
	if err := emailutil.NotifyChange(ctx, rec, 2, "a@b.com", "http://host", "update", nil); err != nil {
		t.Fatalf("notify error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNotifyChangeErrors(t *testing.T) {
	rec := &mockemail.Provider{}
	if err := emailutil.NotifyChange(context.Background(), rec, 0, "", "p", "update", nil); err == nil {
		t.Fatal("expected error for empty email")
	}
}

func TestInsertPendingEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := dbpkg.New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(int32(1), "body").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := q.InsertPendingEmail(context.Background(), dbpkg.InsertPendingEmailParams{ToUserID: 1, Body: "body"}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestEmailQueueWorker(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"id", "to_user_id", "body", "error_count"}).AddRow(1, 2, "b", 0)
	mock.ExpectQuery("SELECT id, to_user_id").WillReturnRows(rows)
	mock.ExpectQuery("SELECT idusers, email, username FROM users WHERE idusers = ?").WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(2, "e", "bob"))
	mock.ExpectExec("UPDATE pending_emails SET sent_at").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	rec := &mockemail.Provider{}
	emailutil.ProcessPendingEmail(context.Background(), q, rec, nil)

	if len(rec.Messages) != 1 || rec.Messages[0].To.String() != "\"bob\" <e@>" {
		t.Fatalf("got %#v", rec.Messages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

type errProvider struct{}

func (errProvider) Send(context.Context, mail.Address, []byte) error {
	return fmt.Errorf("fail")
}

func TestProcessPendingEmailDLQ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	rows := sqlmock.NewRows([]string{"id", "to_user_id", "body", "error_count"}).AddRow(1, 2, "b", 4)
	mock.ExpectQuery("SELECT id, to_user_id").WillReturnRows(rows)
	mock.ExpectQuery("SELECT idusers, email, username FROM users WHERE idusers = ?").WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(2, "a@test", "a"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE pending_emails SET error_count = error_count + 1 WHERE id = ?")).WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT error_count FROM pending_emails WHERE id = ?").WithArgs(int32(1)).WillReturnRows(sqlmock.NewRows([]string{"error_count"}).AddRow(5))
	mock.ExpectExec("DELETE FROM pending_emails WHERE id = ?").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	p := errProvider{}
	dlqRec := &mockdlq.Provider{}
	emailutil.ProcessPendingEmail(context.Background(), q, p, dlqRec)

	if len(dlqRec.Records) != 1 {
		t.Fatalf("dlq records=%d", len(dlqRec.Records))
	}
	msg := dlqRec.Records[0].Message
	if !strings.Contains(msg, "b") || !strings.Contains(msg, "fail") {
		t.Fatalf("unexpected dlq message: %s", msg)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetEmailProviderSMTP(t *testing.T) {
	p := email.ProviderFromConfig(runtimeconfig.RuntimeConfig{
		EmailProvider:     "smtp",
		EmailSMTPHost:     "localhost",
		EmailSMTPPort:     "25",
		EmailFrom:         "from@example.com",
		EmailSMTPStartTLS: true,
	})
	s, ok := p.(smtpProv.Provider)
	if !ok {
		t.Fatalf("expected SMTPProvider, got %#v", p)
	}
	if s.Addr != "localhost:25" || s.From != "from@example.com" || !s.StartTLS {
		t.Errorf("unexpected provider values: %#v", s)
	}
}

func TestGetEmailProviderLocal(t *testing.T) {
	if _, ok := email.ProviderFromConfig(runtimeconfig.RuntimeConfig{EmailProvider: "local"}).(localProv.Provider); !ok {
		t.Fatalf("expected LocalProvider")
	}
}

func TestGetEmailProviderJMAP(t *testing.T) {
	p := email.ProviderFromConfig(runtimeconfig.RuntimeConfig{
		EmailProvider:     "jmap",
		EmailJMAPEndpoint: "http://example.com",
		EmailJMAPAccount:  "acct",
		EmailJMAPIdentity: "id",
	})
	j, ok := p.(jmapProv.Provider)
	if !ok {
		t.Fatalf("expected JMAPProvider, got %#v", p)
	}
	if j.Endpoint != "http://example.com" || j.AccountID != "acct" || j.Identity != "id" {
		t.Errorf("unexpected provider values: %#v", j)
	}
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	if p := email.ProviderFromConfig(runtimeconfig.RuntimeConfig{EmailProvider: "ses", EmailAWSRegion: "us-east-1"}); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}

func TestProviderRegistry(t *testing.T) {
	called := false
	email.RegisterProvider("testprov", func(cfg runtimeconfig.RuntimeConfig) email.Provider {
		called = true
		return logProv.Provider{}
	})
	p := email.ProviderFromConfig(runtimeconfig.RuntimeConfig{EmailProvider: "testprov"})
	if !called {
		t.Fatal("factory not called")
	}
	if _, ok := p.(logProv.Provider); !ok {
		t.Fatalf("expected LogProvider, got %#v", p)
	}
}
