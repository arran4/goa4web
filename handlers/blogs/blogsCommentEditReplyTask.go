package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// EditReplyTask updates an existing comment.
type EditReplyTask struct{ tasks.TaskString }

var editReplyTask = &EditReplyTask{TaskString: TaskEditReply}

var _ tasks.Task = (*EditReplyTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*EditReplyTask)(nil)
var _ tasks.EmailTemplatesRequired = (*EditReplyTask)(nil)

func (EditReplyTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationBlogCommentEdit.EmailTemplates(), true
}

func (EditReplyTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationBlogCommentEdit.NotificationTemplate()
	return &v
}

func (EditReplyTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminNotificationBlogCommentEdit.RequiredTemplates()
}

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("replytext")

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])
	commentId, _ := strconv.Atoi(vars["comment"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
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

	if err = cd.UpdateBlogReply(int32(commentId), uid, int32(languageId), text); err != nil {
		return fmt.Errorf("update comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             thread.Idforumthread,
		TopicID:              thread.ForumtopicIdforumtopic,
		CommentID:            int32(commentId),
		LabelItem:            "blog",
		LabelItemID:          int32(blogId),
		CommentURL:           cd.AbsoluteURL(fmt.Sprintf("/blogs/blog/%d/comments", blogId)),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
	}); err != nil {
		log.Printf("blog comment edit side effects: %v", err)
	}

	return handlers.RedirectHandler(fmt.Sprintf("/blogs/blog/%d/comments", blogId))
}
