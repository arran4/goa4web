package forum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/gorilla/mux"
)

// CreateThreadTask handles creating a new forum thread.
type CreateThreadTask struct{ tasks.TaskString }

var (
	createThreadTask = &CreateThreadTask{TaskString: TaskCreateThread}

	// CreateThreadTaskHandler handles creating threads and is exported for reuse.
	CreateThreadTaskHandler = createThreadTask

	// Interface checks ensure the new thread hooks into notifications so
	// authors follow replies, administrators are alerted and subscribers see
	// new discussions.
	_ tasks.Task                                    = (*CreateThreadTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*CreateThreadTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*CreateThreadTask)(nil)
	_ notif.AutoSubscribeProvider                   = (*CreateThreadTask)(nil)
	_ notif.GrantsRequiredProvider                  = (*CreateThreadTask)(nil)
	_ tasks.EmailTemplatesRequired                  = (*CreateThreadTask)(nil)
	_ searchworker.IndexedTask                      = CreateThreadTask{}
)

func (CreateThreadTask) IndexType() string { return searchworker.TypeComment }

func (CreateThreadTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

func (CreateThreadTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateForumThreadCreate.EmailTemplates(), true
}

func (CreateThreadTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateForumThread.NotificationTemplate()
	return &s
}

func (CreateThreadTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationForumThreadCreate.EmailTemplates(), true
}

func (CreateThreadTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationForumThreadCreate.NotificationTemplate()
	return &v
}

func (CreateThreadTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{tasks.Template(ForumThreadNewPageTmpl)},
		append(EmailTemplateForumThreadCreate.RequiredTemplates(), EmailTemplateAdminNotificationForumThreadCreate.RequiredTemplates()...)...)
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
		base := "/forum"
		if idx := strings.Index(evt.Path, "/topic/"); idx > 0 {
			base = evt.Path[:idx]
		}
		return string(TaskReply), fmt.Sprintf("%s/topic/%d/thread/%d", base, data.TopicID, data.ThreadID), nil
	}
	return string(TaskCreateThread), evt.Path, nil
}

func (CreateThreadTask) AutoSubscribeGrants(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
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

func (CreateThreadTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	return privateThreadSubscriberGrants(evt)
}

func (CreateThreadTask) Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Languages          []*db.Language
		SelectedLanguageId int
		BasePath           string
		Topic              *db.GetForumTopicByIdForUserRow
		QuoteText          string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - New Thread"

	vars := mux.Vars(r)
	topicId, err := strconv.Atoi(vars["topic"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid topic id: %w", err))
		return
	}

	uid := cd.UserID
	queries := cd.Queries()
	topic, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
		ViewerID:      uid,
		Idforumtopic:  int32(topicId),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("topic not found: %w", err))
		return
	}

	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}

	data := Data{
		SelectedLanguageId: int(cd.PreferredLanguageID(cd.Config.DefaultLanguage)),
		BasePath:           base,
		Topic:              topic,
	}

	// Handle quoting if query parameters are present.
	// This logic mirrors the QuoteApi functionality but runs server-side to
	// pre-populate the thread creation form.
	quoteCommentId := r.URL.Query().Get("quote_comment_id")
	if quoteCommentId != "" {
		if cId, err := strconv.Atoi(quoteCommentId); err == nil {
			// Retrieve the comment ensuring the user has permission to view it.
			if c, err := cd.CommentByID(int32(cId)); err == nil && c != nil {
				quoteType := r.URL.Query().Get("quote_type")
				var text string
				switch quoteType {
				case "paragraph":
					text = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithParagraphQuote()) + "\n\n"
				case "full":
					text = a4code.QuoteText(c.Username.String, c.Text.String) + "\n\n"
				case "selected":
					start, _ := strconv.Atoi(r.URL.Query().Get("quote_start"))
					end, _ := strconv.Atoi(r.URL.Query().Get("quote_end"))
					sub, err := a4code.Substring(c.Text.String, start, end)
					if err != nil {
						log.Printf("Substring error: %v", err)
					}
					text = a4code.QuoteText(c.Username.String, sub) + "\n\n"
				default:
					text = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithParagraphQuote()) + "\n\n"
				}

				data.QuoteText = text
			}
		}
	}

	languageRows, err := cd.Languages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	data.Languages = languageRows

	_ = ForumThreadNewPageTmpl.Handle(w, r, data)
}

const ForumThreadNewPageTmpl tasks.Template = "domains/forum/threadNewPage.gohtml"

