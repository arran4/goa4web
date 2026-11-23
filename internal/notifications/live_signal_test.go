package notifications

import (
	"errors"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

func TestNotificationLiveSignalFollowsPersistence(t *testing.T) {
	for _, tc := range []struct {
		name     string
		activity string
	}{
		{name: "ordinary notification"},
		{name: "activity notification", activity: "comment:42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := &db.QuerierStub{}
			bus := eventbus.NewBus()
			var signals []eventbus.NotificationsChanged
			bus.SyncPublish = func(msg eventbus.Message) {
				evt, ok := msg.(eventbus.NotificationsChanged)
				if !ok {
					t.Fatalf("unexpected notification signal %T", msg)
				}
				if len(q.SystemCreateNotificationCalls) != 1 ||
					(tc.activity != "" && len(q.SystemCreateNotificationForActivityCalls) != 1) {
					t.Fatal("signal was published before persistence")
				}
				signals = append(signals, evt)
			}
			notifier := &Notifier{Queries: q, Bus: bus}
			const recipient = 15 // Owner of the stored notification.
			action := eventbus.TaskEvent{ActivityID: tc.activity, UserID: 99}
			if err := notifier.sendInternalNotificationForEvent(t.Context(), action, recipient, "/private", "reserved text", 20); err != nil {
				t.Fatal(err)
			}
			if len(signals) != 1 || signals[0].RecipientID != recipient {
				t.Fatalf("signals = %#v", signals)
			}
			failure := errors.New("write failed")
			q.SystemCreateNotificationErr = failure
			q.SystemCreateNotificationForActivityErr = failure
			if err := notifier.sendInternalNotificationForEvent(t.Context(), action, recipient, "/private", "reserved text", 20); !errors.Is(err, failure) {
				t.Fatalf("notification error = %v, want %v", err, failure)
			}
			if len(signals) != 1 {
				t.Fatal("failed notification emitted a signal")
			}
		})
	}
}
