package news

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
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type newsPostTask struct {
}

var _ tasks.Task = (*newsPostTask)(nil)

const (
	NewsPostPageTmpl tasks.Template = "news/postPage.gohtml"
)

func NewNewsPostTask() tasks.Task {
	return &newsPostTask{}
}

func (t *newsPostTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{NewsPostPageTmpl}
}

func (t *newsPostTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

type newsPostPageData struct {
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
	ShareURL       string
}

func (t *newsPostTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "News"
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	data := newsPostPageData{
		IsReplyable: true,
		BackURL:     r.URL.Path,
	}
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["news"])
	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)

	post, err := getNewsPost(cd, pid)
	if err != nil {
		if errors.Is(err, handlers.ErrForbidden) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		} else {
			log.Printf("LatestNewsList: %v", err)
			handlers.RenderErrorPage(w, r, err)
		}
		return
	}

	setupNewsOpenGraph(cd, post, r.URL.String())

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

	setupCommentHelpers(&data, cd, pid, editCommentId)

	if editCommentId != 0 {
		data.IsReplyable = false
	}

	data.ReplyText = getReplyText(cd, quoteId, replyType)

	data.Labels, data.PublicLabels = fetchNewsLabels(cd, post)

	cd.CustomIndexItems = append(cd.CustomIndexItems, NewsPageSpecificItems(cd, r, post)...)

	NewsPostPageTmpl.Handle(w, r, data)
}

func setupNewsOpenGraph(cd *common.CoreData, post *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, reqURL string) {
	if post.Occurred.Valid {
		cd.PageTitle = fmt.Sprintf("News - %s", cd.FormatLocalTime(post.Occurred.Time))
	}

	desc := a4code.SnipText(post.News.String, 128)
	imgURL, err := share.MakeImageURL(cd.AbsoluteURL(""), cd.PageTitle, desc, cd.ShareSignKey, false)
	if err != nil {
		log.Printf("Error making image URL: %v", err)
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       cd.PageTitle,
		Description: desc,
		Image:       imgURL,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(reqURL),
		Type:        "article",
	}
}

func getNewsPost(cd *common.CoreData, pid int) (*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	posts, err := cd.LatestNewsList(0, 50)
	if err != nil {
		return nil, err
	}
	for _, p := range posts {
		if p.Idsitenews == int32(pid) {
			return p, nil
		}
	}
	return nil, handlers.ErrForbidden
}

func setupCommentHelpers(data *newsPostPageData, cd *common.CoreData, pid int, editCommentId int) {
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
}

func getReplyText(cd *common.CoreData, quoteId int, replyType string) string {
	if quoteId == 0 {
		return ""
	}
	c, err := cd.CommentByID(int32(quoteId))
	if err != nil || c == nil {
		return ""
	}
	if replyType == "full" {
		return a4code.QuoteText(c.Username.String, c.Text.String, a4code.WithFullQuote())
	}
	return a4code.QuoteText(c.Username.String, c.Text.String)
}

func fetchNewsLabels(cd *common.CoreData, post *db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow) ([]templates.TopicLabel, []templates.TopicLabel) {
	var labels []templates.TopicLabel
	var publicLabels []templates.TopicLabel

	if als, err := cd.NewsAuthorLabels(post.Idsitenews); err == nil {
		for _, l := range als {
			tl := templates.TopicLabel{Name: l, Type: "author"}
			labels = append(labels, tl)
			publicLabels = append(publicLabels, tl)
		}
	}
	if pls, err := cd.NewsPrivateLabels(post.Idsitenews, post.UsersIdusers); err == nil {
		for _, l := range pls {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	}
	return labels, publicLabels
}