func (CreateThreadTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	vars := mux.Vars(r)
	topicId, err := strconv.Atoi(vars["topic"])
	if err != nil {
		return fmt.Errorf("topic id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)

	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	section := consts.PermissionSectionForum
	if base == "/private" {
		section = consts.PermissionSectionPrivateForum
	}
	allowed, err := UserCanCreateThread(r.Context(), queries, section, int32(topicId), uid)
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

	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm error: %v", err)
	}
	if err := cd.SetThreadPublicLabels(int32(threadId), r.PostForm["public"]); err != nil {
		log.Printf("set public labels: %v", err)
	}
	if err := cd.SetThreadPrivateLabels(int32(threadId), r.PostForm["private"]); err != nil {
		log.Printf("set private labels: %v", err)
	}

	var topicTitle, author string
	var topic *db.GetForumTopicByIdForUserRow
	if trow, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{ViewerID: uid, Idforumtopic: int32(topicId), ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0}}); err == nil {
		topicTitle = trow.Title.String
		topic = trow
	}
	if u := cd.UserByID(uid); u != nil {
		author = u.Username.String
	}
	quoteCommentId := r.URL.Query().Get("quote_comment_id")
	var replyToCommentId, replyToThreadId sql.NullInt32
	if quoteCommentId != "" {
		if cId, err := strconv.Atoi(quoteCommentId); err == nil {
			c, err := cd.CommentByID(int32(cId))
			if err != nil || c == nil {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden: cannot access source comment"))
				return nil
			}

			th, err := cd.ForumThreadByID(c.ForumthreadID)
			if err != nil || th == nil {
				w.WriteHeader(http.StatusForbidden)
				handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden: cannot access source thread"))
				return nil
			}

			if th.ForumtopicIdforumtopic != int32(topicId) {
				w.WriteHeader(http.StatusBadRequest)
				handlers.RenderErrorPage(w, r, fmt.Errorf("bad request: fork must belong to the same topic"))
				return nil
			}

			replyToCommentId = sql.NullInt32{Int32: int32(cId), Valid: true}
			replyToThreadId = sql.NullInt32{Int32: c.ForumthreadID, Valid: true}
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("%s/topic/%d/thread/%d", base, topicId, threadId)

	var cid int64
	if topic.Handler == "private" {
		if replyToThreadId.Valid {
			if err := cd.CopyPrivateThreadGrantsToThread(replyToThreadId.Int32, int32(threadId)); err != nil {
				return fmt.Errorf("copying private thread grants to thread: %w", err)
			}
		} else {
			if err := cd.CopyPrivateTopicGrantsToThread(int32(topicId), int32(threadId)); err != nil {
				return fmt.Errorf("copying private topic grants to thread: %w", err)
			}
		}
		cid, err = cd.CreatePrivateForumOpeningCommentForPoster(uid, int32(threadId), int32(topicId), int32(languageId), text)
		if err != nil {
			log.Printf("Error: create forum comment: %s", err)
			return fmt.Errorf("creating private topic comment: %w", err)
		}
	} else {
		cid, err = cd.CreateForumOpeningCommentForPoster(uid, int32(threadId), int32(topicId), int32(languageId), text)
		if err != nil {
			log.Printf("Error: create forum comment: %s", err)
			return fmt.Errorf("create forum comment %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	if cid == 0 {
		log.Printf("Error: cid == 0 on comment create - no error")
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(handlers.ErrForbidden))
	}

	if replyToThreadId.Valid {
		if err := cd.Queries().SetThreadReplyTo(r.Context(), db.SetThreadReplyToParams{
			ReplyToCommentID: replyToCommentId,
			ReplyToThreadID:  replyToThreadId,
			Idforumthread:    int32(threadId),
		}); err != nil {
			log.Printf("Error: setting thread reply to: %s", err)
			return fmt.Errorf("setting thread reply to %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	if evt := cd.Event(); evt != nil {
		evt.Path = endUrl
	}

	subjectPrefix := "Forum"
	if topic.Handler == "private" {
		subjectPrefix = "Private Forum"
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:         int32(threadId),
		TopicID:          int32(topicId),
		CommentID:        int32(cid),
		TopicTitle:       topicTitle,
		Author:           author,
		Username:         author,
		CommentText:      text,
		PostURL:          cd.AbsoluteURL(endUrl),
		ThreadURL:        cd.AbsoluteURL(endUrl),
		IncludePostCount: true,
		IncludeSearch:    true,
		MarkThreadRead:   true,
		AdditionalData: map[string]any{
			"ThreadOpenerPreview": a4code.SnipTextWords(text, 10),
			"SubjectPrefix":       subjectPrefix,
		},
	}); err != nil {
		log.Printf("thread create side effects: %v", err)
	}

	return handlers.RedirectHandler(endUrl)
}

func ThreadNewCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicId, _ := strconv.Atoi(vars["topic"])
	base := "/forum"
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if cd.ForumBasePath != "" {
			base = cd.ForumBasePath
		}
	}
	endUrl := fmt.Sprintf("%s/topic/%d", base, topicId)
	http.Redirect(w, r, endUrl, http.StatusSeeOther)
}
