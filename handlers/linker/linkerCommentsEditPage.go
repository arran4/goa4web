package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/gorilla/mux"
)

// CommentEditActionPage updates a comment then refreshes thread metadata.
type EditReplyTask struct{ tasks.TaskString }

var commentEditAction = &EditReplyTask{TaskString: TaskEditReply}
var _ tasks.Task = (*EditReplyTask)(nil)

func (t EditReplyTask) Page(w http.ResponseWriter, r *http.Request) {
	t.Action(w, r)
}

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	commentId, _ := strconv.Atoi(vars["comment"])

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	comment := r.Context().Value(consts.KeyComment).(*db.GetCommentByIdForUserRow)

	thread, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      comment.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			return fmt.Errorf("thread lookup fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	if err = queries.UpdateComment(r.Context(), db.UpdateCommentParams{
		Idcomments:         int32(commentId),
		LanguageIdlanguage: int32(languageId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	}); err != nil {
		return fmt.Errorf("update comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: thread.Idforumthread, TopicID: thread.ForumtopicIdforumtopic}
		}
	}

	return handlers.RedirectHandler(fmt.Sprintf("/linker/comments/%d", linkId))
}

// CommentEditActionCancelPage aborts editing a comment.
func CommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	http.Redirect(w, r, fmt.Sprintf("/linker/comments/%d", linkId), http.StatusTemporaryRedirect)
}

type cancelEditReplyTask struct{ tasks.TaskString }

var commentEditActionCancel = &cancelEditReplyTask{TaskString: TaskCancel}
var _ tasks.Task = (*cancelEditReplyTask)(nil)

func (cancelEditReplyTask) Page(w http.ResponseWriter, r *http.Request) {
	CommentEditActionCancelPage(w, r)
}
