package emailutil_test

import (
	"context"
	"flag"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	jmapProv "github.com/arran4/goa4web/internal/email/jmap"
	localProv "github.com/arran4/goa4web/internal/email/local"
	logProv "github.com/arran4/goa4web/internal/email/log"
	sesProv "github.com/arran4/goa4web/internal/email/ses"
	smtpProv "github.com/arran4/goa4web/internal/email/smtp"
	"github.com/arran4/goa4web/internal/emailutil"
	"github.com/arran4/goa4web/runtimeconfig"
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
	if p := email.ProviderFromConfig(cfg); reflect.TypeOf(p) != reflect.TypeOf(logProv.Provider{}) {
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

type recordMail struct {
	to, sub string
	raw     []byte
}

func (r *recordMail) Send(ctx context.Context, to, subject string, rawEmailMessage []byte) error {
	r.to, r.sub, r.raw = to, subject, rawEmailMessage
	return nil
}

func TestNotifyChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs("a@b.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	ctx := context.WithValue(context.Background(), common.KeyQueries, q)
	rec := &recordMail{}
	if err := emailutil.NotifyChange(ctx, rec, "a@b.com", "http://host", "update", nil); err != nil {
		t.Fatalf("notify error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestNotifyChangeErrors(t *testing.T) {
	rec := &recordMail{}
	if err := emailutil.NotifyChange(context.Background(), rec, "", "p", "update", nil); err == nil {
		t.Fatal("expected error for empty email")
	}
}

type emailRecordProvider struct {
	to   string
	subj string
	raw  []byte
}

func (r *emailRecordProvider) Send(ctx context.Context, to, sub string, rawEmailMessage []byte) error {
	r.to = to
	r.subj = sub
	r.raw = rawEmailMessage
	return nil
}

func TestInsertPendingEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	q := dbpkg.New(db)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs("t@test", "sub", "body", "html").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := q.InsertPendingEmail(context.Background(), dbpkg.InsertPendingEmailParams{ToEmail: "t@test", Subject: "sub", Body: "body", HtmlBody: "html"}); err != nil {
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
	rows := sqlmock.NewRows([]string{"id", "to_email", "subject", "body", "html_body"}).AddRow(1, "a@test", "s", "b", "h")
	mock.ExpectQuery("SELECT id, to_email").WillReturnRows(rows)
	mock.ExpectExec("UPDATE pending_emails SET sent_at").WithArgs(int32(1)).WillReturnResult(sqlmock.NewResult(1, 1))

	rec := &emailRecordProvider{}
	emailutil.ProcessPendingEmail(context.Background(), q, rec)

	if rec.to != "a@test" {
		t.Fatalf("got %q", rec.to)
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
