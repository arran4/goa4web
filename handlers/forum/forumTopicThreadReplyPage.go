package forum

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
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
	ReplyTaskHandler                                               = replyTask
	_                tasks.Task                                    = (*ReplyTask)(nil)
	_                notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
	_                notif.AdminEmailTemplateProvider              = (*ReplyTask)(nil)
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
func (ReplyTask) AutoSubscribeGrants(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		base := "/forum"
		if idx := strings.Index(evt.Path, "/topic/"); idx > 0 {
			base = evt.Path[:idx]
		}
		section := consts.PermissionSectionForum
		if base == "/private" {
			section = consts.PermissionSectionPrivateForumThread
		}
		return []notif.GrantRequirement{{Section: section, Item: consts.PermissionItemThread, ItemID: data.ThreadID, Action: consts.PermissionActionView}}, nil
	}
	return nil, nil
}

func (ReplyTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	return privateThreadSubscriberGrants(evt)
}

type forumReplyResult struct {
	CommentID       int64
	CommentText     string
	Appended        bool
	CanonicalTextOK bool
}

func performForumReply(
	cd *common.CoreData,
	userID int32,
	thread *db.GetThreadLastPosterAndPermsForUserRow,
	topic *db.GetForumTopicByIdForUserRow,
	languageID int32,
	text string,
) (forumReplyResult, error) {
	// 1. load current thread comments / candidate;
	commentsList, err := cd.ThreadComments(thread.Idforumthread)
	if err != nil {
		return forumReplyResult{}, err
	}

	if len(commentsList) > 0 {
		lastComment := commentsList[len(commentsList)-1]
		// 3. attempt append;
		res, err := cd.AttemptAppendForumComment(userID, thread.Idforumthread, topic.Idforumtopic, languageID, lastComment.Idcomments, text, topic.Handler == "private")

		// 6. append DB error -> error, no creation;
		if err != nil {
			return forumReplyResult{}, err
		}

		// 4. append succeeds -> return append result;
		// 7. post-mutation/reload problem must preserve the fact that append occurred.
		if res.Appended {
			return forumReplyResult{
				CommentID:       res.CommentID,
				CommentText:     res.CanonicalText,
				Appended:        true,
				CanonicalTextOK: res.TextAvailable,
			}, nil
		}
	}

	// 5. append returns zero rows cleanly -> create normal reply;
	var cid int64
	if topic.Handler == "private" {
		cid, err = cd.CreatePrivateForumCommentForCommenter(userID, thread.Idforumthread, topic.Idforumtopic, languageID, text)
	} else {
		cid, err = cd.CreateForumCommentForCommenter(userID, thread.Idforumthread, topic.Idforumtopic, languageID, text)
	}

	if err != nil {
		return forumReplyResult{}, err
	}

	return forumReplyResult{
		CommentID:       cid,
		CommentText:     text,
		Appended:        false,
		CanonicalTextOK: true,
	}, nil
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	log.Printf("START ACTION")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.SetCurrentThreadAndTopic(100, 10)
	session := cd.GetSession()
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Forum - Reply"
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		log.Printf("FAIL THREAD id=%d err=%v", cd.SelectedThreadID(), err)
		return fmt.Errorf("thread fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		log.Printf("FAIL TOPIC")
		return fmt.Errorf("topic fetch %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if session == nil { panic("session is nil!") }
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
	var isAppend bool

	// Check if this might be an append. We need the last comment ID.
	commentsList, err := cd.ThreadComments(threadRow.Idforumthread)
	if err != nil {
		log.Printf("Error fetching thread comments: %v", err)
		r.Form.Set("replytext", text)
		ThreadPageWithBasePath(w, r, base)
		return nil
	}
	if len(commentsList) > 0 {
		lastComment := commentsList[len(commentsList)-1]
		// Attempt append first
		var res common.AppendResult
		res, err = cd.AttemptAppendForumComment(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), lastComment.Idcomments, text, topicRow.Handler == "private")
		if err != nil {
			log.Printf("Append attempt error: %v", err)
		}
		if res.Appended {
			isAppend = true
			cid = res.CommentID
			if res.TextAvailable {
				text = res.CanonicalText
			} else {
				text = "" // don't index reconstructed string
			}
		}
	}

	if !isAppend && err == nil {
		if topicRow.Handler == "private" {
			cid, err = cd.CreatePrivateForumCommentForCommenter(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), text)
		} else {
			cid, err = cd.CreateForumCommentForCommenter(uid, threadRow.Idforumthread, topicRow.Idforumtopic, int32(languageId), text)
		}
	}
	log.Printf("DEBUG: cid=%d err=%v", cid, err)
	if err != nil || cid == 0 {
		if err == nil {
			err = handlers.ErrForbidden
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
	subjectPrefix := "Forum"
	if topicRow.Handler == "private" {
		subjectPrefix = "Private Forum"
	}
	data["SubjectPrefix"] = subjectPrefix
	log.Printf("DEBUG: HandleThreadUpdated")
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
		log.Printf("DEBUG: thread reply side effects: %v", err)
	}
	if evt := cd.Event(); evt != nil {
		evt.Data["URL"] = cd.AbsoluteURL(endUrl)
	}
	log.Printf("DEBUG: returning redirect to %s", endUrl)
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
