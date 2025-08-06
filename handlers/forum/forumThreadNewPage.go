package forum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/blogs"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

// CreateThreadTask handles creating a new forum thread.
type CreateThreadTask struct{ tasks.TaskString }

var (
	createThreadTask = &CreateThreadTask{TaskString: TaskCreateThread}

	// Interface checks ensure the new thread hooks into notifications so
	// authors follow replies, administrators are alerted and subscribers see
	// new discussions.
	_ tasks.Task                                    = (*CreateThreadTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*CreateThreadTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*CreateThreadTask)(nil)
	_ notif.AutoSubscribeProvider                   = (*CreateThreadTask)(nil)
	_ searchworker.IndexedTask                      = CreateThreadTask{}
)

func (CreateThreadTask) IndexType() string { return searchworker.TypeComment }

func (CreateThreadTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

func (CreateThreadTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("threadEmail")
}

func (CreateThreadTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("thread")
	return &s
}

func (CreateThreadTask) AdminEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumThreadCreateEmail")
}

func (CreateThreadTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumThreadCreateEmail")
	return &v
}

// AutoSubscribePath records the created thread so the author and topic
// followers automatically receive updates when others reply.
// When a user creates a thread they expect to follow any replies.
// AutoSubscribePath allows new thread creators to automatically watch for replies.

// AutoSubscribePath implements notif.AutoSubscribeProvider. When the
// postcountworker provides context, a subscription to the created thread is
// generated.
func (CreateThreadTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskCreateThread), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return string(TaskCreateThread), evt.Path, nil
}

func (CreateThreadTask) Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Languages          []*db.Language
		SelectedLanguageId int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - New Thread"
	data := Data{
		CoreData:           cd,
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
	}

	languageRows, err := cd.Languages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Languages = languageRows

	blogs.CustomBlogIndex(data.CoreData, r)

	handlers.TemplateHandler(w, r, "threadNewPage.gohtml", data)
}

func (CreateThreadTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	vars := mux.Vars(r)
	topicId, err := strconv.Atoi(vars["topic"])
	if err != nil {
		return fmt.Errorf("topic id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	allowed, err := UserCanCreateThread(r.Context(), queries, int32(topicId), uid)
	if err != nil {
		log.Printf("UserCanCreateThread error: %v", err)
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
		return nil
	}
	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
		return nil
	}

	threadId, err := queries.SystemCreateThread(r.Context(), int32(topicId))
	if err != nil {
		log.Printf("Error: makeThread: %s", err)
		return fmt.Errorf("make thread %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	var topicTitle, author string
	if trow, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{ViewerID: uid, Idforumtopic: int32(topicId), ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0}}); err == nil {
		topicTitle = trow.Title.String
	}
	if u, err := queries.SystemGetUserByID(r.Context(), uid); err == nil {
		author = u.Username.String
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["TopicTitle"] = topicTitle
			evt.Data["Author"] = author
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d", topicId, threadId)

	cid, err := queries.CreateCommentForCommenter(r.Context(), db.CreateCommentForCommenterParams{
		LanguageID:         int32(languageId),
		CommenterID:        uid,
		ForumthreadID:      int32(threadId),
		Text:               sql.NullString{String: text, Valid: true},
		GrantForumthreadID: sql.NullInt32{Int32: int32(threadId), Valid: true},
		GranteeID:          sql.NullInt32{Int32: uid, Valid: true},
	})
	if err != nil {
		log.Printf("Error: makeThread: %s", err)
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: int32(threadId), TopicID: int32(topicId)}
			evt.Data["PostURL"] = cd.AbsoluteURL(endUrl)
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

func ThreadNewCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])

	endUrl := fmt.Sprintf("/forum/topic/%d", topicId)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
