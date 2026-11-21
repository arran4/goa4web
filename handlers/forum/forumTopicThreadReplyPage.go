package forum

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

const (
	// EmailTemplateForumReply is the email template prefix for forum replies.
	EmailTemplateForumReply notif.EmailTemplateName = "forumReplyEmail"
	// NotificationTemplateForumReply is the internal notification template name for forum replies.
	NotificationTemplateForumReply notif.NotificationTemplateName = "reply"
)

// ReplyTask handles replying to an existing thread.
type ReplyTask struct{ tasks.TaskString }

// compile-time assertions that ReplyTask provides notifications, indexing and
// auto-subscription for thread replies.
var (
	replyTask = &ReplyTask{TaskString: TaskReply}
	// ReplyTaskHandler exposes the reply task for registration on other routes.
	ReplyTaskHandler                                               = replyTask
	_                tasks.Task                                    = (*ReplyTask)(nil)
	_                notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
	_                notif.AutoSubscribeProvider                   = (*ReplyTask)(nil)
	_                notif.GrantsRequiredProvider                  = (*ReplyTask)(nil)
	_                tasks.EmailTemplatesRequired                  = (*ReplyTask)(nil)
	_                searchworker.IndexedTask                      = ReplyTask{}
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
func (ReplyTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateForumReply.RequiredTemplates()
}

// AutoSubscribePath ensures authors automatically receive updates on replies.
// AutoSubscribePath implements notif.AutoSubscribeProvider. The subscription is
// created for the originating forum thread when that information is available.
func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	topicID, threadID, private := privateThreadIDs(evt)
	if topicID != 0 && threadID != 0 {
		base := common.CanonicalForumBasePath(private, evt.Path)
		return string(TaskReply), fmt.Sprintf("%s/topic/%d/thread/%d", base, topicID, threadID), nil
	}
	return string(TaskReply), evt.Path, nil
}
func (ReplyTask) AutoSubscribeGrants(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	_, threadID, private := privateThreadIDs(evt)
	if threadID == 0 {
		return nil, nil
	}
	if private {
		return privateThreadSubscriberGrants(evt)
	}
	return []notif.GrantRequirement{{Section: consts.PermissionSectionForum, Item: consts.PermissionItemThread, ItemID: threadID, Action: consts.PermissionActionView}}, nil
}

func (ReplyTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	return privateThreadSubscriberGrants(evt)
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Forum - Reply"
	uid, _ := session.Values["UID"].(int32)
	text := r.PostFormValue("replytext")
	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	lastDisplayedCommentID, _ := strconv.Atoi(r.PostFormValue("last_comment"))
	lastDisplayedActivity := r.PostFormValue("last_activity")
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	result, err := cd.ReplyForumThread(r.Context(), common.ReplyForumThreadParams{
		ActorID:                uid,
		TopicID:                cd.SelectedTopicID(),
		ThreadID:               cd.SelectedThreadID(),
		LastDisplayedCommentID: int32(lastDisplayedCommentID),
		LastDisplayedActivity:  lastDisplayedActivity,
		LanguageID:             int32(languageID),
		Text:                   text,
		Private:                base == "/private" || strings.HasPrefix(base, "/private"),
		EnforceHandler:         true,
		BasePath:               base,
	})
	if err != nil {
		var notFound common.ForumResourceNotFoundError
		var mismatch common.ForumHandlerMismatchError
		var forbidden common.ForumOperationForbiddenError
		switch {
		case errors.As(err, &notFound):
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("thread not found"))
			return nil
		case errors.As(err, &mismatch):
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("forum handler does not match topic"))
			return nil
		case errors.As(err, &forbidden):
			w.WriteHeader(http.StatusForbidden)
			handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
			return nil
		}
		log.Printf("Error: CreateComment: %s", err)
		message := fmt.Sprintf("Error creating comment: %v", err)
		var userError interface{ UserErrorMessage() string }
		if errors.As(err, &userError) && userError.UserErrorMessage() != "" {
			message = userError.UserErrorMessage()
		}
		cd.SetCurrentError(message)
		if r.Form == nil {
			r.Form = make(url.Values)
		}
		r.Form.Set("replytext", text)
		ThreadPageWithBasePath(w, r, base)
		return nil
	}
	if evt := cd.Event(); evt != nil {
		evt.Data["URL"] = cd.AbsoluteURL(result.URL)
	}
	return handlers.RedirectHandler(result.URL)
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
