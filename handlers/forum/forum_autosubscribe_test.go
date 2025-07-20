package forum

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestCreateThreadTaskImplementsAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(createThreadTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("CreateThreadTask must auto subscribe as users will want updates")
	}
}

func TestReplyTaskImplementsAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(replyTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask must auto subscribe as users will want updates")
	}
}
