package notifications

import (
	"context"
	"database/sql"
	"net/mail"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

func TestNotificationsQueries(t *testing.T) {
	ctx := context.Background()
	q := &db.QuerierStub{}

	insert := func(recipient int32, msg string) {
		t.Helper()
		if err := q.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{
			RecipientID: recipient,
			Link:        sql.NullString{String: "/x", Valid: true},
			Message:     sql.NullString{String: msg, Valid: true},
		}); err != nil {
			t.Fatalf("insert %s: %v", msg, err)
		}
	}

	insert(1, "hi")
	insert(1, "later")
	insert(2, "other")

	if c, err := q.GetUnreadNotificationCountForLister(ctx, 1); err != nil || c != 2 {
		t.Fatalf("count=%d err=%v", c, err)
	}

	unread, err := q.ListUnreadNotificationsForLister(ctx, db.ListUnreadNotificationsForListerParams{ListerID: 1, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("get unread: %v", err)
	}
	if len(unread) != 2 {
		t.Fatalf("expected 2 unread, got %d", len(unread))
	}

	last, err := q.SystemGetLastNotificationForRecipientByMessage(ctx, db.SystemGetLastNotificationForRecipientByMessageParams{
		RecipientID: 1,
		Message:     sql.NullString{String: "hi", Valid: true},
	})
	if err != nil || last == nil {
		t.Fatalf("last notification lookup: %v", err)
	}
	if last.Message.String != "hi" {
		t.Fatalf("expected message hi got %s", last.Message.String)
	}

	if err := q.SetNotificationReadForLister(ctx, db.SetNotificationReadForListerParams{ID: last.ID, ListerID: 1}); err != nil {
		t.Fatalf("mark: %v", err)
	}
	if c, err := q.GetUnreadNotificationCountForLister(ctx, 1); err != nil || c != 1 {
		t.Fatalf("count after mark=%d err=%v", c, err)
	}
	if err := q.SetNotificationUnreadForLister(ctx, db.SetNotificationUnreadForListerParams{ID: last.ID, ListerID: 1}); err != nil {
		t.Fatalf("unmark: %v", err)
	}
	if c, err := q.GetUnreadNotificationCountForLister(ctx, 1); err != nil || c != 2 {
		t.Fatalf("count after unmark=%d err=%v", c, err)
	}

	if err := q.SetNotificationReadForLister(ctx, db.SetNotificationReadForListerParams{ID: last.ID, ListerID: 1}); err != nil {
		t.Fatalf("mark again: %v", err)
	}
	if err := q.AdminPurgeReadNotifications(ctx); err != nil {
		t.Fatalf("purge: %v", err)
	}
	all, err := q.ListNotificationsForLister(ctx, db.ListNotificationsForListerParams{ListerID: 1, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 || all[0].ReadAt.Valid {
		t.Fatalf("expected 1 unread notification after purge, got %+v", all)
	}
}

type dummyProvider struct{ to string }

func (r *dummyProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	r.to = to.Address
	return nil
}

func (r *dummyProvider) TestConfig(ctx context.Context) error { return nil }

func TestNotifierNotifyAdmins(t *testing.T) {
	q := &db.QuerierStub{
		SystemGetUserByEmailRow: &db.SystemGetUserByEmailRow{Idusers: 1, Email: "a@test", Username: sql.NullString{String: "a", Valid: true}},
	}
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.AdminEmails = "a@test"
	cfg.EmailFrom = "from@example.com"
	cfg.NotificationsEnabled = true

	rec := &dummyProvider{}
	n := New(WithQueries(q), WithEmailProvider(rec), WithConfig(cfg))
	n.NotifyAdmins(context.Background(), &EmailTemplates{}, EmailData{})
	if rec.to != "" {
		t.Fatalf("expected no direct mail got %s", rec.to)
	}
	if len(q.SystemGetUserByEmailCalls) != 1 || q.SystemGetUserByEmailCalls[0] != "a@test" {
		t.Fatalf("expected user lookup for a@test, got %v", q.SystemGetUserByEmailCalls)
	}
	if len(q.InsertPendingEmailCalls) != 1 {
		t.Fatalf("expected 1 queued email got %d", len(q.InsertPendingEmailCalls))
	}
	if !q.InsertPendingEmailCalls[0].ToUserID.Valid || q.InsertPendingEmailCalls[0].ToUserID.Int32 != 1 {
		t.Fatalf("expected queued email for user 1, got %+v", q.InsertPendingEmailCalls[0].ToUserID)
	}
	if q.InsertPendingEmailCalls[0].DirectEmail {
		t.Fatalf("expected queued email to be non-direct")
	}
	if q.InsertPendingEmailCalls[0].Body == "" {
		t.Fatalf("expected non-empty email body")
	}
}

func TestNotifierInitialization(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	n := New(WithConfig(cfg))
	if n.Queries != nil {
		t.Fatalf("expected nil Queries")
	}
	q := &db.QuerierStub{}
	n = New(WithQueries(q), WithConfig(cfg))
	if n.Queries != q {
		t.Fatalf("queries not set via option")
	}
}
