package blogs

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestHappyPathReplyBlogTaskAutoSubscribe(t *testing.T) {
	if !notif.HasAutoSubscribe(replyBlogTask) {
		t.Fatalf("ReplyBlogTask must auto subscribe as users want comment updates")
	}
}
