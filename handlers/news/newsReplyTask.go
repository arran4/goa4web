package news

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
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
)

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
	return notif.NewEmailTemplates("adminNotificationNewsReplyEmail")
}

func (ReplyTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsReplyEmail")
	return &v
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
	pid, err := strconv.Atoi(vars["post"])

	if err != nil {
		return fmt.Errorf("post id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if pid == 0 {
		log.Printf("Error: no bid")
		return fmt.Errorf("no bid %w", handlers.ErrRedirectOnSamePageHandler(errors.New("No bid")))
	}

	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("news", "post", "reply", int32(pid)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}

	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		log.Printf("GetNewsPostByIdWithWriterIdAndThreadCommentCountForUser Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil
	}

	var pthid = post.ForumthreadID
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: NewsTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: NewsTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: NewsTopicDescription,
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("Error: createForumTopic: %s", err)
			return fmt.Errorf("create forum topic fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByTitle: %s", err)
		return fmt.Errorf("find forum topic fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			return fmt.Errorf("make thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		pthid = int32(pthidi)
		if err := queries.SystemAssignNewsThreadID(r.Context(), db.SystemAssignNewsThreadIDParams{
			ForumthreadID: pthid,
			Idsitenews:    int32(pid),
		}); err != nil {
			log.Printf("Error: assign_news_to_thread: %s", err)
			return fmt.Errorf("assign news thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := cd.AbsoluteURL(fmt.Sprintf("/news/news/%d", pid))

	evt := cd.Event()
	if evt.Data == nil {
		evt.Data = map[string]any{}
	}
	evt.Data["CommentURL"] = endUrl

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      pthid,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: createComment: %s", err)
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: text}
		}
	}

	return nil
}
