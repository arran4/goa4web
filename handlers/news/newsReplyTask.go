package news

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

type ReplyTask struct{ tasks.TaskString }

// ReplyTask sends notifications and auto-subscribes authors and followers when someone replies to a news post.
var (
	replyTask = &ReplyTask{TaskString: TaskReply}

	_ tasks.Task                                    = (*ReplyTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*ReplyTask)(nil)
	_ notif.AutoSubscribeProvider                   = (*ReplyTask)(nil)
	_ tasks.EmailTemplatesRequired                  = (*ReplyTask)(nil)
)

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateNewsReply.EmailTemplates(), evt.Outcome == eventbus.TaskOutcomeSuccess
}

func (ReplyTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	s := NotificationTemplateNewsReply.NotificationTemplate()
	return &s
}

func (ReplyTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationNewsReply.EmailTemplates(), evt.Outcome == eventbus.TaskOutcomeSuccess
}

func (ReplyTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	v := EmailTemplateAdminNotificationNewsReply.NotificationTemplate()
	return &v
}

func (ReplyTask) EmailTemplatesRequired() []tasks.Page {
	return append(EmailTemplateNewsReply.RequiredPages(), EmailTemplateAdminNotificationNewsReply.RequiredPages()...)
}

// AutoSubscribePath registers this reply so the author automatically follows subsequent comments on the news post.
// When users reply to a news post we automatically subscribe them so they receive updates to the thread they just engaged with.
// AutoSubscribePath allows commenters to automatically watch for further replies.
// AutoSubscribePath implements notif.AutoSubscribeProvider. A subscription to the underlying discussion thread is created using event data when available.
func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskReply), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return string(TaskReply), evt.Path, nil
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}

	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["news"])
	if err != nil {
		return fmt.Errorf("post id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if pid == 0 {
		return fmt.Errorf("no bid %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("No bid")))
	}

	uid, _ := session.Values["UID"].(int32)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("news", "post", "reply", int32(pid)) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return nil
	}

	text := r.PostFormValue("replytext")
	languageID, _ := strconv.Atoi(r.PostFormValue("language"))

	cid, ti, err := cd.CreateNewsReply(uid, int32(pid), int32(languageID), text)
	if err != nil {
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	endURL := cd.AbsoluteURL(fmt.Sprintf("/news/news/%d", pid))
	username := ""
	if user, err := cd.CurrentUser(); err == nil && user != nil {
		username = user.Username.String
	}
	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             ti.ThreadID,
		TopicID:              ti.TopicID,
		CommentID:            int32(cid),
		LabelItem:            "news",
		LabelItemID:          int32(pid),
		CommentText:          text,
		CommentURL:           endURL,
		PostURL:              endURL,
		Username:             username,
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        true,
	}); err != nil {
		log.Printf("news reply side effects: %v", err)
	}

	return nil
}
