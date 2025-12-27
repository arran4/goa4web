package notifications_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type allowTaskNoEmail struct{ user.PermissionUserAllowTask }

func (allowTaskNoEmail) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return nil, false
}

type disallowTaskNoEmail struct {
	user.PermissionUserDisallowTask
}

func (disallowTaskNoEmail) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return nil, false
}

type updateTaskNoEmail struct{ user.PermissionUpdateTask }

func (updateTaskNoEmail) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return nil, false
}

func assertCallCount(t *testing.T, method string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %d %s calls got %d", want, method, got)
	}
}

func TestProcessEventPermissionTasks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = true
	cfg.EmailFrom = "from@example.com"

	bus := eventbus.NewBus()
	q := &db.QuerierStub{
		SystemGetUserByIDRow: &db.SystemGetUserByIDRow{
			Idusers:                2,
			Email:                  sql.NullString{String: "u@test", Valid: true},
			Username:               sql.NullString{String: "bob", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
	}
	n := notif.New(notif.WithQueries(q), notif.WithConfig(cfg))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.BusWorker(ctx, bus, nil)
	}()

	time.Sleep(10 * time.Millisecond)

	cases := []struct {
		task tasks.Task
		tmpl string
	}{
		{allowTaskNoEmail{user.PermissionUserAllowTask{TaskString: user.TaskUserAllow}}, "set_user_role.gotxt"},
		{disallowTaskNoEmail{user.PermissionUserDisallowTask{TaskString: user.TaskUserDisallow}}, "delete_user_role.gotxt"},
		{updateTaskNoEmail{user.PermissionUpdateTask{TaskString: user.TaskUpdate}}, "update_user_role.gotxt"},
	}

	for _, c := range cases {
		bus.Publish(eventbus.TaskEvent{Path: "/admin", Task: c.task, UserID: 1, Data: map[string]any{"targetUserID": int32(2), "Username": "bob"}, Outcome: eventbus.TaskOutcomeSuccess})
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	assertCallCount(t, "SystemGetUserByID", len(q.SystemGetUserByIDCalls), len(cases))
	assertCallCount(t, "SystemCreateNotification", len(q.SystemCreateNotificationCalls), len(cases))
	for _, call := range q.SystemCreateNotificationCalls {
		if call.RecipientID != int32(2) {
			t.Fatalf("expected recipient 2 got %d", call.RecipientID)
		}
		if call.Link.String != "/admin" || !call.Link.Valid {
			t.Fatalf("expected link /admin got %q (valid=%v)", call.Link.String, call.Link.Valid)
		}
		if !call.Message.Valid || call.Message.String == "" {
			t.Fatalf("expected non-empty message")
		}
	}
}
