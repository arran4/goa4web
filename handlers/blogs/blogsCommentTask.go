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
	"github.com/arran4/goa4web/internal/tasks"
)

type blogsCommentTask struct {
}

const (
	BlogsCommentPageTmpl = "blogs/commentPage.gohtml"
)

func NewBlogsCommentTask() tasks.Task {
	return &blogsCommentTask{}
}

func (t *blogsCommentTask) TemplatesRequired() []string {
	return []string{BlogsCommentPageTmpl}
}

func (t *blogsCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *blogsCommentTask) Get(w http.ResponseWriter, r *http.Request) {
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
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
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
		fmt.Println("TODO: FIx: Add enforced Access in router rather than task")
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	data := Data{BackURL: r.URL.RequestURI()}
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

	if err := cd.ExecuteSiteTemplate(w, r, BlogsCommentPageTmpl, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
