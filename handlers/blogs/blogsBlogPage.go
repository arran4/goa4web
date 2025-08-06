package blogs

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
)

func BlogPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*db.GetBlogEntryForListerByIDRow
		EditUrl     string
		IsReplyable bool
	}
	type Data struct {
		Blog           *BlogRow
		Comments       []*db.GetCommentsByThreadIdForUserRow
		IsReplyable    bool
		Text           string
		EditUrl        string
		CanReply       bool
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	queries := cd.Queries()
	data := Data{
		IsReplyable: true,
		EditUrl:     fmt.Sprintf("/blogs/blog/%d/edit", blogId),
		CanReply:    cd.UserID != 0,
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
		if blog.Username.Valid {
			cd.PageTitle = fmt.Sprintf("Blog by %s", blog.Username.String)
		} else {
			cd.PageTitle = fmt.Sprintf("Blog %d", blog.Idblogs)
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
	if !cd.HasGrant("blogs", "entry", "view", blog.Idblogs) {
		if err := templates.GetCompiledSiteTemplates(r.Context().Value(consts.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", struct{}{}); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}

	editUrl := ""
	if uid == blog.UsersIdusers {
		editUrl = fmt.Sprintf("/blogs/blog/%d/edit", blog.Idblogs)
	}

	data.Blog = &BlogRow{
		GetBlogEntryForListerByIDRow: blog,
		EditUrl:                      editUrl,
		IsReplyable:                  true,
	}

	if !blog.ForumthreadID.Valid {
		data.IsReplyable = false
		data.Blog.IsReplyable = false
	} else {
		threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
			ViewerID:      uid,
			ThreadID:      blog.ForumthreadID.Int32,
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("GetThreadLastPosterAndPerms: %v", err)
			}
			data.IsReplyable = false
			data.Blog.IsReplyable = false
		} else if threadRow.Locked.Valid && threadRow.Locked.Bool {
			data.IsReplyable = false
			data.Blog.IsReplyable = false
		}

		cd.SetCurrentThreadAndTopic(blog.ForumthreadID.Int32, 0)
		rows, err := cd.SectionThreadComments("blogs", "entry", blog.ForumthreadID.Int32)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("thread comments: %s", err)
				handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
				return
			}
		}

		data.Comments = rows
		commentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
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
	}

	handlers.TemplateHandler(w, r, "blogPage.gohtml", data)
}
