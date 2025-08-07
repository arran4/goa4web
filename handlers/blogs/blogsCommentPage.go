package blogs

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
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func CommentPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*db.GetBlogEntryForListerByIDRow
		EditUrl string
	}
	type Data struct {
		Blog           *BlogRow
		Comments       []*db.GetCommentsByThreadIdForUserRow
		IsReplyable    bool
		Text           string
		EditUrl        string
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
	}

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()

	data := Data{
		IsReplyable: true,
		EditUrl:     fmt.Sprintf("/blogs/blog/%d/edit", blogId),
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	blog, err := queries.GetBlogEntryForListerByID(r.Context(), db.GetBlogEntryForListerByIDParams{
		ListerID: uid,
		ID:       int32(blogId),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err == nil {
		cd.PageTitle = fmt.Sprintf("Blog %d Comments", blog.Idblogs)
		if blog.ForumthreadID.Valid {
			cd.SetCurrentThreadAndTopic(blog.ForumthreadID.Int32, 0)
		}
	}
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := templates.GetCompiledSiteTemplates(r.Context().Value(consts.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("getBlogEntryForListerByID_comments Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	if !(cd.HasGrant("blogs", "entry", "view", blog.Idblogs) ||
		cd.HasGrant("blogs", "entry", "reply", blog.Idblogs) ||
		cd.SelectedThreadCanReply()) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	canReply := cd.HasGrant("blogs", "entry", "reply", blog.Idblogs)

	data.IsReplyable = canReply

	editUrl := ""
	if uid == blog.UsersIdusers {
		editUrl = fmt.Sprintf("/blogs/blog/%d/edit", blog.Idblogs)
	}

	data.Blog = &BlogRow{
		GetBlogEntryForListerByIDRow: blog,
		EditUrl:                      editUrl,
	}

	threadRow, err := cd.SelectedThread()
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("GetThreadLastPosterAndPerms: %v", err)
		}
		data.IsReplyable = false
	} else if threadRow != nil && threadRow.Locked.Valid && threadRow.Locked.Bool {
		data.IsReplyable = false
	}

	replyType := r.URL.Query().Get("type")
	editCommentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	quoteId, _ := strconv.Atoi(r.URL.Query().Get("quote"))
	if blog.ForumthreadID.Valid {
		cd.SetCurrentThreadAndTopic(blog.ForumthreadID.Int32, 0)
		rows, err := cd.SectionThreadComments("blogs", "entry", blog.ForumthreadID.Int32)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("thread comments: %s", err)
				handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
				return
			}
		}
		data.Comments = rows

		data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
			return cd.CanEditAny() || cmt.IsOwner
		}
		data.EditURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
			if !data.CanEditComment(cmt) {
				return ""
			}
			return fmt.Sprintf("/blogs/blog/%d/comments?comment=%d#edit", blog.Idblogs, cmt.Idcomments)
		}
		data.EditSaveURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
			if !data.CanEditComment(cmt) {
				return ""
			}
			return fmt.Sprintf("/blogs/blog/%d/comment/%d", blog.Idblogs, cmt.Idcomments)
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
		if editCommentId != 0 {
			data.IsReplyable = false
		}

		if quoteId != 0 {
			comment, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
				ViewerID: uid,
				ID:       int32(quoteId),
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err == nil {
				switch replyType {
				case "full":
					data.Text = a4code.FullQuoteOf(comment.Username.String, comment.Text.String)
				default:
					data.Text = a4code.QuoteOfText(comment.Username.String, comment.Text.String)
				}
			}
		}
	}

	handlers.TemplateHandler(w, r, "blogs/commentPage.gohtml", data)
}
