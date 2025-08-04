package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
)

// EditReplyTask updates an existing comment.
type EditReplyTask struct{ tasks.TaskString }

var editReplyTask = &EditReplyTask{TaskString: TaskEditReply}

var _ tasks.Task = (*EditReplyTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*EditReplyTask)(nil)

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	vars := mux.Vars(r)
	writingID, err := strconv.Atoi(vars["writing"])
	if err != nil {
		return fmt.Errorf("writing id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	commentID, err := strconv.Atoi(vars["comment"])
	if err != nil {
		return fmt.Errorf("comment id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	comment := cd.CurrentCommentLoaded()
	if comment == nil {
		var err error
		comment, err = cd.CommentByID(int32(commentID))
		if err != nil {
			return fmt.Errorf("load comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	thread, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      comment.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err = queries.UpdateCommentForCommenter(r.Context(), db.UpdateCommentForCommenterParams{
		CommentID:      int32(commentID),
		GrantCommentID: sql.NullInt32{Int32: int32(commentID), Valid: true},
		LanguageID:     int32(languageID),
		Text:           sql.NullString{String: text, Valid: true},
		GranteeID:      sql.NullInt32{Int32: uid, Valid: uid != 0},
		CommenterID:    uid,
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

	return handlers.RedirectHandler(fmt.Sprintf("/writings/article/%d", writingID))
}

func (EditReplyTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsCommentEditEmail")
}

func (EditReplyTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsCommentEditEmail")
	return &v
}
