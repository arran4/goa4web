package goa4web

import (
	"context"
	"os"
	"testing"
)

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
