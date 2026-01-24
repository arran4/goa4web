package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// topicThreadCommentEditActionCancelTask aborts editing a comment.
type topicThreadCommentEditActionCancelTask struct{ tasks.TaskString }

var topicThreadCommentEditActionCancel = &topicThreadCommentEditActionCancelTask{TaskString: forumcommon.TaskCancel}

// TopicThreadCommentEditActionCancelHandler aborts editing a comment. Exported for reuse.
var TopicThreadCommentEditActionCancelHandler = topicThreadCommentEditActionCancel

var _ tasks.Task = (*topicThreadCommentEditActionCancelTask)(nil)

func (topicThreadCommentEditActionCancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		return fmt.Errorf("thread fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		return fmt.Errorf("topic fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	endURL := fmt.Sprintf("%s/topic/%d/thread/%d#bottom", base, topicRow.Idforumtopic, threadRow.Idforumthread)
	return handlers.RedirectHandler(endURL)
}
