package imagebbs

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

// Ensure reply tasks auto subscribe commenters.
func TestReplyTaskAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(replyTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask should implement AutoSubscribeProvider so commenters get updates")
	}
}
