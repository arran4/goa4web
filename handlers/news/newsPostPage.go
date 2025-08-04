package news

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func NewsPostPage(w http.ResponseWriter, r *http.Request) {
	type CommentPlus struct {
		*db.GetCommentsByThreadIdForUserRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Languages          []*db.Language
		SelectedLanguageId int
		EditSaveUrl        string
		AdminUrl           string
	}
	type Data struct {
		Post               *common.NewsPost
		Languages          []*db.Language
		SelectedLanguageId int32
		Topic              *db.Forumtopic
		Comments           []*CommentPlus
		Offset             int
		IsReplying         bool
		IsReplyable        bool
		Thread             *db.GetThreadLastPosterAndPermsRow
		ReplyText          string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "News"
	queries := cd.Queries()
	data := Data{
		IsReplying:         r.URL.Query().Has("comment"),
		IsReplyable:        true,
		SelectedLanguageId: cd.PreferredLanguageID(cd.Config.DefaultLanguage),
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
	var post *common.NewsPost
	for _, p := range posts {
		if p.Idsitenews == int32(pid) {
			post = p
			break
		}
	}
	if post == nil {
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}
	if !cd.HasGrant("news", "post", "view", post.Idsitenews) {
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}

	editingId, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	replyType := r.URL.Query().Get("type")

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: int32(post.ForumthreadID),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
			handlers.RenderErrorPage(w, r, err)
			return
		}
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
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	cd = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	languageRows, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	data.Languages = languageRows

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)

	editCommentIdString := r.URL.Query().Get("editComment")
	editCommentId, _ := strconv.Atoi(editCommentIdString)
	for i, row := range commentRows {
		editUrl := ""
		editSaveUrl := ""
		if cd.CanEditAny() || row.IsOwner {
			editUrl = fmt.Sprintf("?editComment=%d#edit", row.Idcomments)
			editSaveUrl = fmt.Sprintf("/news/news/%d/comment/%d", pid, row.Idcomments)
			if commentId != 0 && int32(commentId) == row.Idcomments {
				data.IsReplyable = false
			}
		}

		if int32(commentId) == row.Idcomments {
			switch replyType {
			case "full":
				data.ReplyText = a4code.FullQuoteOf(row.Posterusername.String, row.Text.String)
			default:
				data.ReplyText = a4code.QuoteOfText(row.Posterusername.String, row.Text.String)
			}
		}

		data.Comments = append(data.Comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       cd.UserID != 0,
			EditUrl:                         editUrl,
			EditSaveUrl:                     editSaveUrl,
			Editing:                         editCommentId != 0 && int32(editCommentId) == row.Idcomments,
			Offset:                          i + offset,
			Languages:                       languageRows,
			SelectedLanguageId:              int(row.LanguageIdlanguage),
			AdminUrl: func() string {
				if cd.HasRole("administrator") {
					return fmt.Sprintf("/admin/comment/%d", row.Idcomments)
				} else {
					return ""
				}
			}(),
		})
	}

	data.Thread = threadRow
	post.Editing = editingId == int(post.Idsitenews)
	data.Post = post

	handlers.TemplateHandler(w, r, "postPage.gohtml", data)
}
