package blogs

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestReplyBlogTaskAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(replyBlogTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyBlogTask must auto subscribe as users want comment updates")
	}
}
