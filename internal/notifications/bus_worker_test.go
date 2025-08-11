package notifications

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/workers/postcountworker"
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

func TestBuildPatternsAdditional(t *testing.T) {
	type testCase struct {
		task tasks.TaskString
		path string
		want []string
	}
	cases := []testCase{
		{tasks.TaskString("Reply"), "/writings/article/2", []string{"reply:/writings/article/2", "reply:/writings/article/*", "reply:/writings/*", "reply:/*"}},
		{tasks.TaskString("Reply"), "/news/news/14", []string{"reply:/news/news/14", "reply:/news/news/*", "reply:/news/*", "reply:/*"}},
		{tasks.TaskString("Post"), "/blog/3", []string{"post:/blog/3", "post:/blog/*", "post:/*"}},
		{tasks.TaskString("Post"), "/writing/5", []string{"post:/writing/5", "post:/writing/*", "post:/*"}},
		{tasks.TaskString("Post"), "/news/8", []string{"post:/news/8", "post:/news/*", "post:/*"}},
		{tasks.TaskString("Post"), "/image/9", []string{"post:/image/9", "post:/image/*", "post:/*"}},
	}
	for _, tc := range cases {
		got := buildPatterns(tc.task, tc.path)
		if len(got) != len(tc.want) {
			t.Fatalf("%s len %d", tc.path, len(got))
		}
		for i, p := range tc.want {
			if got[i] != p {
				t.Fatalf("%s pattern %d = %s want %s", tc.path, i, got[i], p)
			}
		}
	}
}

