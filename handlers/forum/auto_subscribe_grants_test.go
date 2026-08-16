package forum

import (
	"testing"

<<<<<<< HEAD
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
)

func TestHappyPathPrivateForumTasksRequireParentAndThreadViewGrantsForSubscriberDelivery(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{ThreadID: 55, TopicID: 44},
		},
		Path: "/private/topic/44/thread/55",
	}
	tests := []struct {
		name string
		task any
	}{
		{name: "create thread", task: createThreadTask},
		{name: "reply", task: replyTask},
	}
	want := []notif.GrantRequirement{
		{
			Section: consts.PermissionSectionPrivateForum,
			Item:    consts.PermissionItemTopic,
			ItemID:  44,
			Action:  consts.PermissionActionView,
		},
		{
			Section: consts.PermissionSectionPrivateForumThread,
			Item:    consts.PermissionItemThread,
			ItemID:  55,
			Action:  consts.PermissionActionView,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider, ok := test.task.(notif.GrantsRequiredProvider)
			if !ok {
				t.Fatalf("task does not filter subscriber delivery by grants")
			}
			requirements, err := provider.GrantsRequired(evt)
			if err != nil {
				t.Fatalf("GrantsRequired: %v", err)
			}
			if len(requirements) != len(want) {
				t.Fatalf("grant requirements = %d, want %d", len(requirements), len(want))
			}
			for i := range want {
				if requirements[i] != want[i] {
					t.Errorf("grant requirement %d = %+v, want %+v", i, requirements[i], want[i])
				}
			}
		})
	}
}

=======
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/workers/postcountworker"
)

>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
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
<<<<<<< HEAD
	if req.Section != consts.PermissionSectionForum || req.Item != consts.PermissionItemThread || req.ItemID != 55 || req.Action != consts.PermissionActionView {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}

func TestHappyPathCreateThreadTaskAutoSubscribeGrantsPrivateThread(t *testing.T) {
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
=======
	if req.Section != "forum" || req.Item != "thread" || req.ItemID != 55 || req.Action != "view" {
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
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
<<<<<<< HEAD
	if req.Section != consts.PermissionSectionForum || req.Item != consts.PermissionItemThread || req.ItemID != 77 || req.Action != consts.PermissionActionView {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}

func TestHappyPathReplyTaskAutoSubscribeGrantsPrivateThread(t *testing.T) {
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
=======
	if req.Section != "forum" || req.Item != "thread" || req.ItemID != 77 || req.Action != "view" {
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}
