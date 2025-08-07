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

func BlogPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Blog *db.GetBlogEntryForListerByIDRow
		Text string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	vars := mux.Vars(r)
	blogID, _ := strconv.Atoi(vars["blog"])

	queries := cd.Queries()
	data := Data{}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	blog, err := queries.GetBlogEntryForListerByID(r.Context(), db.GetBlogEntryForListerByIDParams{
		ListerID: uid,
		ID:       int32(blogID),
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
			if err := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", struct{}{}); err != nil {
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
		if err := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", struct{}{}); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}

	data.Blog = blog

	if blog.ForumthreadID.Valid {
		cd.SetCurrentThreadAndTopic(blog.ForumthreadID.Int32, 0)
	}

	quoteID, _ := strconv.Atoi(r.URL.Query().Get("quote"))
	replyType := r.URL.Query().Get("type")
	if quoteID != 0 {
		comment, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
			ViewerID: uid,
			ID:       int32(quoteID),
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

	handlers.TemplateHandler(w, r, "blogPage.gohtml", data)
}
