package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers/share"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
)

func BlogPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Text     string
		Labels   []templates.TopicLabel
		BackURL  string
		ShareURL string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	blog, err := cd.BlogPost()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		default:
			log.Printf("BlogPost: %v", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	dateSuffix := ""
	if !blog.Written.IsZero() {
		dateSuffix = fmt.Sprintf(" - %s", cd.FormatLocalTime(blog.Written))
	}
	if blog.Username.Valid {
		cd.PageTitle = fmt.Sprintf("Blog by %s%s", blog.Username.String, dateSuffix)
	} else {
		cd.PageTitle = fmt.Sprintf("Blog %d%s", blog.Idblogs, dateSuffix)
	}

	desc := a4code.Snip(blog.Blog.String, 128)
	imageURL, _ := share.MakeImageURL(cd.AbsoluteURL(""), cd.PageTitle, desc, cd.ShareSignKey, false)
	cd.OpenGraph = &common.OpenGraph{
		Title:       cd.PageTitle,
		Description: a4code.Snip(blog.Blog.String, 128),
		Image:       imageURL,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
		Type:        "article",
	}

	if _, err := cd.BlogCommentThread(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("BlogCommentThread: %v", err)
	}

	data := Data{BackURL: r.URL.Path}
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
	if pls, err := cd.BlogPrivateLabels(blog.Idblogs, blog.UsersIdusers); err == nil {
		for _, l := range pls {
			data.Labels = append(data.Labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	}

	cd.CustomIndexItems = append(cd.CustomIndexItems, BlogsPageSpecificItems(cd, r)...)

	BlogsBlogPageTmpl.Handle(w, r, data)
}

const BlogsBlogPageTmpl handlers.Page = "blogs/blogPage.gohtml"
