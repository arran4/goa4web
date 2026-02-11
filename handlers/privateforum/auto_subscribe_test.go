package privateforum

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
)

func TestPrivateTopicCreateTaskAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(privateTopicCreateTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("PrivateTopicCreateTask should implement AutoSubscribeProvider so creators follow replies")
	}
}

func TestPrivateTopicCreateTaskAutoSubscribePath(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID:  7,
				TopicID:   88,
				CommentID: 999,
			},
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
