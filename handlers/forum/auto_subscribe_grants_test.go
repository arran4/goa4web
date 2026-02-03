package forum

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/workers/postcountworker"
)

func TestCreateThreadTaskAutoSubscribeGrants(t *testing.T) {
	evt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID:  55,
				TopicID:   44,
				CommentID: 777,
			},
		},
		Path: "/forum/topic/44/thread/new",
	}
	reqs, err := createThreadTask.AutoSubscribeGrants(evt)
	if err != nil {
		t.Fatalf("AutoSubscribeGrants error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 grant requirement, got %d", len(reqs))
	}
	req := reqs[0]
	if req.Section != "forum" || req.Item != "thread" || req.ItemID != 55 || req.Action != "view" {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}

func TestReplyTaskAutoSubscribeGrants(t *testing.T) {
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
	if req.Section != "forum" || req.Item != "thread" || req.ItemID != 77 || req.Action != "view" {
		t.Errorf("unexpected grant requirement: %+v", req)
	}
}
