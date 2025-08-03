package forum

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/gorilla/mux"
)

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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	threadRow, err := cd.CurrentThread()
	if err != nil || threadRow == nil {
		return fmt.Errorf("thread fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		return fmt.Errorf("topic fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	commentID, _ := strconv.Atoi(mux.Vars(r)["comment"])

	if err = queries.UpdateCommentForCommenter(r.Context(), db.UpdateCommentForCommenterParams{
		CommentID:      int32(commentID),
		GrantCommentID: sql.NullInt32{Int32: int32(commentID), Valid: true},
		LanguageID:     int32(languageID),
		Text:           sql.NullString{String: text, Valid: true},
		GranteeID:      sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		CommenterID:    cd.UserID,
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
