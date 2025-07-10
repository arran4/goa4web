package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	hcommon "github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

type busDummyProvider struct{ to string }

func (d *busDummyProvider) Send(_ context.Context, to mail.Address, _ []byte) error {
	d.to = to.Address
	return nil
}

type recordDLQ struct{ msg string }

func (r *recordDLQ) Record(_ context.Context, m string) error {
	r.msg = m
	return nil
}

func TestBuildPatterns(t *testing.T) {
	cases := map[string][]string{
		"/blog/a/b": {"reply:/blog/a/b", "reply:/blog/a/*", "reply:/blog/*", "reply:/*"},
		"/":         {"reply:/*"},
		"":          {"reply:/*"},
		"/x/y/":     {"reply:/x/y", "reply:/x/*", "reply:/*"},
	}
	for path, expected := range cases {
		got := buildPatterns("Reply", path)
		if len(got) != len(expected) {
			t.Fatalf("%s len %d", path, len(got))
		}
		for i, p := range expected {
			if got[i] != p {
				t.Fatalf("%s pattern %d = %s want %s", path, i, got[i], p)
			}
		}
	}
}

func TestParseEvent(t *testing.T) {
	typ, id, ok := parseEvent("/forum/topic/23/thread/42")
	if !ok || typ != "thread" || id != 42 {
		t.Fatalf("thread parse got %s %d %v", typ, id, ok)
	}
	typ, id, ok = parseEvent("/news/news/9")
	if !ok || typ != "news" || id != 9 {
		t.Fatalf("news parse got %s %d %v", typ, id, ok)
	}
	if _, _, ok := parseEvent("/bad/path"); ok {
		t.Fatalf("unexpected match")
	}
}

func TestRenderMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	mock.MatchExpectationsInOrder(false)
	defer db.Close()
	q := dbpkg.New(db)
	mock.ExpectQuery("SELECT body FROM template_overrides").WithArgs("notify_reply").WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow("Hello {{.Path}}"))
	msg := renderMessage(context.Background(), q, "Reply", "/p", nil)
	if msg != "Hello /p" {
		t.Fatalf("override message %s", msg)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if msg := renderMessage(context.Background(), nil, "Reply", "/p", nil); msg != "New reply in \"/p\" by " {
		t.Fatalf("default msg %q", msg)
	}
}

type errProvider struct{}

func (errProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	return fmt.Errorf("send error")
}

func TestProcessEventDLQ(t *testing.T) {
	ctx := context.Background()
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.NotificationsEnabled = true
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	defer db.Close()
	q := dbpkg.New(db)
	prov := &errProvider{}
	n := Notifier{EmailProvider: prov, Queries: q}
	dlqRec := &recordDLQ{}
	prefRows := sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size", "auto_subscribe_replies"}).
		AddRow(1, 1, 1, true, 15, true)
	mock.ExpectQuery("preferences").WithArgs(int32(1)).WillReturnRows(prefRows)

	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectExec("INSERT INTO subscriptions").WithArgs(int32(1), "reply:/p", "internal").WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(3))

	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(3))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(1))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/*", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/*", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	processEvent(ctx, eventbus.Event{Path: "/p", Task: hcommon.TaskReply, UserID: 1}, n, dlqRec)
	if dlqRec.msg == "" {
		t.Fatal("expected dlq message")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestProcessEventSubscribeSelf(t *testing.T) {
	ctx := context.Background()
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.NotificationsEnabled = true
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	defer db.Close()
	q := dbpkg.New(db)
	n := Notifier{Queries: q}

	prefRows := sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size", "auto_subscribe_replies"}).
		AddRow(1, 1, 1, true, 15, true)
	mock.ExpectQuery("preferences").WithArgs(int32(1)).WillReturnRows(prefRows)

	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectExec("INSERT INTO subscriptions").WithArgs(int32(1), "reply:/p", "internal").WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectExec("INSERT INTO subscriptions").WithArgs(int32(1), "reply:/p", "email").WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(1))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/p", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(1))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/*", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/*", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))

	processEvent(ctx, eventbus.Event{Path: "/p", Task: hcommon.TaskReply, UserID: 1}, n, nil)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestProcessEventNoAutoSubscribe(t *testing.T) {
	ctx := context.Background()
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.NotificationsEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	n := Notifier{Queries: q}

	prefRows := sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size", "auto_subscribe_replies"}).
		AddRow(1, 1, 1, true, 15, false)
	mock.ExpectQuery("preferences").WithArgs(int32(1)).WillReturnRows(prefRows)

	processEvent(ctx, eventbus.Event{Path: "/p", Task: hcommon.TaskReply, UserID: 1}, n, nil)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestProcessEventAdminNotify(t *testing.T) {
	ctx := context.Background()
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.AdminEmails = "a@test"
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	config.AppRuntimeConfig.NotificationsEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	prov := &busDummyProvider{}
	n := Notifier{EmailProvider: prov, Queries: q}

	mock.ExpectQuery("UserByEmail").
		WithArgs(sql.NullString{String: "a@test", Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(1, "a@test", "a"))
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(int32(1), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WithArgs(int32(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	processEvent(ctx, eventbus.Event{Path: "/admin/x", Task: hcommon.TaskSetTopicRestriction, UserID: 1, Admin: true}, n, nil)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestBusWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.NotificationsEnabled = true
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	bus := eventbus.NewBus()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)

	prov := &busDummyProvider{}
	n := Notifier{EmailProvider: prov, Queries: q}

	mock.ExpectQuery("SELECT body FROM template_overrides").
		WithArgs("notify_register").
		WillReturnRows(sqlmock.NewRows([]string{"body"}))

	mock.ExpectQuery("subscriptions").
		WithArgs("register:/*", "email").
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(2))

	mock.ExpectQuery("subscriptions").
		WithArgs("register:/*", "internal").
		WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(3))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username FROM users u LEFT JOIN user_emails ue ON ue.id = ( SELECT id FROM user_emails ue2 WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1 ) WHERE u.idusers = ?")).
		WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).
			AddRow(2, sql.NullString{String: "e@example.com", Valid: true}, sql.NullString{String: "u", Valid: true}))

	mock.ExpectExec("INSERT INTO pending_emails").
		WithArgs(int32(2), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO notifications").
		WithArgs(int32(3), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		BusWorker(ctx, bus, n, nil)
	}()

	time.Sleep(10 * time.Millisecond)

	bus.Publish(eventbus.Event{Path: "/", Task: hcommon.TaskRegister, UserID: 1, Item: SignupInfo{Username: "bob"}})
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
