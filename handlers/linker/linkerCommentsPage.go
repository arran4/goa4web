package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

func CommentsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Link           *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow
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
	cd.LoadSelectionsFromRequest(r)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	common.WithOffset(offset)(cd)
	data := Data{
		CanEdit:     false,
		IsReplyable: true,
	}
	vars := mux.Vars(r)
	linkId := 0
	if lid, err := strconv.Atoi(vars["link"]); err == nil {
		linkId = lid
	}
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)
	data.UserId = uid

	queries = r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		ID:           int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		default:
			log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	canReply := cd.HasGrant("linker", "link", "reply", link.ID)
	if !(cd.HasGrant("linker", "link", "view", link.ID) ||
		canReply ||
		cd.SelectedThreadCanReply()) {
		// TODO: Fix: Add enforced Access in router rather than task
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data.IsReplyable = canReply

	data.Link = link
	cd.PageTitle = fmt.Sprintf("Link %d Comments", link.ID)
	data.CanEdit = cd.IsAdmin() || cd.HasGrant("linker", "link", "edit-any", link.ID) || cd.HasGrant("linker", "link", "edit", link.ID)

	cd.SetCurrentThreadAndTopic(link.ThreadID, 0)
	commentRows, err := cd.SectionThreadComments("linker", "link", link.ThreadID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("thread comments: %s", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      link.ThreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: getThreadByIdForUserByIdWithLastPosterUserNameAndPermissions: %s", err)
			//http.Redirect(w, r, "?error="+err.Error(), http.StatusSeeOther)
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
	}

	replyType := r.URL.Query().Get("type")
	editCommentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	quoteId, _ := strconv.Atoi(r.URL.Query().Get("quote"))
	if text := r.URL.Query().Get("text"); text != "" {
		data.Text = text
	}
	data.Comments = commentRows
	data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return cmt.IsOwner
	}
	data.EditURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/linker/comments/%d?comment=%d#edit", link.ID, cmt.Idcomments)
	}
	data.EditSaveURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/linker/comments/%d/comment/%d", link.ID, cmt.Idcomments)
	}
	data.Editing = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return data.CanEditComment(cmt) && editCommentId != 0 && int32(editCommentId) == cmt.Idcomments
	}
	data.AdminURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if cd.IsAdmin() && cd.IsAdminMode() {
			return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
		}
		return ""
	}
	if editCommentId != 0 {
		data.IsReplyable = false
	}
	if quoteId != 0 {
		if c, err := cd.CommentByID(int32(quoteId)); err == nil && c != nil {
			switch replyType {
			case "full":
				data.Text = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithFullQuote())
			default:
				data.Text = a4code.QuoteText(c.Username.String, c.Text.String)
			}
		}
	}

	data.Thread = threadRow

	LinkerCommentsPageTmpl.Handle(w, r, data)
}

const LinkerCommentsPageTmpl tasks.Template = "linker/commentsPage.gohtml"

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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session := cd.GetSession()

	vars := mux.Vars(r)
	linkId, err := strconv.Atoi(vars["link"])

	if err != nil {
		return fmt.Errorf("link id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if linkId == 0 {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("no bid"))
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		ID:           int32(linkId),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := cd.ExecuteSiteTemplate(w, r, "admin/noAccessPage.gohtml", struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return nil
		default:
			log.Printf("getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending Error: %s", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return nil
		}
	}

	if !(cd.HasGrant("linker", "link", "view", link.ID) ||
		cd.HasGrant("linker", "link", "reply", link.ID)) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return nil
	}

	var pthid int32 = link.ThreadID
	pt, err := queries.SystemGetForumTopicByTitle(r.Context(), sql.NullString{
		String: LinkerTopicName,
		Valid:  true,
	})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := queries.CreateForumTopicForPoster(r.Context(), db.CreateForumTopicForPosterParams{
			ForumcategoryID: 0,
			ForumLang:       link.LanguageID,
			Title: sql.NullString{
				String: LinkerTopicName,
				Valid:  true,
			},
			Description: sql.NullString{
				String: LinkerTopicDescription,
				Valid:  true,
			},
			Handler:         "linker",
			Section:         "forum",
			GrantCategoryID: sql.NullInt32{},
			GranteeID:       sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			PosterID:        cd.UserID,
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
			ThreadID: pthid,
			ID:       int32(linkId),
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

	cid, err := cd.CreateLinkerCommentForCommenter(uid, pthid, int32(linkId), int32(languageId), text)
	if err != nil {
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		return fmt.Errorf("create comment fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             pthid,
		TopicID:              ptid,
		CommentID:            int32(cid),
		LabelItem:            "link",
		LabelItemID:          int32(linkId),
		CommentText:          text,
		CommentURL:           cd.AbsoluteURL(endUrl),
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        true,
	}); err != nil {
		log.Printf("linker comment side effects: %v", err)
	}

	return handlers.RefreshDirectHandler{TargetURL: endUrl}
}
