package news

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

// Test tasks that should auto subscribe implement the interface so
// readers get updates when threads continue.
func TestHappyPathNewsAutoSubscribeTasks(t *testing.T) {
	// group tasks under test for easy extension
	tests := []struct {
		name string
		task any
	}{
		{"ReplyTask", replyTask},
		{"NewPostTask", newPostTask},
	}
	for _, tt := range tests {
		if _, ok := tt.task.(notif.AutoSubscribeProvider); !ok {
			t.Fatalf("%s should implement AutoSubscribeProvider so users are notified about replies", tt.name)
		}
	}
}
