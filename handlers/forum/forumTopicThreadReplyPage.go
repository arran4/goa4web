package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

const (
	EmailTemplateForumReply                  notif.EmailTemplateName        = "forumReplyEmail"
	NotificationTemplateForumReply           notif.NotificationTemplateName = "reply"
	EmailTemplateAdminNotificationForumReply notif.EmailTemplateName        = "adminNotificationForumReplyEmail"
)

// ReplyTask handles replying to an existing thread.
type ReplyTask struct{ tasks.TaskString }

// compile-time assertions that ReplyTask provides notifications, indexing and
// auto-subscription for thread replies.
var (
	replyTask = &ReplyTask{TaskString: TaskReply}

	// ReplyTaskHandler exposes the reply task for registration on other routes.
	ReplyTaskHandler = replyTask

	_ tasks.Task                                    = (*ReplyTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*ReplyTask)(nil)
	_ notif.AutoSubscribeProvider                   = (*ReplyTask)(nil)
	_ tasks.EmailTemplatesRequired                  = (*ReplyTask)(nil)
	_ searchworker.IndexedTask                      = ReplyTask{}
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
	return EmailTemplateForumReply.EmailTemplates(), evt.Outcome == eventbus.TaskOutcomeSuccess
}

func (ReplyTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	s := NotificationTemplateForumReply.NotificationTemplate()
	return &s
}

func (ReplyTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumReply.EmailTemplates(), evt.Outcome == eventbus.TaskOutcomeSuccess
}

func (ReplyTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	v := EmailTemplateAdminNotificationForumReply.NotificationTemplate()
	return &v
}

func (ReplyTask) RequiredTemplates() []tasks.Template {
	return append(EmailTemplateForumReply.RequiredTemplates(), EmailTemplateAdminNotificationForumReply.RequiredTemplates()...)
}

// AutoSubscribePath ensures authors automatically receive updates on replies.
// AutoSubscribePath implements notif.AutoSubscribeProvider. The subscription is
// created for the originating forum thread when that information is available.
func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		base := "/forum"
		if idx := strings.Index(evt.Path, "/topic/"); idx > 0 {
			base = evt.Path[:idx]
		}
		return string(TaskReply), fmt.Sprintf("%s/topic/%d/thread/%d", base, data.TopicID, data.ThreadID), nil
	}
	return string(TaskReply), evt.Path, nil
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Forum - Reply"
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		return fmt.Errorf("thread fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		return fmt.Errorf("topic fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	uid, _ := session.Values["UID"].(int32)
	var username string
	if u := cd.UserByID(uid); u != nil {
		username = u.Username.String
	}
	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	var cid int64
	if topicRow.Handler == "private" {
		cid, err = cd.CreatePrivateForumCommentForCommenter(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), text)
	} else {
		cid, err = cd.CreateForumCommentForCommenter(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), text)
	}
	if err != nil {
		log.Printf("Error: CreateComment: %s", err)
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		log.Printf("Error: CreateComment: %s", err)
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	anchor := fmt.Sprintf("c%d", cid)
	comments, err := cd.ThreadComments(threadRow.Idforumthread)
	if err != nil {
		log.Printf("Error fetching comments to determine index: %s", err)
	} else if len(comments) > 0 {
		anchor = fmt.Sprintf("c%d", len(comments))
	}

	endUrl := fmt.Sprintf("%s/topic/%d/thread/%d#%s", base, topicRow.Idforumtopic, threadRow.Idforumthread, anchor)

	data := map[string]any{}
	if firstPost, err := cd.CommentByID(threadRow.Firstpost); err == nil && firstPost != nil && firstPost.Text.Valid {
		data["ThreadOpenerPreview"] = a4code.SnipTextWords(firstPost.Text.String, 10)
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             threadRow.Idforumthread,
		TopicID:              topicRow.Idforumtopic,
		CommentID:            int32(cid),
		Thread:               threadRow,
		TopicTitle:           topicRow.Title.String,
		Username:             username,
		CommentText:          text,
		CommentURL:           cd.AbsoluteURL(endUrl),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        true,
		AdditionalData:       data,
	}); err != nil {
		log.Printf("thread reply side effects: %v", err)
	}

	return handlers.RedirectHandler(endUrl)
}

func TopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Reply"
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	endUrl := fmt.Sprintf("%s/topic/%d/thread/%d#bottom", base, topicRow.Idforumtopic, threadRow.Idforumthread)
	http.Redirect(w, r, endUrl, http.StatusSeeOther)
}
