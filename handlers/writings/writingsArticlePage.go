package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
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
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}
	if writing.Title.Valid {
		cd.PageTitle = fmt.Sprintf("Writing: %s", writing.Title.String)
	} else {
		cd.PageTitle = fmt.Sprintf("Writing %d", writing.Idwriting)
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
	data.CanEdit = cd.HasContentWriterRole() && data.IsAuthor

	if quoteId != 0 {
		if c, err := cd.CommentByID(int32(quoteId)); err == nil && c != nil {
			switch replyType {
			case "full":
				data.ReplyText = a4code.FullQuoteOf(c.Username.String, c.Text.String)
			default:
				data.ReplyText = a4code.QuoteOfText(c.Username.String, c.Text.String)
			}
		}
	}

	handlers.TemplateHandler(w, r, "articlePage.gohtml", data)
}

func ArticleReplyActionPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	writing, err := cd.Article()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
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

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["target"] = notifications.Target{Type: "writing", ID: writing.Idwriting}
		}
	}

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))

	cid, threadID, topicID, err := cd.CreateWritingReply(writing, int32(languageId), text)
	if err != nil {
		log.Printf("Error: createComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if cid == 0 {
		http.Redirect(w, r, "?error="+"failed to create comment", http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(cid), ThreadID: threadID, TopicID: topicID}
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: text}
		}
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
