package forum

import (
	"testing"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
)

func TestPrivateForumTasksRequireThreadViewGrantForSubscriberDelivery(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{ThreadID: 55, TopicID: 44},
		},
		Path: "/private/topic/44/thread/55",
	}
	for name, task := range map[string]any{
		"create thread": createThreadTask,
		"reply":         replyTask,
	} {
		provider, ok := task.(notif.GrantsRequiredProvider)
		if !ok {
			t.Errorf("%s task does not filter subscriber delivery by grants", name)
			continue
		}
		requirements, err := provider.GrantsRequired(evt)
		if err != nil {
			t.Errorf("%s GrantsRequired: %v", name, err)
			continue
		}
		if len(requirements) != 1 {
			t.Errorf("%s grant requirements = %d, want 1", name, len(requirements))
			continue
		}
		want := notif.GrantRequirement{
			Section: consts.PermissionSectionPrivateForumThread,
			Item:    consts.PermissionItemThread,
			ItemID:  55,
			Action:  consts.PermissionActionView,
		}
		if requirements[0] != want {
			t.Errorf("%s grant requirement = %+v, want %+v", name, requirements[0], want)
		}
	}
}

func TestHappyPathCreateThreadTaskAutoSubscribeGrants(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID:  55,
				TopicID:   44,
				CommentID: 777,
			},
		},
		Path: "/forum/topic/44/thread",
	}
	reqs, err := createThreadTask.AutoSubscribeGrants(evt)
	if err != nil {
		t.Fatalf("AutoSubscribeGrants error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 grant requirement, got %d", len(reqs))
	}
	req := reqs[0]
	if req.Section != consts.PermissionSectionForum || req.Item != consts.PermissionItemThread || req.ItemID != 55 || req.Action != consts.PermissionActionView {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}

func TestCreateThreadTaskAutoSubscribeGrantsPrivateThread(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID: 55,
				TopicID:  44,
			},
		},
		Path: "/private/topic/44/thread",
	}
	reqs, err := createThreadTask.AutoSubscribeGrants(evt)
	if err != nil {
		t.Fatalf("AutoSubscribeGrants error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 grant requirement, got %d", len(reqs))
	}
	if req := reqs[0]; req.Section != consts.PermissionSectionPrivateForumThread || req.Item != consts.PermissionItemThread || req.ItemID != 55 || req.Action != consts.PermissionActionView {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}

func TestHappyPathReplyTaskAutoSubscribeGrants(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID:  77,
				TopicID:   88,
				CommentID: 999,
			},
		},
		Path: "/forum/topic/88/thread/77/reply",
	}
	reqs, err := replyTask.AutoSubscribeGrants(evt)
	if err != nil {
		t.Fatalf("AutoSubscribeGrants error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 grant requirement, got %d", len(reqs))
	}
	req := reqs[0]
	if req.Section != consts.PermissionSectionForum || req.Item != consts.PermissionItemThread || req.ItemID != 77 || req.Action != consts.PermissionActionView {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}

func TestReplyTaskAutoSubscribeGrantsPrivateThread(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID: 77,
				TopicID:  88,
			},
		},
		Path: "/private/topic/88/thread/77/reply",
	}
	reqs, err := replyTask.AutoSubscribeGrants(evt)
	if err != nil {
		t.Fatalf("AutoSubscribeGrants error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 grant requirement, got %d", len(reqs))
	}
	if req := reqs[0]; req.Section != consts.PermissionSectionPrivateForumThread || req.Item != consts.PermissionItemThread || req.ItemID != 77 || req.Action != consts.PermissionActionView {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}
