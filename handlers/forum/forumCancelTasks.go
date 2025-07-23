package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/gorilla/mux"
)

// threadNewCancelTask aborts creating a new thread.
type threadNewCancelTask struct{ tasks.TaskString }

var threadNewCancelAction = &threadNewCancelTask{TaskString: TaskCancel}

var _ tasks.Task = (*threadNewCancelTask)(nil)

func (threadNewCancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	topicID, _ := strconv.Atoi(mux.Vars(r)["topic"])
	return handlers.RedirectHandler(fmt.Sprintf("/forum/topic/%d", topicID))
}

// topicThreadReplyCancelTask cancels replying to a thread.
type topicThreadReplyCancelTask struct{ tasks.TaskString }

var topicThreadReplyCancel = &topicThreadReplyCancelTask{TaskString: TaskCancel}

var _ tasks.Task = (*topicThreadReplyCancelTask)(nil)

func (topicThreadReplyCancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	threadRow := r.Context().Value(consts.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(consts.KeyTopic).(*db.GetForumTopicByIdForUserRow)
	endURL := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)
	return handlers.RedirectHandler(endURL)
}

// topicThreadCommentEditActionTask updates a comment and refreshes thread metadata.
type topicThreadCommentEditActionTask struct{ tasks.TaskString }

var topicThreadCommentEditAction = &topicThreadCommentEditActionTask{TaskString: TaskEditReply}

var _ tasks.Task = (*topicThreadCommentEditActionTask)(nil)

func (topicThreadCommentEditActionTask) Action(w http.ResponseWriter, r *http.Request) any {
	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	threadRow := r.Context().Value(consts.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(consts.KeyTopic).(*db.GetForumTopicByIdForUserRow)
	commentID, _ := strconv.Atoi(mux.Vars(r)["comment"])

	if err = queries.UpdateComment(r.Context(), db.UpdateCommentParams{
		Idcomments:         int32(commentID),
		LanguageIdlanguage: int32(languageID),
		Text:               sql.NullString{String: text, Valid: true},
	}); err != nil {
		return fmt.Errorf("update comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: threadRow.Idforumthread, TopicID: topicRow.Idforumtopic}
		}
	}

	return handlers.RedirectHandler(fmt.Sprintf("/forum/topic/%d/thread/%d#comment-%d", topicRow.Idforumtopic, threadRow.Idforumthread, commentID))
}

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
