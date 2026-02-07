package blogs

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

// Ensure reply tasks auto subscribe commenters so they know about further discussion.
func TestHappyPathBlogsAutoSubscribeTasks(t *testing.T) {
	// list tasks so future interface checks can reuse this structure
	tests := []struct {
		name string
		task any
	}{
		{"ReplyBlogTask", replyBlogTask},
	}
	for _, tt := range tests {
		if _, ok := tt.task.(notif.AutoSubscribeProvider); !ok {
			t.Fatalf("%s should implement AutoSubscribeProvider to notify commenters of new replies", tt.name)
		}
	}
}
