package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
)

// ReplyTask posts a reply within a thread.
type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

var _ tasks.Task = (*ReplyTask)(nil)

// ReplyTask alerts watchers of new posts and auto-subscribes the replier so
// they see further responses.
var _ notif.SubscribersNotificationTemplateProvider = (*ReplyTask)(nil)
var _ notif.AutoSubscribeProvider = (*ReplyTask)(nil)

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

func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	return string(TaskReply), evt.Path, nil
}

func BoardThreadPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*db.GetCommentsByThreadIdForUserRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*db.Language
		SelectedLanguageId int32
		EditSaveUrl        string
		AdminUrl           string
	}
	type Data struct {
		*common.CoreData
		Replyable          bool
		Languages          []*db.Language
		SelectedLanguageId int
		ForumThreadId      int
		Comments           []*CommentPlus
		BoardId            int
		ImagePost          *db.GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCountForUserRow
		Thread             *db.GetThreadLastPosterAndPermsRow
		Offset             int
		IsReplyable        bool
	}

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])
	thid, _ := strconv.Atoi(vars["thread"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = fmt.Sprintf("Thread %d/%d", bid, thid)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	var uid int32
	uid, _ = session.Values["UID"].(int32)

	queries := cd.Queries()
	data := Data{
		CoreData:      cd,
		Replyable:     true,
		BoardId:       bid,
		ForumThreadId: thid,
	}

	if !data.CoreData.HasGrant("imagebbs", "board", "view", int32(bid)) {
		_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd)
		return
	}

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: int32(thid),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      int32(thid),
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: getThreadByIdForUserByIdWithLastPosterUserNameAndPermissions: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if data.CoreData.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", threadRow.ForumtopicIdforumtopic, threadRow.Idforumthread, row.Idcomments)
			editSaveUrl = fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", threadRow.ForumtopicIdforumtopic, threadRow.Idforumthread, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       true,
			EditUrl:                         editUrl,
			EditSaveUrl:                     editSaveUrl,
			Editing:                         commentId != 0 && int32(commentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       languageRows,
			SelectedLanguageId:              row.LanguageIdlanguage,
			AdminUrl: func() string {
				if data.CoreData.HasRole("administrator") {
					return fmt.Sprintf("/admin/comment/%d", row.Idcomments)
				} else {
					return ""
				}
			}(),
		})
	}

	data.Thread = threadRow
	post, err := queries.GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCountForUser(r.Context(), db.GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCountForUserParams{
		ViewerID:     uid,
		ID:           int32(bid),
		ViewerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd)
			return
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.ImagePost = post

	data.Languages = languageRows

	handlers.TemplateHandler(w, r, "boardThreadPage.gohtml", data)
}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}

	var uid int32

	vars := mux.Vars(r)
	bid, err := strconv.Atoi(vars["boardno"])

	uid, _ = session.Values["UID"].(int32)

	if err != nil {
		return fmt.Errorf("board id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if bid == 0 {
		return fmt.Errorf("no bid %w", handlers.ErrRedirectOnSamePageHandler(errors.New("no bid")))
	}

	queries := cd.Queries()

	post, err := queries.GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCountForUser(r.Context(), db.GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCountForUserParams{
		ViewerID:     uid,
		ID:           int32(bid),
		ViewerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd)
			return nil
		default:
			return fmt.Errorf("get image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	var pthid int32 = post.ForumthreadID
	pt, err := queries.FindForumTopicByTitle(r.Context(), sql.NullString{
		String: ImageBBSTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopic(r.Context(), db.CreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: ImageBBSTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: ImageBBSTopicDescription,
				Valid:  true,
			},
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
		if err := queries.UpdateImagePostByIdForumThreadId(r.Context(), db.UpdateImagePostByIdForumThreadIdParams{
			ForumthreadID: pthid,
			Idimagepost:   int32(bid),
		}); err != nil {
			return fmt.Errorf("assign imagebbs to thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ = session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/imagebbss/imagebbs/%d/comments", bid)

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
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: pthid, TopicID: ptid}
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
