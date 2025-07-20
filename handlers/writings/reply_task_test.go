package writings

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestRepliesMustAutoSubscribe(t *testing.T) {
	var task ReplyTask
	if _, ok := interface{}(task).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("AutoSubscribeProvider must auto subscribe as users will want updates")
	}
}
