package news

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestHappyPathNewsReplyAutoSubscribe(t *testing.T) {
	if _, ok := any(replyTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask must auto subscribe so commenters are notified of responses")
	}
}

func TestHappyPathNewsPostAutoSubscribe(t *testing.T) {
	if _, ok := any(newPostTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("NewPostTask must auto subscribe so authors follow replies")
	}
}
