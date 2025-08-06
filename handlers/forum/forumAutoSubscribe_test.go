package forum

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

// Forum participants expect updates when threads they interact with change.
func TestForumAutoSubscribeTasks(t *testing.T) {
	if _, ok := interface{}(replyTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask should implement AutoSubscribeProvider so users get notified about thread replies")
	}
	if _, ok := interface{}(createThreadTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("CreateThreadTask should implement AutoSubscribeProvider so thread authors follow their threads")
	}
}
