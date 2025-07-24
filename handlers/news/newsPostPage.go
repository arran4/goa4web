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
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type NewsPost struct {
	ShowReply bool
	// ShowEdit is true when the current user can modify the post. Users with
	// the writer, moderator or administrator role are permitted to edit.
	ShowEdit bool
}

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
	}
	type Post struct {
		*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		ShowReply    bool
		ShowEdit     bool
		Editing      bool
		Announcement *db.SiteAnnouncement
	}
	type Data struct {
		Post               *Post
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

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		IsReplying:         r.URL.Query().Has("comment"),
		IsReplyable:        true,
		SelectedLanguageId: cd.PreferredLanguageID(config.AppRuntimeConfig.DefaultLanguage),
	}
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", cd); err != nil {
				log.Printf("render no access page: %v", err)
			}
			return
		default:
			log.Printf("GetNewsPostByIdWithWriterIdAndThreadCommentCountForUser Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		})
	}

	data.Thread = threadRow
	ann, err := cd.NewsAnnouncement(post.Idsitenews)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("announcementForNews: %v", err)
	}
	data.Post = &Post{
		GetNewsPostByIdWithWriterIdAndThreadCommentCountRow: post,
		ShowReply:    cd.UserID != 0,
		ShowEdit:     canEditNewsPost(cd, post.Idsitenews),
		Editing:      editingId == int(post.Idsitenews),
		Announcement: ann,
	}

	handlers.TemplateHandler(w, r, "postPage.gohtml", data)
}
