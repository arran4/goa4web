package imagebbs

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

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
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

func (ReplyTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) *notif.EmailTemplates {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	return notif.NewEmailTemplates("replyEmail")
}

func (ReplyTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}
	s := notif.NotificationTemplateFilenameGenerator("reply")
	return &s
}

func (ReplyTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	return string(TaskReply), evt.Path, nil
}

func BoardThreadPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Replyable      bool
		ForumThreadId  int
		Comments       []*db.GetCommentsByThreadIdForUserRow
		BoardId        int
		ImagePost      *db.GetImagePostByIDForListerRow
		Thread         *db.GetThreadLastPosterAndPermsRow
		IsReplyable    bool
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)
	thidStr := vars["thread"]
	if thidStr == "" {
		thidStr = vars["replyTo"]
	}
	thid, _ := strconv.Atoi(thidStr)
	cd.PageTitle = fmt.Sprintf("Thread %d/%d", bid, thid)

	data := Data{Replyable: true, BoardId: bid, ForumThreadId: thid, IsReplyable: true}

	if !cd.HasGrant("imagebbs", "board", "view", int32(bid)) {
		_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{})
		return
	}

	cd.SetCurrentThreadAndTopic(int32(thid), 0)
	commentRows, err := cd.SelectedSectionThreadComments()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("thread comments: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	threadRow, err := cd.SelectedThread()
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
	common.WithOffset(offset)(cd)

	commentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	data.Comments = commentRows
	data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return cd.CanEditAny() || cmt.IsOwner
	}
	data.EditURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/forum/topic/%d/thread/%d?comment=%d#edit", threadRow.ForumtopicIdforumtopic, threadRow.Idforumthread, cmt.Idcomments)
	}
	data.EditSaveURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/forum/topic/%d/thread/%d/comment/%d", threadRow.ForumtopicIdforumtopic, threadRow.Idforumthread, cmt.Idcomments)
	}
	data.Editing = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return data.CanEditComment(cmt) && commentId != 0 && int32(commentId) == cmt.Idcomments
	}
	data.AdminURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if cd.HasRole("administrator") {
			return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
		}
		return ""
	}
	if commentId != 0 {
		data.IsReplyable = false
	}

	data.Thread = threadRow
	post, err := cd.ImagePostByID(int32(bid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{})
			return
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	data.ImagePost = post

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
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)

	uid, _ = session.Values["UID"].(int32)

	if bid == 0 {
		return fmt.Errorf("no bid %w", handlers.ErrRedirectOnSamePageHandler(errors.New("no bid")))
	}

	queries := cd.Queries()

	post, err := queries.GetImagePostByIDForLister(r.Context(), db.GetImagePostByIDForListerParams{
		ListerID:     uid,
		ID:           int32(bid),
		ListerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{})
			return nil
		default:
			return fmt.Errorf("get image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	var pthid int32 = post.ForumthreadID
	pt, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{
		String: ImageBBSTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.SystemCreateForumTopic(r.Context(), db.SystemCreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			LanguageIdlanguage:           0,
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
		pthidi, err := queries.SystemCreateThread(r.Context(), ptid)
		if err != nil {
			return fmt.Errorf("make thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		pthid = int32(pthidi)
		if err := queries.SystemAssignImagePostThreadID(r.Context(), db.SystemAssignImagePostThreadIDParams{
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

	cid, err := queries.CreateCommentForCommenter(r.Context(), db.CreateCommentForCommenterParams{
		LanguageID:         int32(languageId),
		CommenterID:        uid,
		ForumthreadID:      pthid,
		Text:               sql.NullString{String: text, Valid: true},
		GrantForumthreadID: sql.NullInt32{Int32: pthid, Valid: true},
		GranteeID:          sql.NullInt32{Int32: uid, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(cid), ThreadID: pthid, TopicID: ptid}
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
