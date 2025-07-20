package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	postcountworker "github.com/arran4/goa4web/workers/postcountworker"
	searchworker "github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// ReplyTask handles replying to an existing thread.
type ReplyTask struct{ tasks.TaskString }

// ReplyTask asserts these interfaces so that anyone responding to a thread
// automatically follows subsequent replies and administrators receive an email
// summary. This keeps discussions active and lets mods monitor forum activity.
var _ tasks.Task = (*ReplyTask)(nil)

// send notifications to thread subscribers when someone replies
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)

// replies should automatically watch the thread so users see future updates
var _ notif.AdminEmailTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)


// Build time checks so replying to a thread always triggers subscription and
// notification delivery using the standard templates, keeping readers in the
// conversation.
var replyTask = &ReplyTask{TaskString: TaskReply}

// Compile-time interface checks with user focused reasoning. Subscribing allows
// thread followers to hear about replies while administrators are alerted to new
// content. AutoSubscribeProvider ensures the author is kept in the loop.
var _ tasks.Task = (*ReplyTask)(nil)

// ReplyTask notifies thread subscribers and automatically subscribes the author
// to keep them in the conversation.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)

// admins track replies across the forum
var _ notif.AdminEmailTemplateProvider = (*ReplyTask)(nil)

// participants expect to automatically follow discussions they reply to
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

var replyTask = &ReplyTask{TaskString: TaskReply}

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

func (ReplyTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumReplyEmail")
}

func (ReplyTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumReplyEmail")
	return &v
}

var _ searchworker.IndexedTask = ReplyTask{}

// AutoSubscribePath ensures authors automatically receive updates on replies.
var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("forumReplyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("forum_reply")
	return &s
}

// AutoSubscribePath implements notif.AutoSubscribeProvider. The subscription is
// created for the originating forum thread when that information is available.
func (ReplyTask) AutoSubscribePath(evt eventbus.Event) (string, string) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskReply), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID)
	}
	return string(TaskReply), evt.Path
}

var _ searchworker.IndexedTask = ReplyTask{}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	threadRow := r.Context().Value(consts.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(consts.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["reply"] = notifications.ForumReplyInfo{TopicTitle: topicRow.Title.String, ThreadID: threadRow.Idforumthread, Thread: threadRow}
		}
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      threadRow.Idforumthread,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: CreateComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: threadRow.Idforumthread, TopicID: topicRow.Idforumtopic}
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

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func TopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	threadRow := r.Context().Value(consts.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(consts.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
