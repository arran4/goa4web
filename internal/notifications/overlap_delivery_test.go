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

	const (
		actorID            int32 = 1
		overlappingAdminID int32 = 2
		ordinarySubID      int32 = 3
		nonMemberAdminID   int32 = 4
		revokedSubID       int32 = 5
		topicID            int32 = 5
		threadID           int32 = 42
	)

	queries := &db.QuerierStub{
		SystemGetUserByEmailFn: func(_ context.Context, addr string) (*db.SystemGetUserByEmailRow, error) {
			switch addr {
			case "admin@example.com":
				return &db.SystemGetUserByEmailRow{Idusers: overlappingAdminID, Email: addr}, nil
			case "admin-only@example.com":
				return &db.SystemGetUserByEmailRow{Idusers: nonMemberAdminID, Email: addr}, nil
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
					return []int32{overlappingAdminID, nonMemberAdminID}, nil
				}
			}
			// Actor 1, overlapping admin 2 (matched twice across patterns), ordinary subscriber 3, and revoked subscriber 5.
			return []int32{actorID, overlappingAdminID, overlappingAdminID, ordinarySubID, revokedSubID}, nil
		},
		SystemCheckGrantFn: func(arg db.SystemCheckGrantParams) (int32, error) {
			switch arg.UserID.Int32 {
			case actorID, overlappingAdminID, ordinarySubID:
				if (arg.Section == "privateforum" && arg.Item.String == "topic" && arg.ItemID.Int32 == topicID && arg.Action == "view") ||
					(arg.Section == "privateforum_thread" && arg.Item.String == "thread" && arg.ItemID.Int32 == threadID && arg.Action == "view") {
					return 1, nil
				}
			case revokedSubID:
				if arg.Section == "privateforum" && arg.Item.String == "topic" && arg.ItemID.Int32 == topicID && arg.Action == "view" {
					return 1, nil
				}
			}
			return 0, sql.ErrNoRows
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
		Path:       "/private/topic/5/thread/42",
		Task:       overlapReplyTask{TaskString: tasks.TaskString("Reply")},
		UserID:     actorID,
		ActivityID: "comment:42",
		Data: map[string]any{
			"Username":      "replier",
			"TopicTitle":    "Private Ops",
			"Text":          "Original thread opener",
			"CommentText":   "New reply preview body",
			"CommentID":     int32(42),
			"CommentAnchor": "c42",
			"URL":           "/private/topic/5/thread/42#c42",
		},
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	if err := notifier.ProcessEvent(context.Background(), event, nil); err != nil {
		t.Fatalf("ProcessEvent: %v", err)
	}

	calls := queries.SystemCreateNotificationForActivityCalls
	if len(calls) != 2 {
		t.Fatalf("activity contributions = %d, want 2 (overlapping admin+subscriber 2 and ordinary subscriber 3 only): %+v", len(calls), calls)
	}

	byRecipient := map[int32]db.SystemCreateNotificationForActivityParams{}
	for _, call := range calls {
		if call.Message.String == "admin version" {
			t.Fatalf("recipient %d received competing admin template %q instead of canonical reply notification", call.RecipientID, call.Message.String)
		}
		if call.ActivityKey.String != "reply:comment:42" {
			t.Fatalf("recipient %d activity key = %q, want reply:comment:42", call.RecipientID, call.ActivityKey.String)
		}
		if call.Link.String != "/private/topic/5/thread/42#c42" {
			t.Fatalf("recipient %d link = %q, want /private/topic/5/thread/42#c42", call.RecipientID, call.Link.String)
		}
		byRecipient[call.RecipientID] = call
	}

	if _, ok := byRecipient[overlappingAdminID]; !ok {
		t.Errorf("expected overlapping admin+subscriber (user %d) to receive canonical notification", overlappingAdminID)
	}
	if _, ok := byRecipient[ordinarySubID]; !ok {
		t.Errorf("expected ordinary subscriber (user %d) to receive canonical notification", ordinarySubID)
	}
	if _, ok := byRecipient[nonMemberAdminID]; ok {
		t.Errorf("expected non-participant admin (user %d) to receive zero notifications", nonMemberAdminID)
	}
	if _, ok := byRecipient[revokedSubID]; ok {
		t.Errorf("expected revoked-grant subscriber (user %d) to receive zero notifications", revokedSubID)
	}
	if _, ok := byRecipient[actorID]; ok {
		t.Errorf("expected actor (user %d) to receive zero self-notifications", actorID)
	}
	if byRecipient[overlappingAdminID].Message.String != byRecipient[ordinarySubID].Message.String {
		t.Errorf("expected identical canonical message for users %d and %d, got %q vs %q",
			overlappingAdminID, ordinarySubID,
			byRecipient[overlappingAdminID].Message.String,
			byRecipient[ordinarySubID].Message.String)
	}
}
