package notifications_test

import (
	"context"
	"database/sql"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type allowTaskNoEmail struct{ user.PermissionUserAllowTask }

func (allowTaskNoEmail) TargetEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates { return nil }

type disallowTaskNoEmail struct {
	user.PermissionUserDisallowTask
}

func (disallowTaskNoEmail) TargetEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return nil
}

type updateTaskNoEmail struct{ user.PermissionUpdateTask }

func (updateTaskNoEmail) TargetEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return nil
}

func TestProcessEventPermissionTasks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = true
	cfg.EmailFrom = "from@example.com"

	bus := eventbus.NewBus()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)
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
		mock.ExpectQuery(regexp.QuoteMeta("SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at FROM users u LEFT JOIN user_emails ue ON ue.id = ( SELECT id FROM user_emails ue2 WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1 ) WHERE u.idusers = ?")).
			WithArgs(int32(2)).
			WillReturnRows(sqlmock.NewRows([]string{"idusers", "email", "username", "public_profile_enabled_at"}).AddRow(2, "u@test", "bob", nil))
		mock.ExpectQuery("SystemGetLastNotificationForRecipientByMessage").
			WithArgs(int32(2), "missing email address").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO notifications (users_idusers, link, message) VALUES (?, ?, ?)")).
			WithArgs(int32(2), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		bus.Publish(eventbus.TaskEvent{Path: "/admin", Task: c.task, UserID: 1, Data: map[string]any{"targetUserID": int32(2), "Username": "bob"}, Outcome: eventbus.TaskOutcomeSuccess})
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

}
