package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// topicThreadReplyCancelTask cancels replying to a thread.
type topicThreadReplyCancelTask struct{ tasks.TaskString }

var topicThreadReplyCancel = &topicThreadReplyCancelTask{TaskString: forumcommon.TaskCancel}

// TopicThreadReplyCancelHandler cancels replying to a thread. Exported for reuse.
var TopicThreadReplyCancelHandler = topicThreadReplyCancel

var _ tasks.Task = (*topicThreadReplyCancelTask)(nil)

func (topicThreadReplyCancelTask) Action(w http.ResponseWriter, r *http.Request) any {
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
