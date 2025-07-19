package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	postcountworker "github.com/arran4/goa4web/workers/postcountworker"
	"github.com/gorilla/mux"
)

// CommentEditActionPage updates a comment then refreshes thread metadata.
type EditReplyTask struct{ tasks.TaskString }

var _ tasks.Task = (*EditReplyTask)(nil)

var commentEditAction = &EditReplyTask{TaskString: TaskEditReply}

func (t EditReplyTask) Page(w http.ResponseWriter, r *http.Request) {
	t.Action(w, r)
}

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	commentId, _ := strconv.Atoi(vars["comment"])

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	comment := r.Context().Value(common.KeyComment).(*db.GetCommentByIdForUserRow)

	thread, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      comment.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: getThreadByIdForUserByIdWithLastPosterUserNameAndPermissions: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
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
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(common.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: thread.Idforumthread, TopicID: thread.ForumtopicIdforumtopic}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/linker/comments/%d", linkId), http.StatusTemporaryRedirect)
}

// CommentEditActionCancelPage aborts editing a comment.
func CommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	http.Redirect(w, r, fmt.Sprintf("/linker/comments/%d", linkId), http.StatusTemporaryRedirect)
}

type cancelEditReplyTask struct{ tasks.TaskString }

var _ tasks.Task = (*cancelEditReplyTask)(nil)

var commentEditActionCancel = &cancelEditReplyTask{TaskString: TaskCancel}

func (cancelEditReplyTask) Page(w http.ResponseWriter, r *http.Request) {
	CommentEditActionCancelPage(w, r)
}
