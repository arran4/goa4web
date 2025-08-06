package linker

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
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

func CommentsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Link           *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
		CanReply       bool
		Comments       []*db.GetCommentsByThreadIdForUserRow
		IsReplyable    bool
		Text           string
		CanEdit        bool
		UserId         int32
		Thread         *db.GetThreadLastPosterAndPermsRow
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
	}

	offset := 0
	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = off
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	common.WithOffset(offset)(cd)
	data := Data{
		CoreData:    cd,
		CanReply:    cd.UserID != 0,
		CanEdit:     false,
		IsReplyable: true,
	}
	vars := mux.Vars(r)
	linkId := 0
	if lid, err := strconv.Atoi(vars["link"]); err == nil {
		linkId = lid
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	queries = r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		Idlinker:     int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	if !(cd.HasGrant("linker", "link", "view", link.Idlinker) ||
		cd.HasGrant("linker", "link", "comment", link.Idlinker) ||
		cd.HasGrant("linker", "link", "reply", link.Idlinker)) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data.Link = link
	data.CoreData.PageTitle = fmt.Sprintf("Link %d Comments", link.Idlinker)
	data.CanEdit = cd.HasRole("administrator") || uid == link.UsersIdusers

	cd.SetCurrentThreadAndTopic(link.ForumthreadID, 0)
	commentRows, err := cd.SectionThreadComments("linker", "link", link.ForumthreadID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("thread comments: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      link.ForumthreadID,
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

	commentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	data.Comments = commentRows
	data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return data.CoreData.CanEditAny() || cmt.IsOwner
	}
	data.EditURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/linker/comments/%d?comment=%d#edit", link.Idlinker, cmt.Idcomments)
	}
	data.EditSaveURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/linker/comments/%d/comment/%d", link.Idlinker, cmt.Idcomments)
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

	handlers.TemplateHandler(w, r, "commentsPage.gohtml", data)
}

type replyTask struct{ tasks.TaskString }

var replyTaskEvent = &replyTask{TaskString: TaskReply}
var _ tasks.Task = (*replyTask)(nil)

func (replyTask) IndexType() string { return searchworker.TypeComment }

func (replyTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = replyTask{}

func (replyTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		return fmt.Errorf("link id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if linkId == 0 {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("no bid"))
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		Idlinker:     int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return nil
		default:
			log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return nil
		}
	}

	if !(cd.HasGrant("linker", "link", "view", link.Idlinker) ||
		cd.HasGrant("linker", "link", "comment", link.Idlinker) ||
		cd.HasGrant("linker", "link", "reply", link.Idlinker)) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return nil
	}

	var pthid int32 = link.ForumthreadID
	pt, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{
		String: LinkerTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.SystemCreateForumTopic(r.Context(), db.SystemCreateForumTopicParams{
			ForumcategoryIdforumcategory: 0,
			Title: sql.NullString{
				String: LinkerTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: LinkerTopicDescription,
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
		if err := queries.SystemAssignLinkerThreadID(r.Context(), db.SystemAssignLinkerThreadIDParams{
			ForumthreadID: pthid,
			Idlinker:      int32(linkId),
		}); err != nil {
			return fmt.Errorf("assign linker thread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	text := r.PostFormValue("replytext")
	languageId := 0
	if lid, err := strconv.Atoi(r.PostFormValue("language")); err == nil {
		languageId = lid
	}
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/linker/comments/%d", linkId)

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

	return handlers.RefreshDirectHandler{TargetURL: endUrl}
}
