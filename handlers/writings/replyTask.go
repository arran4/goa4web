package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
	"github.com/gorilla/mux"
)

// ReplyTask posts a comment reply.
type ReplyTask struct{ tasks.TaskString }

// ReplyTask implements these interfaces so that when a user replies to a
// writing everyone following the discussion is automatically subscribed and
// receives a notification using the shared reply templates. This keeps readers
// informed when conversations continue.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

var replyTask = &ReplyTask{TaskString: TaskReply}

var _ tasks.Task = (*ReplyTask)(nil)
var _ notif.GrantsRequiredProvider = (*ReplyTask)(nil)

// replying should notify anyone following the discussion
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)

// repliers expect to automatically follow further conversation
// ReplyTask notifies followers and auto-subscribes the author so replies aren't missed.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

func (ReplyTask) IndexType() string { return searchworker.TypeComment }

func (ReplyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

func (ReplyTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

// GrantsRequired implements notif.GrantsRequiredProvider for replies.
func (ReplyTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

// AutoSubscribePath implements notif.AutoSubscribeProvider. It builds the
// subscription path for the writing's forum thread when that data is provided
// by the event.
func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskReply), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return string(TaskReply), evt.Path, nil
}

var _ searchworker.IndexedTask = ReplyTask{}
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}

	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["article"])
	if err != nil {
		return fmt.Errorf("article id parse %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if aid == 0 {
		log.Printf("no article id")
		return fmt.Errorf("no article %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("no article")))
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), db.GetWritingByIdForUserDescendingByPublishedDateParams{
		ViewerID:      uid,
		Idwriting:     int32(aid),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			return templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", cd)
		}
		log.Printf("getArticlePost: %v", err)
		return err
	}

	pthid := post.ForumthreadID
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{String: WritingTopicName, Valid: true})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title:                        sql.NullString{String: WritingTopicName, Valid: true},
			Description:                  sql.NullString{String: WritingTopicDescription, Valid: true},
		})
		if err != nil {
			log.Printf("createForumTopic: %v", err)
			return fmt.Errorf("create forum topic %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("findForumTopicByTitle: %v", err)
		return fmt.Errorf("find forum topic %w", handlers.ErrRedirectOnSamePageHandler(err))
	} else {
		ptid = pt.Idforumtopic
	}

	if pthid == 0 {
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("makeThread: %v", err)
			return fmt.Errorf("make thread %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		pthid = int32(pthidi)
		if err := queries.AssignWritingThisThreadId(r.Context(), db.AssignWritingThisThreadIdParams{ForumthreadID: pthid, Idwriting: int32(aid)}); err != nil {
			log.Printf("assign article thread: %v", err)
			return fmt.Errorf("assign article thread %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["target"] = notif.Target{Type: "writing", ID: int32(aid)}
		}
	}

	text := r.PostFormValue("replytext")
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if _, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      pthid,
		Text:               sql.NullString{String: text, Valid: true},
	}); err != nil {
		log.Printf("createComment: %v", err)
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: pthid, TopicID: ptid}
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: 0, Text: text}
		}
	}

	return nil
}
