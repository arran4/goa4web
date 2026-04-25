package imagebbs

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

// Test that replyTask auto subscribes commenters so they see responses.
func TestHappyPathReplyTaskAutoSubscribe(t *testing.T) {
	if !notif.HasAutoSubscribe(replyTask) {
		t.Fatalf("ReplyTask should implement AutoSubscribeProvider so commenters are notified about replies")
	}
}
