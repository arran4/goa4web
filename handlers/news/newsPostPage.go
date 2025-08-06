package news

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func NewsPostPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Post           *common.NewsPost
		Thread         *db.GetThreadLastPosterAndPermsRow
		Comments       []*db.GetCommentsByThreadIdForUserRow
		ReplyText      string
		IsReplyable    bool
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "News"
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	data := Data{
		IsReplyable: true,
	}
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["news"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	posts, err := cd.LatestNewsList(0, 50)
	if err != nil {
		log.Printf("LatestNewsList: %v", err)
		handlers.RenderErrorPage(w, r, err)
		return
	}
	var post *common.NewsPost
	for _, p := range posts {
		if p.Idsitenews == int32(pid) {
			post = p
			break
		}
	}
	if post == nil {
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}
	if !cd.HasGrant("news", "post", "view", post.Idsitenews) {
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}

	editingId, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	replyType := r.URL.Query().Get("type")

	cd.SetCurrentThreadAndTopic(post.ForumthreadID, 0)
	commentRows, err := cd.SectionThreadComments("news", "post", post.ForumthreadID)
	if err != nil {
		log.Printf("thread comments: %v", err)
	}

	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      int32(post.ForumthreadID),
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
	common.WithOffset(offset)(cd)
	editCommentId, _ := strconv.Atoi(r.URL.Query().Get("editComment"))

	data.Comments = commentRows
	data.Thread = threadRow
	post.Editing = editingId == int(post.Idsitenews)
	data.Post = post

	data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return cd.CanEditAny() || cmt.IsOwner
	}
	data.EditURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("?editComment=%d#edit", cmt.Idcomments)
	}
	data.EditSaveURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if !data.CanEditComment(cmt) {
			return ""
		}
		return fmt.Sprintf("/news/news/%d/comment/%d", pid, cmt.Idcomments)
	}
	data.Editing = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return data.CanEditComment(cmt) && editCommentId != 0 && int32(editCommentId) == cmt.Idcomments
	}
	data.AdminURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if cd.HasRole("administrator") {
			return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
		}
		return ""
	}

	if c, err := cd.CurrentComment(r); err == nil && c != nil {
		data.IsReplyable = false
		switch replyType {
		case "full":
			data.ReplyText = a4code.FullQuoteOf(c.Username.String, c.Text.String)
		default:
			data.ReplyText = a4code.QuoteOfText(c.Username.String, c.Text.String)
		}
	} else if r.URL.Query().Has("comment") {
		data.IsReplyable = false
	}

	handlers.TemplateHandler(w, r, "postPage.gohtml", data)
}
