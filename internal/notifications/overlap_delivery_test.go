package notifications

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// overlapReplyTask represents the real forum-reply combination of admin and
// subscriber notification providers without importing handlers/forum.
type overlapReplyTask struct{ tasks.TaskString }

func (overlapReplyTask) Action(http.ResponseWriter, *http.Request) any { return nil }

func (overlapReplyTask) AdminEmailTemplate(eventbus.TaskEvent) (*EmailTemplates, bool) {
	return nil, true // Admin internal notification without queuing email in this test.
}

func (overlapReplyTask) AdminInternalNotificationTemplate(eventbus.TaskEvent) *string {
	name := NotificationTemplateFilenameGenerator("overlap-admin")
	return &name
}

func (overlapReplyTask) SubscribedEmailTemplate(eventbus.TaskEvent) (*EmailTemplates, bool) {
	return nil, false
}

func (overlapReplyTask) SubscribedInternalNotificationTemplate(eventbus.TaskEvent) *string {
	name := NotificationTemplateFilenameGenerator("overlap-subscriber")
	return &name
}

func TestAdminAndSubscriberContributeToSameReplyActivity(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.AdminNotify = true
	cfg.NotificationsEnabled = true
	cfg.AdminEmails = "admin@example.com,admin-only@example.com"

	queries := &db.QuerierStub{
		SystemGetUserByEmailFn: func(_ context.Context, addr string) (*db.SystemGetUserByEmailRow, error) {
			switch addr {
			case "admin@example.com":
				return &db.SystemGetUserByEmailRow{Idusers: 7, Email: addr}, nil
			case "admin-only@example.com":
				return &db.SystemGetUserByEmailRow{Idusers: 9, Email: addr}, nil
			default:
				return nil, sql.ErrNoRows
			}
		},
		ListSubscribersForPatternsFn: func(_ context.Context, arg db.ListSubscribersForPatternsParams) ([]int32, error) {
			if arg.Method != "internal" {
				return nil, nil
			}
			for _, pattern := range arg.Patterns {
				if pattern == "notify:/admin/*" {
					// Admin 9 does not subscribe to ordinary replies.
					return []int32{7, 9}, nil
				}
			}
			// Admin 7 also subscribes, ordinary subscriber 8 receives the same
			// reply, and multiple effective matches for 7 must not duplicate.
			return []int32{7, 7, 8}, nil
		},
		SystemGetTemplateOverrideFn: func(_ context.Context, name string) (string, error) {
			if name == "overlap-admin.gotxt" {
				return "admin version", nil
			}
			if name == "overlap-subscriber.gotxt" {
				return "subscriber version", nil
			}
			return "", sql.ErrNoRows
		},
	}

	notifier := New(WithQueries(queries), WithConfig(cfg))
	event := eventbus.TaskEvent{
		Path:       "/forum/topic/1/thread/2",
		Task:       overlapReplyTask{TaskString: tasks.TaskString("Reply")},
		UserID:     1,
		ActivityID: "comment:42",
		Data:       map[string]any{"URL": "/forum/topic/1/thread/2#c42"},
		Outcome:    eventbus.TaskOutcomeSuccess,
	}
	if err := notifier.ProcessEvent(context.Background(), event, nil); err != nil {
		t.Fatalf("ProcessEvent: %v", err)
	}

	calls := queries.SystemCreateNotificationForActivityCalls
	if len(calls) != 4 {
		t.Fatalf("activity contributions = %d, want 4 (two admin + two subscriber): %+v", len(calls), calls)
	}
	if len(queries.SystemCreateNotificationCalls) != 0 {
		t.Fatalf("unexpected non-idempotent notification inserts: %+v", queries.SystemCreateNotificationCalls)
	}

	admins := map[int32]bool{}
	for _, admin := range calls[:2] {
		if admin.Priority != notificationPriorityAdmin ||
			admin.Message.String != "admin version" || admin.ActivityKey.String != "reply:comment:42" {
			t.Fatalf("admin contribution = %+v", admin)
		}
		admins[admin.RecipientID] = true
	}
	if len(admins) != 2 || !admins[7] || !admins[9] {
		t.Fatalf("admin recipients = %+v, want overlapping admin 7 and admin-only 9", admins)
	}
	subscribers := map[int32]bool{}
	for _, call := range calls[2:] {
		if call.Priority != notificationPrioritySubscriber ||
			call.Message.String != "subscriber version" ||
			call.ActivityKey.String != "reply:comment:42" ||
			call.Link.String != "/forum/topic/1/thread/2#c42" {
			t.Fatalf("subscriber contribution = %+v", call)
		}
		subscribers[call.RecipientID] = true
	}
	if len(subscribers) != 2 || !subscribers[7] || !subscribers[8] {
		t.Fatalf("subscriber recipients = %+v, want users 7 and 8", subscribers)
	}

	// No source is suppressed; the DB's (recipient, activity) unique key
	// consolidates overlapping contributions while preserving admin-only 9 and subscriber-only 8.
}
