package blogs

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestHappyPathReplyBlogTaskAutoSubscribe(t *testing.T) {
	if _, ok := any(replyBlogTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyBlogTask must auto subscribe as users want comment updates")
	}
}
