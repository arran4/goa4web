package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)

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

	if err := cd.ValidateCodeImagesForThread(cd.UserID, thread.Idforumthread, text); err != nil {
		return fmt.Errorf("validate images: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err = queries.UpdateCommentForEditor(r.Context(), db.UpdateCommentForEditorParams{
		LanguageID: sql.NullInt32{Int32: int32(languageId), Valid: languageId != 0},
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
		CommentID:   int32(commentId),
		CommenterID: cd.UserID,
		EditorID:    sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	}); err != nil {
		return fmt.Errorf("update comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             thread.Idforumthread,
		TopicID:              thread.ForumtopicIdforumtopic,
		CommentID:            int32(commentId),
		LabelItem:            "link",
		LabelItemID:          int32(linkId),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
	}); err != nil {
		log.Printf("linker comment edit side effects: %v", err)
	}
	if err := cd.RecordThreadImages(thread.Idforumthread, text); err != nil {
		log.Printf("record thread images: %v", err)
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/linker/comments/%d", linkId)}
}
