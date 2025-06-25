package goa4web

import (
	"context"
	"flag"
	"os"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestGetEmailProviderLog(t *testing.T) {
	cfg := runtimeconfig.RuntimeConfig{EmailProvider: "log"}
	if p := providerFromConfig(cfg); reflect.TypeOf(p) != reflect.TypeOf(email.LogProvider{}) {
		t.Errorf("expected LogProvider, got %#v", p)
	}
}

func TestGetEmailProviderUnknown(t *testing.T) {
	cfg := runtimeconfig.RuntimeConfig{EmailProvider: "unknown"}
	if p := providerFromConfig(cfg); p != nil {
		t.Errorf("expected nil for unknown provider, got %#v", p)
	}
}

func TestEmailConfigPrecedence(t *testing.T) {
	os.Setenv(config.EnvEmailProvider, "ses")
	os.Setenv(config.EnvSMTPHost, "env")
	defer os.Unsetenv(config.EnvEmailProvider)
	defer os.Unsetenv(config.EnvSMTPHost)

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("email-provider", "smtp", "")
	fs.String("smtp-port", "25", "")
	vals := map[string]string{
		config.EnvEmailProvider: "log",
		config.EnvSMTPHost:      "file",
	}
	_ = fs.Parse([]string{"--email-provider=smtp", "--smtp-port=25"})
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals)
	if cfg.EmailProvider != "smtp" || cfg.EmailSMTPHost != "file" || cfg.EmailSMTPPort != "25" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadEmailConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvEmailProvider: "log",
	}
	cfg := runtimeconfig.GenerateRuntimeConfig(fs, vals)
	if cfg.EmailProvider != "log" {
		t.Fatalf("want log got %q", cfg.EmailProvider)
	}
}

type recordMail struct{ to, sub, body string }

func (r *recordMail) Send(ctx context.Context, to, subject, body string) error {
	r.to, r.sub, r.body = to, subject, body
	return nil
}

func TestNotifyChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs("a@b.com", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	ctx := context.WithValue(context.Background(), common.KeyQueries, q)
	rec := &recordMail{}
	if err := notifyChange(ctx, rec, "a@b.com", "http://host"); err != nil {
		t.Fatalf("notify error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNotifyChangeErrors(t *testing.T) {
	rec := &recordMail{}
	if err := notifyChange(context.Background(), rec, "", "p"); err == nil {
		t.Fatal("expected error for empty email")
	}
}

type emailRecordProvider struct {
	to   string
	subj string
	body string
}

func (r *emailRecordProvider) Send(ctx context.Context, to, sub, body string) error {
	r.to = to
	r.subj = sub
	r.body = body
	return nil
}

func TestInsertPendingEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs("t@test", "sub", "body").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := q.InsertPendingEmail(context.Background(), InsertPendingEmailParams{ToEmail: "t@test", Subject: "sub", Body: "body"}); err != nil {
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
	q := New(db)
	rows := sqlmock.NewRows([]string{"id", "to_email", "subject", "body"}).AddRow(1, "a@test", "s", "b")
	mock.ExpectQuery("SELECT id, to_email").WillReturnRows(rows)
	mock.ExpectExec("UPDATE pending_emails SET sent_at").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	rec := &emailRecordProvider{}
	processPendingEmail(context.Background(), q, rec)

	if rec.to != "a@test" {
		t.Fatalf("got %q", rec.to)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSendGridProviderFromConfig(t *testing.T) {
	p := providerFromConfig(runtimeconfig.RuntimeConfig{EmailProvider: "sendgrid", EmailSendGridKey: "k"})
	if email.SendgridBuilt {
		if _, ok := p.(email.SendGridProvider); !ok {
			t.Fatalf("expected SendGridProvider, got %#v", p)
		}
	} else {
		if p != nil {
			t.Fatalf("expected nil provider when sendgrid tag not enabled")
		}
	}
}

func TestGetEmailProviderSMTP(t *testing.T) {
	p := providerFromConfig(runtimeconfig.RuntimeConfig{
		EmailProvider: "smtp",
		EmailSMTPHost: "localhost",
		EmailSMTPPort: "25",
	})
	s, ok := p.(email.SMTPProvider)
	if !ok {
		t.Fatalf("expected SMTPProvider, got %#v", p)
	}
	if s.Addr != "localhost:25" || s.From != email.SourceEmail {
		t.Errorf("unexpected provider values: %#v", s)
	}
}

func TestGetEmailProviderLocal(t *testing.T) {
	if _, ok := providerFromConfig(runtimeconfig.RuntimeConfig{EmailProvider: "local"}).(email.LocalProvider); !ok {
		t.Fatalf("expected LocalProvider")
	}
}

func TestGetEmailProviderJMAP(t *testing.T) {
	p := providerFromConfig(runtimeconfig.RuntimeConfig{
		EmailProvider:     "jmap",
		EmailJMAPEndpoint: "http://example.com",
		EmailJMAPAccount:  "acct",
		EmailJMAPIdentity: "id",
	})
	j, ok := p.(email.JMAPProvider)
	if !ok {
		t.Fatalf("expected JMAPProvider, got %#v", p)
	}
	if j.Endpoint != "http://example.com" || j.AccountID != "acct" || j.Identity != "id" {
		t.Errorf("unexpected provider values: %#v", j)
	}
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	if p := providerFromConfig(runtimeconfig.RuntimeConfig{EmailProvider: "ses", EmailAWSRegion: "us-east-1"}); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}
