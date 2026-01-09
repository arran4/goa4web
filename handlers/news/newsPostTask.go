package news

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type newsPostTask struct {
}

const (
	NewsPostPageTmpl = "news/postPage.gohtml"
)

func NewNewsPostTask() tasks.Task {
	return &newsPostTask{}
}

func (t *newsPostTask) TemplatesRequired() []string {
	return []string{NewsPostPageTmpl}
}

func (t *newsPostTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *newsPostTask) Get(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Post           *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
		Thread         *db.GetThreadLastPosterAndPermsRow
		Comments       []*db.GetCommentsByThreadIdForUserRow
		ReplyText      string
		IsReplyable    bool
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
		Labels         []templates.TopicLabel
		PublicLabels   []templates.TopicLabel
		BackURL        string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "News"
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	data := Data{
		IsReplyable: true,
		BackURL:     r.URL.Path,
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
	var post *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
	for _, p := range posts {
		if p.Idsitenews == int32(pid) {
			post = p
			break
		}
	}
	if post == nil {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	if !cd.HasGrant("news", "post", "view", post.Idsitenews) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       strings.Split(post.News.String, "\n")[0],
		Description: a4code.Snip(post.News.String, 128),
		Image:       cd.AbsoluteURL("/static/default-og-image.png"),
		URL:         cd.AbsoluteURL(r.URL.String()),
	}

	replyType := r.URL.Query().Get("type")
	quoteId, _ := strconv.Atoi(r.URL.Query().Get("quote"))

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
			handlers.RedirectSeeOtherWithError(w, r, "", err)
			return
		}
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	common.WithOffset(offset)(cd)
	editCommentId, _ := strconv.Atoi(r.URL.Query().Get("editComment"))

	data.Comments = commentRows
	data.Thread = threadRow
	data.Post = post

	data.CanEditComment = func(cmt *db.GetCommentsByThreadIdForUserRow) bool {
		return cmt.IsOwner
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
				data.ReplyText = a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithFullQuote())
			default:
				data.ReplyText = a4code.QuoteText(c.Username.String, c.Text.String)
			}
		}
	}

	if als, err := cd.NewsAuthorLabels(post.Idsitenews); err == nil {
		for _, l := range als {
			tl := templates.TopicLabel{Name: l, Type: "author"}
			data.Labels = append(data.Labels, tl)
			data.PublicLabels = append(data.PublicLabels, tl)
		}
	}
	if pls, err := cd.NewsPrivateLabels(post.Idsitenews, post.UsersIdusers); err == nil {
		for _, l := range pls {
			data.Labels = append(data.Labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	}

	cd.CustomIndexItems = append(cd.CustomIndexItems, NewsPageSpecificItems(cd, r, post)...)

	if err := cd.ExecuteSiteTemplate(w, r, NewsPostPageTmpl, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
