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

func BlogPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Text    string
		Labels  []templates.TopicLabel
		BackURL string
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
	if blog.Username.Valid {
		cd.PageTitle = fmt.Sprintf("Blog by %s", blog.Username.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Blog %d", blog.Idblogs)
	}
	if _, err := cd.BlogCommentThread(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("BlogCommentThread: %v", err)
	}

	data := Data{BackURL: r.URL.RequestURI()}
	quoteID, _ := strconv.Atoi(r.URL.Query().Get("quote"))
	replyType := r.URL.Query().Get("type")
	if quoteID != 0 {
		if comment, err := cd.CommentByID(int32(quoteID)); err == nil {
			switch replyType {
			case "full":
				data.Text = a4code.QuoteText(comment.Username.String, comment.Text.String, a4code.WithFullQuote())
			default:
				data.Text = a4code.QuoteText(comment.Username.String, comment.Text.String)
			}
		}
	}

	if als, err := cd.BlogAuthorLabels(blog.Idblogs); err == nil {
		for _, l := range als {
			data.Labels = append(data.Labels, templates.TopicLabel{Name: l, Type: "author"})
		}
	}
	if pls, err := cd.BlogPrivateLabels(blog.Idblogs); err == nil {
		for _, l := range pls {
			data.Labels = append(data.Labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	}

	handlers.TemplateHandler(w, r, "blogPage.gohtml", data)
}
