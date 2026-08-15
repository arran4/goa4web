package forum

import (
	"testing"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/workers/postcountworker"
)

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
