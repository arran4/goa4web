package privateforum

import (
	"testing"

	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestPrivateTopicCreateTaskAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(privateTopicCreateTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("PrivateTopicCreateTask should implement AutoSubscribeProvider so creators follow replies")
	}
}
