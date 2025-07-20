package notifications_test

import (
	"context"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	user "github.com/arran4/goa4web/handlers/user"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestProcessEventPermissionTasks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	origCfg := config.AppRuntimeConfig
	config.AppRuntimeConfig.NotificationsEnabled = true
	t.Cleanup(func() { config.AppRuntimeConfig = origCfg })

	bus := eventbus.NewBus()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	q := dbpkg.New(db)
	n := notif.New(notif.WithQueries(q))

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
		{user.PermissionUserAllowTask{TaskString: user.TaskUserAllow}, "permission_user_allow.gotxt"},
		{user.PermissionUserDisallowTask{TaskString: user.TaskUserDisallow}, "permission_user_disallow.gotxt"},
		{user.PermissionUpdateTask{TaskString: user.TaskUpdate}, "permission_update.gotxt"},
	}

	for _, c := range cases {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username FROM users u LEFT JOIN user_emails ue ON ue.id = ( SELECT id FROM user_emails ue2 WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1 ) WHERE u.idusers = ?")).
			WithArgs(int32(2)).
			WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username"}).AddRow(2, "u@test", "bob"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT body FROM template_overrides WHERE name = ?")).
			WithArgs(c.tmpl).
			WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow(""))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO notifications (users_idusers, link, message) VALUES (?, ?, ?)")).
			WithArgs(int32(2), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		bus.Publish(eventbus.Event{Path: "/admin", Task: c.task, UserID: 1, Data: map[string]any{"UserID": int32(2), "Username": "bob"}})
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
