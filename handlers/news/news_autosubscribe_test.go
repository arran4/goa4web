package news

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestNewsTasksImplementAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(replyTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask must auto subscribe as users will want updates")
	}
	if _, ok := interface{}(newPostTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("NewPostTask must auto subscribe as users will want updates")
	}
}
