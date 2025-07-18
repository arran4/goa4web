package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"net/mail"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
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
		got := buildPatterns(tasks.TaskString("Reply"), path)
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
	evt := eventbus.Event{Data: map[string]any{"target": Target{Type: "thread", ID: 42}}}
	typ, id, ok := parseEvent(evt)
	if !ok || typ != "thread" || id != 42 {
		t.Fatalf("thread parse got %s %d %v", typ, id, ok)
	}
	evt = eventbus.Event{Data: map[string]any{"target": Target{Type: "writing", ID: 7}}}
	typ, id, ok = parseEvent(evt)
	if !ok || typ != "writing" || id != 7 {
		t.Fatalf("writing parse got %s %d %v", typ, id, ok)
	}
	if _, _, ok := parseEvent(eventbus.Event{Path: "/bad/path"}); ok {
		t.Fatalf("unexpected match")
	}
	if _, _, ok := parseEvent(eventbus.Event{Path: "/news/news/9"}); ok {
		t.Fatalf("unexpected match with path")
	}
}

const TaskTest = tasks.TaskString("TaskTest")

type TestTask struct {
	TaskString tasks.TaskString
}

func (t TestTask) Action(w http.ResponseWriter, r *http.Request) {

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
	if err := processEvent(ctx, eventbus.Event{Path: "/p", Task: TestTask{TaskString: TaskTest}, UserID: 1}, n, dlqRec); err == nil {
		t.Fatal("expected error")
	}
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

	if err := processEvent(ctx, eventbus.Event{Path: "/p", Task: TaskTest, UserID: 1}, n, nil); err != nil {
		t.Fatalf("process: %v", err)
	}

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

	if err := processEvent(ctx, eventbus.Event{Path: "/p", Task: TaskTest, UserID: 1}, n, nil); err != nil {
		t.Fatalf("process: %v", err)
	}

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

	if err := processEvent(ctx, eventbus.Event{Path: "/admin/x", Task: TaskTest, UserID: 1}, n, nil); err != nil {
		t.Fatalf("process: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestProcessEventWritingSubscribers(t *testing.T) {
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
		AddRow(1, 1, 2, true, 15, true)
	mock.ExpectQuery("preferences").WithArgs(int32(2)).WillReturnRows(prefRows)

	mock.ExpectQuery("subscriptions").WithArgs("reply:/writings/article/1", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectExec("INSERT INTO subscriptions").WithArgs(int32(2), "reply:/writings/article/1", "internal").WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("subscriptions").WithArgs("reply:/writings/article/1", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectExec("INSERT INTO subscriptions").WithArgs(int32(2), "reply:/writings/article/1", "email").WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("subscriptions").WithArgs("reply:/writings/article/1", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(1))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/writings/article/1", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}).AddRow(2))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/*", "email").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectQuery("subscriptions").WithArgs("reply:/*", "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))

	rows := sqlmock.NewRows([]string{
		"idwriting", "users_idusers", "forumthread_id", "language_idlanguage",
		"writing_category_id", "title", "published", "writing", "abstract", "private", "deleted_at",
		"idusers", "username", "deleted_at_2", "idpreferences", "language_idlanguage_2",
		"users_idusers_2", "emailforumupdates", "page_size", "auto_subscribe_replies", "email",
	}).AddRow(1, 2, 3, 1, 4, "t", nil, "w", "a", 0, nil, 2, "bob", nil, 1, 1, 2, 1, 10, true, "e@test")
	mock.ExpectQuery("SELECT idwriting").WithArgs(int32(1), int32(2)).WillReturnRows(rows)
	mock.ExpectExec("INSERT INTO pending_emails").WithArgs(int32(2), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO notifications").WithArgs(int32(2), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := processEvent(ctx, eventbus.Event{Path: "/writings/article/1", Task: TaskTest, UserID: 2, Data: map[string]any{"target": Target{Type: "writing", ID: 1}}}, n, nil); err != nil {
		t.Fatalf("process: %v", err)
	}

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

	bus.Publish(eventbus.Event{Path: "/", Task: TaskTest, UserID: 1, Data: map[string]any{"signup": SignupInfo{Username: "bob"}}})
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
