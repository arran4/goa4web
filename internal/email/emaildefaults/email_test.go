package emaildefaults_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/mail"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	mockdlq "github.com/arran4/goa4web/internal/dlq/mock"
	"github.com/arran4/goa4web/internal/email"
	jmapProv "github.com/arran4/goa4web/internal/email/jmap"
	localProv "github.com/arran4/goa4web/internal/email/local"
	logProv "github.com/arran4/goa4web/internal/email/log"
	mockemail "github.com/arran4/goa4web/internal/email/mock"
	smtpProv "github.com/arran4/goa4web/internal/email/smtp"
	"github.com/arran4/goa4web/workers/emailqueue"
	"strings"
)

func newRegistry() *email.Registry {
	r := email.NewRegistry()
	logProv.Register(r)
	smtpProv.Register(r)
	localProv.Register(r)
	jmapProv.Register(r)
	return r
}

func TestGetEmailProviderLog(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "log"}
	reg := newRegistry()
	p, _ := reg.ProviderFromConfig(&cfg)
	if _, ok := p.(logProv.Provider); !ok {
		t.Errorf("expected LogProvider, got %#v", p)
	}
}

func TestGetEmailProviderUnknown(t *testing.T) {
	cfg := config.RuntimeConfig{EmailProvider: "unknown"}
	reg := newRegistry()
	if p, _ := reg.ProviderFromConfig(&cfg); p != nil {
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
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(k string) string { return env[k] }),
	)
	if cfg.EmailProvider != "smtp" || cfg.EmailSMTPHost != "file" || cfg.EmailSMTPPort != "25" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadEmailConfigFromFileValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	vals := map[string]string{
		config.EnvEmailProvider: "log",
	}
	cfg := config.NewRuntimeConfig(
		config.WithFlagSet(fs),
		config.WithFileValues(vals),
		config.WithGetenv(func(string) string { return "" }),
	)
	if cfg.EmailProvider != "log" {
		t.Fatalf("want log got %q", cfg.EmailProvider)
	}
}
func TestInsertPendingEmail(t *testing.T) {
	q := &db.QuerierStub{}
	if err := q.InsertPendingEmail(context.Background(), db.InsertPendingEmailParams{ToUserID: sql.NullInt32{Int32: 1, Valid: true}, Body: "body", DirectEmail: false}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	emails := q.PendingEmails()
	if len(emails) != 1 {
		t.Fatalf("expected pending email, got %d", len(emails))
	}
	if emails[0].ID == 0 || !emails[0].ToUserID.Valid || emails[0].Body != "body" || emails[0].DirectEmail {
		t.Fatalf("unexpected pending email: %+v", emails[0])
	}
}

func TestEmailQueueWorker(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true

	q := &db.QuerierStub{
		SystemGetUserByIDRow: &db.SystemGetUserByIDRow{
			Idusers:                2,
			Email:                  sql.NullString{String: "bob@example.com", Valid: true},
			Username:               sql.NullString{String: "bob", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
	}
	if err := q.InsertPendingEmail(context.Background(), db.InsertPendingEmailParams{ToUserID: sql.NullInt32{Int32: 2, Valid: true}, Body: "b", DirectEmail: false}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	rec := &mockemail.Provider{}
	if !emailqueue.ProcessPendingEmail(context.Background(), q, rec, nil, cfg) {
		t.Fatal("no email processed")
	}

	if len(rec.Messages) != 1 || rec.Messages[0].To.String() != "\"bob\" <bob@example.com>" {
		t.Fatalf("got %#v", rec.Messages)
	}
	emails := q.PendingEmails()
	if len(emails) != 1 || !emails[0].SentAt.Valid {
		t.Fatalf("pending email not marked sent: %+v", emails)
	}
}

type errProvider struct{}

func (errProvider) Send(context.Context, mail.Address, []byte) error {
	return fmt.Errorf("fail")
}

func (errProvider) TestConfig(context.Context) error {
	return fmt.Errorf("fail")
}

func TestProcessPendingEmailDLQ(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true

	q := &db.QuerierStub{
		SystemGetUserByIDRow: &db.SystemGetUserByIDRow{
			Idusers:                2,
			Email:                  sql.NullString{String: "a@test", Valid: true},
			Username:               sql.NullString{String: "a", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
	}
	if err := q.InsertPendingEmail(context.Background(), db.InsertPendingEmailParams{ToUserID: sql.NullInt32{Int32: 2, Valid: true}, Body: "b", DirectEmail: false}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	pending := q.PendingEmails()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending email, got %d", len(pending))
	}
	for i := 0; i < 4; i++ {
		if err := q.SystemIncrementPendingEmailError(context.Background(), pending[0].ID); err != nil {
			t.Fatalf("increment: %v", err)
		}
	}

	p := errProvider{}
	dlqRec := &mockdlq.Provider{}
	if !emailqueue.ProcessPendingEmail(context.Background(), q, p, dlqRec, cfg) {
		t.Fatal("no email processed")
	}

	if len(dlqRec.Records) != 1 {
		t.Fatalf("dlq records=%d", len(dlqRec.Records))
	}
	msg := dlqRec.Records[0].Message
	if !strings.Contains(msg, "b") || !strings.Contains(msg, "fail") {
		t.Fatalf("unexpected dlq message: %s", msg)
	}
	if len(q.PendingEmails()) != 0 {
		t.Fatalf("expected queue to be empty, got %d entries", len(q.PendingEmails()))
	}
}

func TestGetEmailProviderSMTP(t *testing.T) {
	reg := newRegistry()
	p, _ := reg.ProviderFromConfig(&config.RuntimeConfig{
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
	reg := newRegistry()
	p, _ := reg.ProviderFromConfig(&config.RuntimeConfig{EmailProvider: "local"})
	if _, ok := p.(localProv.Provider); !ok {
		t.Fatalf("expected LocalProvider")
	}
}

func TestGetEmailProviderJMAP(t *testing.T) {
	reg := newRegistry()
	p, _ := reg.ProviderFromConfig(&config.RuntimeConfig{
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

func TestProviderRegistry(t *testing.T) {
	reg := email.NewRegistry()
	called := false
	reg.RegisterProvider("testprov", func(cfg *config.RuntimeConfig) (email.Provider, error) {
		called = true
		return logProv.Provider{}, nil
	})
	p, _ := reg.ProviderFromConfig(&config.RuntimeConfig{EmailProvider: "testprov"})
	if !called {
		t.Fatal("factory not called")
	}
	if _, ok := p.(logProv.Provider); !ok {
		t.Fatalf("expected LogProvider, got %#v", p)
	}
}
