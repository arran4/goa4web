package notifications

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"net/mail"
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
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	prov := &errProvider{}
	n := New(q, prov)
	dlqRec := &recordDLQ{}

	if err := n.processEvent(ctx, eventbus.Event{Path: "/p", Task: TestTask{TaskString: TaskTest}, UserID: 1}, dlqRec); err != nil {
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
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.NotificationsEnabled = true
	config.AppRuntimeConfig.EmailFrom = "from@example.com"
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	n := New(q, nil)

	if err := n.processEvent(ctx, eventbus.Event{Path: "/p", Task: TaskTest, UserID: 1}, nil); err != nil {
		t.Fatalf("process: %v", err)
	}
}

func TestProcessEventNoAutoSubscribe(t *testing.T) {
	ctx := context.Background()
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.EmailEnabled = true
	config.AppRuntimeConfig.AdminNotify = true
	config.AppRuntimeConfig.NotificationsEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	n := New(q, nil)

	if err := n.processEvent(ctx, eventbus.Event{Path: "/p", Task: TaskTest, UserID: 1}, nil); err != nil {
		t.Fatalf("process: %v", err)
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
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	prov := &busDummyProvider{}
	n := New(q, prov)

	if err := n.processEvent(ctx, eventbus.Event{Path: "/admin/x", Task: TaskTest, UserID: 1}, nil); err != nil {
		t.Fatalf("process: %v", err)
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
	defer db.Close()
	q := dbpkg.New(db)
	n := New(q, nil)

	if err := n.processEvent(ctx, eventbus.Event{Path: "/writings/article/1", Task: TaskTest, UserID: 2, Data: map[string]any{"target": Target{Type: "writing", ID: 1}}}, nil); err != nil {
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

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)

	prov := &busDummyProvider{}
	n := New(q, prov)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.BusWorker(ctx, bus, nil)
	}()

	time.Sleep(10 * time.Millisecond)

	bus.Publish(eventbus.Event{Path: "/", Task: TaskTest, UserID: 1, Data: map[string]any{"signup": SignupInfo{Username: "bob"}}})
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	if prov.to != "" {
		t.Fatalf("unexpected email sent to %s", prov.to)
	}
}
