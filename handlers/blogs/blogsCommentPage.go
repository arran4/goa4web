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
	type Data struct {
		Blog *db.GetBlogEntryForListerByIDRow
		Text string
	}

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()

	data := Data{}

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
			if _, err := cd.SelectedThread(); err != nil && err != sql.ErrNoRows {
				log.Printf("GetThreadLastPosterAndPerms: %v", err)
			}
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
		cd.HasGrant("blogs", "entry", "comment", blog.Idblogs) ||
		cd.SelectedThreadCanReply()) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data.Blog = blog

	if blog.ForumthreadID.Valid {
		cd.SetCurrentThreadAndTopic(blog.ForumthreadID.Int32, 0)
	}

	replyType := r.URL.Query().Get("type")
	commentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	if commentId != 0 {
		comment, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
			ViewerID: uid,
			ID:       int32(commentId),
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

	handlers.TemplateHandler(w, r, "blogs/commentPage.gohtml", data)
}
