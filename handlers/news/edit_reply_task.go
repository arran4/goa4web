package news

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/gorilla/mux"

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

func (EditReplyTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationNewsCommentEditEmail"), true
}

func (EditReplyTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsCommentEditEmail")
	return &v
}

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("languageId parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("replytext")

	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["news"])
	commentId, _ := strconv.Atoi(vars["comment"])

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	ti, err := cd.UpdateNewsReply(int32(commentId), uid, int32(languageId), text)
	if err != nil {
		return fmt.Errorf("update comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:         ti.ThreadID,
		TopicID:          ti.TopicID,
		CommentID:        int32(commentId),
		CommentURL:       cd.AbsoluteURL(fmt.Sprintf("/news/news/%d", postId)),
		IncludePostCount: true,
	}); err != nil {
		log.Printf("news comment edit side effects: %v", err)
	}

	return handlers.RedirectHandler(fmt.Sprintf("/news/news/%d", postId))
}
