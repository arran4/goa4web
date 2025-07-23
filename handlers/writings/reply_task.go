package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

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
)

// ReplyTask posts a comment reply.
type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

var _ tasks.Task = (*ReplyTask)(nil)
var _ notif.GrantsRequiredProvider = (*ReplyTask)(nil)
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)
var _ searchworker.IndexedTask = ReplyTask{}

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

func (ReplyTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "writing", Item: "article", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

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
	aid, err := strconv.Atoi(vars["article"])
	if err != nil || aid == 0 {
		return fmt.Errorf("article id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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
			if err := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", cd); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return nil
		}
		return fmt.Errorf("get writing fail %w", err)
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
			return fmt.Errorf("create forum topic fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		ptid = int32(ptidi)
	} else if err != nil {
		return fmt.Errorf("find forum topic fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	} else {
		ptid = pt.Idforumtopic
	}

	if pthid == 0 {
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			return fmt.Errorf("make thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		pthid = int32(pthidi)
		if err := queries.AssignWritingThisThreadId(r.Context(), db.AssignWritingThisThreadIdParams{ForumthreadID: pthid, Idwriting: int32(aid)}); err != nil {
			return fmt.Errorf("assign article thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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
	languageID, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		return fmt.Errorf("language parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if _, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageID),
		UsersIdusers:       uid,
		ForumthreadID:      pthid,
		Text:               sql.NullString{String: text, Valid: true},
	}); err != nil {
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: pthid, TopicID: ptid}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: 0, Text: text}
		}
	}

	return nil
}
