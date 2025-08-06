package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/gorilla/mux"
)

// EditReplyTask posts an edited reply and refreshes thread metadata.
type EditReplyTask struct{ tasks.TaskString }

var commentEditAction = &EditReplyTask{TaskString: TaskEditReply}
var _ tasks.Task = (*EditReplyTask)(nil)

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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	comment := cd.CurrentCommentLoaded()
	if comment == nil {
		var err error
		comment, err = cd.CommentByID(int32(commentId))
		if err != nil {
			return fmt.Errorf("load comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

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

	if err = queries.UpdateCommentForCommenter(r.Context(), db.UpdateCommentForCommenterParams{
		CommentID:      int32(commentId),
		GrantCommentID: sql.NullInt32{Int32: int32(commentId), Valid: true},
		LanguageID:     int32(languageId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
		GranteeID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		CommenterID: cd.UserID,
	}); err != nil {
		return fmt.Errorf("update comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(commentId), ThreadID: thread.Idforumthread, TopicID: thread.ForumtopicIdforumtopic}
		}
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/linker/comments/%d", linkId)}
}
