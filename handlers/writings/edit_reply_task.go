package writings

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil {
		return fmt.Errorf("load writing fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if writing == nil {
		return fmt.Errorf("load writing fail %w", handlers.ErrRedirectOnSamePageHandler(sql.ErrNoRows))
	}
	comment, err := cd.ArticleComment(r)
	if err != nil || comment == nil {
		return fmt.Errorf("load comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return handlers.SessionFetchFail{}
	}

	thread, err := cd.UpdateWritingReply(comment.Idcomments, int32(languageID), text)
	if err != nil {
		return fmt.Errorf("update comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             thread.Idforumthread,
		TopicID:              thread.ForumtopicIdforumtopic,
		CommentID:            comment.Idcomments,
		LabelItem:            "writing",
		LabelItemID:          writing.Idwriting,
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
	}); err != nil {
		log.Printf("writing comment edit side effects: %v", err)
	}

	return handlers.RedirectHandler(fmt.Sprintf("/writings/article/%d", writing.Idwriting))
}

func (EditReplyTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationNewsCommentEditEmail"), true
}

func (EditReplyTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsCommentEditEmail")
	return &v
}
