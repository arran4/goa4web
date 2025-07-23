package forum

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// topicThreadCommentEditActionCancelTask aborts editing a comment.
type topicThreadCommentEditActionCancelTask struct{ tasks.TaskString }

var topicThreadCommentEditActionCancel = &topicThreadCommentEditActionCancelTask{TaskString: TaskCancel}

var _ tasks.Task = (*topicThreadCommentEditActionCancelTask)(nil)

func (topicThreadCommentEditActionCancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	threadRow := r.Context().Value(consts.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(consts.KeyTopic).(*db.GetForumTopicByIdForUserRow)
	endURL := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)
	return handlers.RedirectHandler(endURL)
}
