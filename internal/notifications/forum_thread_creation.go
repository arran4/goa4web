package notifications

import (
	"context"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

func isForumThreadCreation(evt eventbus.TaskEvent) bool {
	task, named := evt.Task.(tasks.Name)
	return named && strings.EqualFold(task.Name(), "Create Thread") &&
		(strings.HasPrefix(evt.Path, "/forum/") || strings.HasPrefix(evt.Path, "/private/"))
}

func (n *Notifier) canReceiveForumThreadCreation(ctx context.Context, recipientID int32, evt eventbus.TaskEvent) bool {
	commentID, _ := common.ActivityPositionCommentAndAppend(evt.ActivityPosition)
	if evt.ResourceType != "thread" || evt.ResourceID <= 0 || commentID <= 0 {
		return false
	}
	viewer := common.NewCoreData(ctx, n.Queries, n.Config, common.WithUserID(recipientID))
	return viewer.CanReceiveForumReply(0, evt.ResourceID, commentID)
}
