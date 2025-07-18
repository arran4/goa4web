package news

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	postcountworker "github.com/arran4/goa4web/workers/postcountworker"
)

// EditReplyTask updates an existing news comment.
type EditReplyTask struct{ tasks.TaskString }

var editReplyTask = &EditReplyTask{TaskString: TaskEditReply}

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])
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
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err = queries.UpdateComment(r.Context(), db.UpdateCommentParams{
		Idcomments:         int32(commentId),
		LanguageIdlanguage: int32(languageId),
		Text:               sql.NullString{String: text, Valid: true},
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

	http.Redirect(w, r, fmt.Sprintf("/news/news/%d", postId), http.StatusTemporaryRedirect)
}

// CancelTask aborts comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["post"])
	http.Redirect(w, r, fmt.Sprintf("/news/news/%d", postId), http.StatusTemporaryRedirect)
}
