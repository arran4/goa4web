package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
)

func CommentPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Text string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	blog, err := cd.BlogPost()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("BlogPost: %v", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	cd.PageTitle = fmt.Sprintf("Blog %d Comments", blog.Idblogs)
	if _, err := cd.BlogCommentThread(); err != nil && err != sql.ErrNoRows {
		log.Printf("BlogCommentThread: %v", err)
	}

	if !(cd.HasGrant("blogs", "entry", "view", blog.Idblogs) ||
		cd.HasGrant("blogs", "entry", "reply", blog.Idblogs) ||
		cd.SelectedThreadCanReply()) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data := Data{}
	replyType := r.URL.Query().Get("type")
	commentId, _ := strconv.Atoi(r.URL.Query().Get("comment"))
	if commentId != 0 {
		if comment, err := cd.CommentByID(int32(commentId)); err == nil {
			switch replyType {
			case "full":
				data.Text = a4code.QuoteText(comment.Username.String, comment.Text.String, a4code.WithFullQuote())
			default:
				data.Text = a4code.QuoteText(comment.Username.String, comment.Text.String)
			}
		}
	}

	handlers.TemplateHandler(w, r, "blogs/commentPage.gohtml", data)
}
