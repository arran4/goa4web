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
	cfg.AdminEmails = "admin@example.com"

	queries := &db.QuerierStub{
		SystemGetUserByEmailRow: &db.SystemGetUserByEmailRow{Idusers: 7, Email: "admin@example.com"},
		ListSubscribersForPatternsFn: func(_ context.Context, arg db.ListSubscribersForPatternsParams) ([]int32, error) {
			if arg.Method == "internal" {
				// User 7 is an administrator with overlapping reply rules;
				// user 8 is only a subscriber. Duplicate matches are a set.
				return []int32{7, 7, 8}, nil
			}
			return nil, nil
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
	if len(calls) != 3 {
		t.Fatalf("activity contributions = %d, want 3 (admin + two subscriptions): %+v", len(calls), calls)
	}
	if len(queries.SystemCreateNotificationCalls) != 0 {
		t.Fatalf("unexpected non-idempotent notification inserts: %+v", queries.SystemCreateNotificationCalls)
	}

	admin := calls[0]
	if admin.RecipientID != 7 || admin.Priority != notificationPriorityAdmin ||
		admin.Message.String != "admin version" || admin.ActivityKey.String != "reply:comment:42" {
		t.Fatalf("admin contribution = %+v", admin)
	}
	subscribers := map[int32]bool{}
	for _, call := range calls[1:] {
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
	// consolidates overlapping contributions while preserving recipient 8.
}
