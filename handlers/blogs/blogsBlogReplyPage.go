package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
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

// ReplyBlogTask posts a comment reply on a blog.
type ReplyBlogTask struct{ tasks.TaskString }

var replyBlogTask = &ReplyBlogTask{TaskString: TaskReply}

// compile-time assertions that ReplyBlogTask sends notifications and
// auto-subscribes blog commenters.
var (
	_ tasks.Task                                    = (*ReplyBlogTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*ReplyBlogTask)(nil)
	_ notif.AutoSubscribeProvider                   = (*ReplyBlogTask)(nil)
	_ notif.GrantsRequiredProvider                  = (*ReplyBlogTask)(nil)
)

func (ReplyBlogTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyBlogTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

// GrantsRequired implements notif.GrantsRequiredProvider for blog replies.
func (ReplyBlogTask) GrantsRequired(evt eventbus.TaskEvent) ([]notif.GrantRequirement, error) {
	if t, ok := evt.Data["target"].(notif.Target); ok {
		return []notif.GrantRequirement{{Section: "blogs", Item: "entry", ItemID: t.ID, Action: "view"}}, nil
	}
	return nil, fmt.Errorf("target not provided")
}

// AutoSubscribePath records the reply so the commenter automatically watches
// for any further discussion.
// Automatically subscribe the commenter so they are notified about
// further discussion on the blog post they replied to.
// AutoSubscribePath allows the worker to add a subscription when new replies are
// posted so participants stay in the loop.
// AutoSubscribePath implements notif.AutoSubscribeProvider. It derives the
// subscription path from postcountworker event data when present.
func (ReplyBlogTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		return string(TaskReply), fmt.Sprintf("/forum/topic/%d/thread/%d", data.TopicID, data.ThreadID), nil
	}
	return string(TaskReply), evt.Path, nil
	//return TaskReply, evt.Path
}

func (ReplyBlogTask) IndexType() string { return searchworker.TypeComment }

func (ReplyBlogTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = ReplyBlogTask{}

func (ReplyBlogTask) Action(w http.ResponseWriter, r *http.Request) any {
	BlogReplyPostPage(w, r)
	return nil
}

func BlogReplyPostPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if err := handlers.ValidateForm(r, []string{"language", "replytext"}, []string{"language", "replytext"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	bid, err := strconv.Atoi(vars["blog"])

	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if bid == 0 {
		log.Printf("Error: no bid")
		http.Redirect(w, r, "?error="+"No bid", http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	blog, err := queries.GetBlogEntryForUserById(r.Context(), db.GetBlogEntryForUserByIdParams{
		ViewerIdusers: uid,
		ID:            int32(bid),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			_ = templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", cd)
			return
		default:
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	var pthid int32
	if blog.ForumthreadID.Valid {
		pthid = blog.ForumthreadID.Int32
	}
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: BloggerTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: BloggerTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: BloggerTopicDescription,
				Valid:  true,
			},
		})
		if err != nil {
			log.Printf("Error: createForumTopic: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		ptid = int32(ptidi)
	} else if err != nil {
		log.Printf("Error: findForumTopicByTitle: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := queries.MakeThread(r.Context(), ptid)
		if err != nil {
			log.Printf("Error: makeThread: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
		pthid = int32(pthidi)
		if err := queries.AssignThreadIdToBlogEntry(r.Context(), db.AssignThreadIdToBlogEntryParams{
			ForumthreadID: sql.NullInt32{Int32: pthid, Valid: true},
			Idblogs:       int32(bid),
		}); err != nil {
			log.Printf("Error: assignThreadIdToBlogEntry: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	endUrl := fmt.Sprintf("/blogs/blog/%d/comments", bid)

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
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: pthid, TopicID: ptid}
			evt.Data["CommentURL"] = cd.AbsoluteURL(endUrl)
			evt.Data["target"] = notif.Target{Type: "blog", ID: int32(bid)}
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
