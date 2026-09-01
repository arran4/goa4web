package forum

import (
	"database/sql"
	"errors"
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
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse thread form: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	var fork *forkRequest
	if r.URL.Query().Get("quote_comment_id") != "" {
		topic, topicErr := cd.Queries().GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
			ViewerID:      uid,
			Idforumtopic:  int32(topicId),
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if topicErr != nil || topic == nil {
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("topic not found"))
			return nil
		}
		var status int
		fork, status, err = validateForkRequest(r, cd, topic, uid)
		if err != nil {
			w.WriteHeader(status)
			handlers.RenderErrorPage(w, r, err)
			return nil
		}
	}

	languageID, _ := strconv.Atoi(r.PostFormValue("language"))
	params := common.CreateForumThreadParams{
		ActorID:        uid,
		TopicID:        int32(topicId),
		LanguageID:     int32(languageID),
		Text:           r.PostFormValue("replytext"),
		Private:        base == "/private",
		EnforceHandler: true,
		BasePath:       base,
		PublicLabels:   r.PostForm["public"],
		PrivateLabels:  r.PostForm["private"],
	}
	if fork != nil {
		params.ReplyToCommentID = fork.commentID
		params.ReplyToThreadID = fork.threadID
	}
	result, err := cd.CreateForumThread(r.Context(), params)
	if err != nil {
		var notFound common.ForumResourceNotFoundError
		var mismatch common.ForumHandlerMismatchError
		var forbidden common.ForumOperationForbiddenError
		switch {
		case errors.As(err, &notFound):
			w.WriteHeader(http.StatusNotFound)
			handlers.RenderErrorPage(w, r, fmt.Errorf("topic not found"))
			return nil
		case errors.As(err, &mismatch):
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("forum handler does not match topic"))
			return nil
		case errors.As(err, &forbidden):
			w.WriteHeader(http.StatusForbidden)
			handlers.RenderErrorPage(w, r, fmt.Errorf("forbidden"))
			return nil
		default:
			return fmt.Errorf("create forum thread: %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RedirectHandler(result.URL)
}

type forkRequest struct {
	commentID int32
	threadID  int32
}

func validateForkRequest(r *http.Request, cd *common.CoreData, destinationTopic *db.GetForumTopicByIdForUserRow, uid int32) (*forkRequest, int, error) {
	rawCommentID := r.URL.Query().Get("quote_comment_id")
	if rawCommentID == "" {
		return nil, 0, nil
	}
	parsedCommentID, err := strconv.ParseInt(rawCommentID, 10, 32)
	if err != nil || parsedCommentID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid quote_comment_id")
	}
	commentID := int32(parsedCommentID)
	comment, err := cd.CommentByID(commentID)
	if err != nil || comment == nil {
		return nil, http.StatusForbidden, fmt.Errorf("forbidden: cannot access source comment")
	}
	sourceThread, err := cd.ForumThreadByID(comment.ForumthreadID)
	if err != nil || sourceThread == nil || sourceThread.Idforumthread != comment.ForumthreadID {
		return nil, http.StatusForbidden, fmt.Errorf("forbidden: cannot access source thread")
	}
	sourceTopic, err := cd.Queries().GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
		ViewerID:      uid,
		Idforumtopic:  sourceThread.ForumtopicIdforumtopic,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil || sourceTopic == nil {
		return nil, http.StatusForbidden, fmt.Errorf("forbidden: cannot access source topic")
	}
	if sourceThread.ForumtopicIdforumtopic != destinationTopic.Idforumtopic {
		return nil, http.StatusBadRequest, fmt.Errorf("fork source must belong to the destination topic")
	}
	if sourceTopic.Handler != destinationTopic.Handler {
		return nil, http.StatusBadRequest, fmt.Errorf("fork source and destination forum handlers do not match")
	}

	section := consts.PermissionSectionForum
	itemID := sourceTopic.Idforumtopic
	itemType := consts.PermissionItemTopic
	if sourceTopic.Handler == "private" {
		section = consts.PermissionSectionPrivateForumThread
		itemID = sourceThread.Idforumthread
		itemType = consts.PermissionItemThread
	}
	replyable, err := userCanReplyToThread(r.Context(), cd.Queries(), section, itemType, itemID, sourceThread.Idforumthread, uid)
	if err != nil || !replyable {
		return nil, http.StatusForbidden, fmt.Errorf("forbidden: cannot fork source thread")
	}
	return &forkRequest{commentID: commentID, threadID: sourceThread.Idforumthread}, 0, nil
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
