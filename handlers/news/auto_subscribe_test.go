package news

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestNewsReplyAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(replyTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask must auto subscribe so commenters are notified of responses")
	}
}

func TestNewsPostAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(newPostTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("NewPostTask must auto subscribe so authors follow replies")
	}
}