func TestCollectSubscribersQuery(t *testing.T) {
	ctx := context.Background()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	patterns := []string{"post:/blog/1", "post:/blog/*"}
	rows := sqlmock.NewRows([]string{"users_idusers"}).AddRow(1).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT users_idusers FROM subscriptions WHERE pattern IN (?,?) AND method = ?")).
		WithArgs(patterns[0], patterns[1], "email").
		WillReturnRows(rows)

	subs, err := collectSubscribers(ctx, q, patterns, "email")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("want 2 subs got %d", len(subs))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

const TaskTest = tasks.TaskString("TaskTest")

type TestTask struct {
	TaskString tasks.TaskString
}

func (t TestTask) Action(w http.ResponseWriter, r *http.Request) any {

	return nil
}

type errProvider struct{}

func (errProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	return fmt.Errorf("send error")
}

func TestProcessEventDLQ(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.NotificationsEnabled = true
	cfg.EmailFrom = "from@example.com"

	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	prov := &errProvider{}
	n := New(WithQueries(q), WithEmailProvider(prov), WithConfig(cfg))
	dlqRec := &recordDLQ{}

	if err := n.processEvent(ctx, eventbus.TaskEvent{Path: "/p", Task: TestTask{TaskString: TaskTest}, UserID: 1, Outcome: eventbus.TaskOutcomeSuccess}, dlqRec); err != nil {
		t.Fatalf("process: %v", err)
	}
	if dlqRec.msg != "" {
		t.Fatalf("unexpected dlq message: %s", dlqRec.msg)
	}
	if dlqRec.msg != "" {
		t.Fatalf("unexpected dlq message: %s", dlqRec.msg)
	}
}

func TestProcessEventSubscribeSelf(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.NotificationsEnabled = true
	cfg.EmailFrom = "from@example.com"

	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	n := New(WithQueries(q), WithConfig(cfg))

	if err := n.processEvent(ctx, eventbus.TaskEvent{Path: "/p", Task: TaskTest, UserID: 1, Outcome: eventbus.TaskOutcomeSuccess}, nil); err != nil {
		t.Fatalf("process: %v", err)
	}
}

func TestProcessEventNoAutoSubscribe(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.NotificationsEnabled = true

	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	n := New(WithQueries(q), WithConfig(cfg))

	if err := n.processEvent(ctx, eventbus.TaskEvent{Path: "/p", Task: TaskTest, UserID: 1, Outcome: eventbus.TaskOutcomeSuccess}, nil); err != nil {
		t.Fatalf("process: %v", err)
	}
}

func TestProcessEventAdminNotify(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.AdminEmails = "a@test"
	cfg.EmailFrom = "from@example.com"
	cfg.NotificationsEnabled = true

	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	prov := &busDummyProvider{}
	n := New(WithQueries(q), WithEmailProvider(prov), WithConfig(cfg))

	if err := n.processEvent(ctx, eventbus.TaskEvent{Path: "/admin/x", Task: TaskTest, UserID: 1, Outcome: eventbus.TaskOutcomeSuccess}, nil); err != nil {
		t.Fatalf("process: %v", err)
	}
}

func TestProcessEventWritingSubscribers(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.NotificationsEnabled = true
	cfg.EmailFrom = "from@example.com"

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	n := New(WithQueries(q), WithConfig(cfg))

	if err := n.processEvent(ctx, eventbus.TaskEvent{Path: "/writings/article/1", Task: TaskTest, UserID: 2, Data: map[string]any{"target": Target{Type: "writing", ID: 1}}, Outcome: eventbus.TaskOutcomeSuccess}, nil); err != nil {
		t.Fatalf("process: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

type targetTask struct{ tasks.TaskString }

func (targetTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (targetTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) { return []int32{2, 3}, nil }

func (targetTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *EmailTemplates, send bool) {
	return nil, false
}

func (targetTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	t := NotificationTemplateFilenameGenerator("announcement")
	return &t
}

func TestProcessEventTargetUsers(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = true

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	n := New(WithQueries(q), WithConfig(cfg))

	for _, id := range []int32{2, 3} {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at FROM users u LEFT JOIN user_emails ue ON ue.id = ( SELECT id FROM user_emails ue2 WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1 ) WHERE u.idusers = ?")).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(id, "u@test", fmt.Sprintf("u%d", id), nil))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT body FROM template_overrides WHERE name = ?")).
			WithArgs("announcement.gotxt").
			WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow(""))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO notifications (users_idusers, link, message) VALUES (?, ?, ?)")).
			WithArgs(id, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	evt := eventbus.TaskEvent{Path: "/announce/1", Task: targetTask{TaskString: "Target"}, UserID: 1, Data: map[string]any{"Username": "bob"}, Outcome: eventbus.TaskOutcomeSuccess}

	if err := n.processEvent(ctx, evt, nil); err != nil {
		t.Fatalf("process: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestBusWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.NotificationsEnabled = true
	cfg.EmailFrom = "from@example.com"

	bus := eventbus.NewBus()

	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	prov := &busDummyProvider{}
	n := New(WithQueries(q), WithEmailProvider(prov), WithConfig(cfg))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.BusWorker(ctx, bus, nil)
	}()

	time.Sleep(10 * time.Millisecond)

	bus.Publish(eventbus.TaskEvent{Path: "/", Task: TaskTest, UserID: 1, Data: map[string]any{"Username": "bob"}, Outcome: eventbus.TaskOutcomeSuccess})
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	if prov.to != "" {
		t.Fatalf("unexpected email sent to %s", prov.to)
	}
}

type autoSubTask struct{ tasks.TaskString }

func (autoSubTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (autoSubTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return "AutoSub", fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return "AutoSub", evt.Path, nil
}

func TestProcessEventAutoSubscribe(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = true

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
	n := New(WithQueries(q), WithConfig(cfg))

	prefRows := sqlmock.NewRows([]string{"idpreferences", "language_idlanguage", "users_idusers", "emailforumupdates", "page_size", "auto_subscribe_replies", "timezone"}).
		AddRow(1, 0, 1, nil, 0, true, nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idpreferences, language_idlanguage, users_idusers, emailforumupdates, page_size, auto_subscribe_replies, timezone FROM preferences WHERE users_idusers = ?")).
		WithArgs(int32(1)).WillReturnRows(prefRows)

	pattern := buildPatterns(tasks.TaskString("AutoSub"), "/forum/topic/7/thread/42")[0]
	mock.ExpectQuery(regexp.QuoteMeta("SELECT users_idusers FROM subscriptions WHERE pattern = ? AND method = ?")).
		WithArgs(pattern, "internal").WillReturnRows(sqlmock.NewRows([]string{"users_idusers"}))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO subscriptions (users_idusers, pattern, method) VALUES (?, ?, ?)")).
		WithArgs(int32(1), pattern, "internal").WillReturnResult(sqlmock.NewResult(1, 1))

	evt := eventbus.TaskEvent{
		Path:   "/forum/topic/7/thread/42/reply",
		UserID: 1,
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{CommentID: 1, ThreadID: 42, TopicID: 7},
		},
		Outcome: eventbus.TaskOutcomeSuccess,
	}

	n.handleAutoSubscribe(ctx, evt, autoSubTask{TaskString: "AutoSub"})
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
