package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
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

func (ReplyTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

func (ReplyTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	return notif.NewEmailTemplates("adminNotificationForumReplyEmail")
}

func (ReplyTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumReplyEmail")
	return &v
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

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["TopicTitle"] = topicRow.Title.String
			evt.Data["ThreadID"] = threadRow.Idforumthread
			evt.Data["Thread"] = threadRow
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	endUrl := fmt.Sprintf("%s/topic/%d/thread/%d#bottom", base, topicRow.Idforumtopic, threadRow.Idforumthread)

	cid, err := cd.CreateForumCommentForCommenter(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), text)
	if err != nil {
		log.Printf("Error: CreateComment: %s", err)
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		log.Printf("Error: CreateComment: %s", err)
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.ClearTopicPrivateLabelStatus(topicRow.Idforumtopic); err != nil {
		log.Printf("clear label status: %v", err)
	}
	if err := cd.SetTopicPrivateLabelStatus(topicRow.Idforumtopic, false, false); err != nil {
		log.Printf("set label status: %v", err)
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(cid), ThreadID: threadRow.Idforumthread, TopicID: topicRow.Idforumtopic}
			evt.Data["CommentURL"] = cd.AbsoluteURL(endUrl)
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: text}
		}
	}

	return handlers.RedirectHandler(endUrl)
}

func TopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Reply"
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	endUrl := fmt.Sprintf("%s/topic/%d/thread/%d#bottom", base, topicRow.Idforumtopic, threadRow.Idforumthread)
	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
