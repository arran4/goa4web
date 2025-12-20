package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
	"sync"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/tasks"

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

func (d *busDummyProvider) TestConfig(ctx context.Context) error { return nil }

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

type querierStub struct {
	db.QuerierStub
	mu sync.Mutex

	SystemGetUserByIDFunc func(context.Context, int32) (*db.SystemGetUserByIDRow, error)
}

func (q *querierStub) SystemGetUserByID(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if q.SystemGetUserByIDFunc != nil {
		return q.SystemGetUserByIDFunc(ctx, id)
	}
	return q.QuerierStub.SystemGetUserByID(ctx, id)
}

func TestCollectSubscribersQuery(t *testing.T) {
	ctx := context.Background()
	patterns := []string{"post:/blog/1", "post:/blog/*"}

	q := &querierStub{
		QuerierStub: db.QuerierStub{
			ListSubscribersForPatternsReturn: map[string][]int32{
				"post:/blog/1": {1},
				"post:/blog/*": {2},
			},
		},
	}

	subs, err := collectSubscribers(ctx, q, patterns, "email")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("want 2 subs got %d", len(subs))
	}
	if len(q.ListSubscribersForPatternsParams) != 1 {
		t.Fatalf("expected 1 call, got %d", len(q.ListSubscribersForPatternsParams))
	}
	args := q.ListSubscribersForPatternsParams[0]
	if len(args.Patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(args.Patterns))
	}
	if args.Method != "email" {
		t.Fatalf("expected method email, got %s", args.Method)
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

func (errProvider) TestConfig(ctx context.Context) error {
	return fmt.Errorf("test config error")
}

func TestProcessEventDLQ(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewRuntimeConfig()
	cfg.EmailEnabled = true
	cfg.AdminNotify = true
	cfg.NotificationsEnabled = true
	cfg.EmailFrom = "from@example.com"

	q := &querierStub{}
	prov := &errProvider{}
	n := New(WithQueries(q), WithEmailProvider(prov), WithConfig(cfg))
	dlqRec := &recordDLQ{}

	if err := n.processEvent(ctx, eventbus.TaskEvent{Path: "/p", Task: TestTask{TaskString: TaskTest}, UserID: 1, Outcome: eventbus.TaskOutcomeSuccess}, dlqRec); err != nil {
		t.Fatalf("process: %v", err)
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

	q := &querierStub{}
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

	q := &querierStub{}
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

	q := &querierStub{
		QuerierStub: db.QuerierStub{
			SystemGetUserByEmailRow: &db.SystemGetUserByEmailRow{Idusers: 1, Email: "a@test", Username: sql.NullString{String: "a", Valid: true}},
		},
	}

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

	q := &querierStub{}
	n := New(WithQueries(q), WithConfig(cfg))

	if err := n.processEvent(ctx, eventbus.TaskEvent{Path: "/writings/article/1", Task: TaskTest, UserID: 2, Data: map[string]any{"target": Target{Type: "writing", ID: 1}}, Outcome: eventbus.TaskOutcomeSuccess}, nil); err != nil {
		t.Fatalf("process: %v", err)
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

	q := &querierStub{}
	q.SystemGetTemplateOverrideReturns = ""
	// Override SystemGetUserByID for different users
	q.SystemGetUserByIDFunc = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
		return &db.SystemGetUserByIDRow{Idusers: id, Email: sql.NullString{String: "u@test", Valid: true}, Username: sql.NullString{String: fmt.Sprintf("u%d", id), Valid: true}}, nil
	}

	n := New(WithQueries(q), WithConfig(cfg))

	evt := eventbus.TaskEvent{Path: "/announce/1", Task: targetTask{TaskString: "Target"}, UserID: 1, Data: map[string]any{"Username": "bob"}, Outcome: eventbus.TaskOutcomeSuccess}

	if err := n.processEvent(ctx, evt, nil); err != nil {
		t.Fatalf("process: %v", err)
	}
	// Check calls
	if len(q.SystemCreateNotificationCalls) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(q.SystemCreateNotificationCalls))
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

	q := &querierStub{}
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

	q := &querierStub{
		QuerierStub: db.QuerierStub{
			GetPreferenceForListerReturn: map[int32]*db.Preference{
				1: {Idpreferences: 1, UsersIdusers: 1, AutoSubscribeReplies: true, Emailforumupdates: sql.NullBool{Bool: false, Valid: true}},
			},
			ListSubscribersForPatternReturn: map[string][]int32{
				// "autosub:/forum/topic/7/thread/42": {1}, // Assuming already subscribed if we wanted to test idempotency, but here we expect insert
			},
		},
	}

	n := New(WithQueries(q), WithConfig(cfg))

	evt := eventbus.TaskEvent{
		Path:   "/forum/topic/7/thread/42/reply",
		UserID: 1,
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{CommentID: 1, ThreadID: 42, TopicID: 7},
		},
		Outcome: eventbus.TaskOutcomeSuccess,
	}

	if err := n.handleAutoSubscribe(ctx, evt, autoSubTask{TaskString: "AutoSub"}); err != nil {
		t.Fatalf("handleAutoSubscribe: %v", err)
	}

	if len(q.InsertSubscriptionParams) != 1 {
		t.Fatalf("expected 1 subscription insert, got %d", len(q.InsertSubscriptionParams))
	}
	pattern := buildPatterns(tasks.TaskString("AutoSub"), "/forum/topic/7/thread/42")[0]
	if q.InsertSubscriptionParams[0].Pattern != pattern {
		t.Fatalf("expected pattern %s, got %s", pattern, q.InsertSubscriptionParams[0].Pattern)
	}
}
