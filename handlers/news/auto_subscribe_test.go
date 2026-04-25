package news

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestHappyPathNewsReplyAutoSubscribe(t *testing.T) {
	if !notif.HasAutoSubscribe(replyTask) {
		t.Fatalf("ReplyTask must auto subscribe so commenters are notified of responses")
	}
}

func TestHappyPathNewsPostAutoSubscribe(t *testing.T) {
	if !notif.HasAutoSubscribe(newPostTask) {
		t.Fatalf("NewPostTask must auto subscribe so authors follow replies")
	}
}
