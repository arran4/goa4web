package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers/share"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
)

func ArticlePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Request        *http.Request
		Comments       []*db.GetCommentsByThreadIdForUserRow
		CanEdit        bool
		IsAuthor       bool
		ReplyText      string
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
		Labels         []templates.TopicLabel
		BackURL        string
		ShareURL       string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writing"
	cd.LoadSelectionsFromRequest(r)
	writing, err := cd.Article()
	if err != nil {
		log.Printf("get writing: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	if writing == nil {
		log.Printf("get writing: no writing found")
		handlers.RenderErrorPage(w, r, fmt.Errorf("No writing found"))
		return
	}
	cd.SetCurrentThreadAndTopic(writing.ForumthreadID, 0)
	if !(cd.HasGrant("writing", "article", "view", writing.Idwriting) || cd.SelectedThreadCanReply()) {
		fmt.Println("TODO: FIx: Add enforced Access in router rather than task")
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	if writing.Title.Valid {
		cd.PageTitle = fmt.Sprintf("Writing: %s", writing.Title.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Writing %d", writing.Idwriting)
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       writing.Title.String,
		Description: a4code.Snip(writing.Abstract.String, 128),
		Image:       share.MakeImageURL(cd.AbsoluteURL(), writing.Title.String, cd.ShareSignKey, false),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	common.WithOffset(offset)(cd)
	editCommentId, _ := strconv.Atoi(r.URL.Query().Get("editComment"))
	replyType := r.URL.Query().Get("type")
	quoteId, _ := strconv.Atoi(r.URL.Query().Get("quote"))

	comments, err := cd.ArticleComments()
	if err != nil {
		log.Printf("thread comments: %v", err)
	}
	data := Data{
		Request:  r,
		Comments: comments,
		BackURL:  r.URL.Path,
	}

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
		return fmt.Sprintf("/writings/article/%d/comment/%d", writing.Idwriting, cmt.Idcomments)
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

	data.IsAuthor = writing.UsersIdusers == cd.UserID
	data.CanEdit = cd.HasGrant("writing", "article", "edit", writing.Idwriting)

	if als, err := cd.WritingAuthorLabels(writing.Idwriting); err == nil {
		for _, l := range als {
			data.Labels = append(data.Labels, templates.TopicLabel{Name: l, Type: "author"})
		}
	}
	if pls, err := cd.WritingPrivateLabels(writing.Idwriting, writing.Writerid); err == nil {
		for _, l := range pls {
			data.Labels = append(data.Labels, templates.TopicLabel{Name: l, Type: "private"})
		}
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

	cd.CustomIndexItems = append(cd.CustomIndexItems, WritingsPageSpecificItems(cd, r)...)

	ArticlePageTmpl.Handle(w, r, data)
}

const ArticlePageTmpl handlers.Page = "writings/articlePage.gohtml"

func ArticleReplyActionPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	writing, err := cd.Article()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := AdminNoAccessPageTmpl.Handle(w, r, struct{}{}); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("get writing: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	if writing == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	if !cd.HasGrant("writing", "article", "reply", writing.Idwriting) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	cid, threadID, topicID, err := cd.CreateWritingReply(writing, int32(languageId), text)
	if err != nil {
		log.Printf("Error: createComment: %s", err)
		handlers.RedirectSeeOtherWithMessage(w, r, "", err.Error())
		return
	}
	if cid == 0 {
		handlers.RedirectSeeOtherWithMessage(w, r, "", "failed to create comment")
		return
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             threadID,
		TopicID:              topicID,
		CommentID:            int32(cid),
		LabelItem:            "writing",
		LabelItemID:          writing.Idwriting,
		CommentText:          text,
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
		IncludeSearch:        true,
		AdditionalData: map[string]any{
			"target": notifications.Target{Type: "writing", ID: writing.Idwriting},
		},
	}); err != nil {
		log.Printf("writing article reply side effects: %v", err)
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}

const AdminNoAccessPageTmpl handlers.Page = "admin/noAccessPage.gohtml"
