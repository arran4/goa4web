package forum

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/gorilla/mux"
)

// topicThreadCommentEditActionTask updates a comment and refreshes thread metadata.
type topicThreadCommentEditActionTask struct{ tasks.TaskString }

var topicThreadCommentEditAction = &topicThreadCommentEditActionTask{TaskString: TaskEditReply}

// TopicThreadCommentEditActionHandler updates a comment. Exported for reuse.
var TopicThreadCommentEditActionHandler = topicThreadCommentEditAction

var _ tasks.Task = (*topicThreadCommentEditActionTask)(nil)

func (topicThreadCommentEditActionTask) Action(w http.ResponseWriter, r *http.Request) any {
	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("replytext")

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
	commentID, _ := strconv.Atoi(mux.Vars(r)["comment"])

	if err = cd.UpdateForumComment(int32(commentID), int32(languageID), text); err != nil {
		return fmt.Errorf("update comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(commentID), ThreadID: threadRow.Idforumthread, TopicID: topicRow.Idforumtopic}
		}
	}

	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	return handlers.RedirectHandler(fmt.Sprintf("%s/topic/%d/thread/%d#comment-%d", base, topicRow.Idforumtopic, threadRow.Idforumthread, commentID))
}
