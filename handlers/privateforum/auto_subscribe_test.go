package privateforum

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestHappyPathPrivateTopicCreateTaskAutoSubscribe(t *testing.T) {
	if _, ok := any(privateTopicCreateTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("PrivateTopicCreateTask should implement AutoSubscribeProvider so creators follow replies")
	}
}

func TestHappyPathPrivateTopicCreateTaskAutoSubscribePath(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			topicCreateEventKey: topicCreateEventData{TopicID: 88, BasePath: "/private"},
		},
		Path: "/private/topic/new",
	}

	actionName, path, err := privateTopicCreateTask.AutoSubscribePath(evt)
	if err != nil {
		t.Fatalf("AutoSubscribePath error: %v", err)
	}
	if actionName != string(TaskPrivateTopicCreate) {
		t.Fatalf("expected action name %q, got %q", TaskPrivateTopicCreate, actionName)
	}
	if path != "/private/topic/88" {
		t.Fatalf("expected auto-subscribe path /private/topic/88, got %q", path)
	}
}
