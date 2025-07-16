package blogs

import (
	"database/sql"
	"errors"
	"fmt"

	db "github.com/arran4/goa4web/internal/db"

	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/a4code"
	corecommon "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	hcommon "github.com/arran4/goa4web/handlers/common"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func CommentPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*db.GetBlogEntryForUserByIdRow
		EditUrl string
	}
	type BlogComment struct {
		*db.GetCommentsByThreadIdForUserRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Idblogs            int32
		Languages          []*db.Language
		SelectedLanguageId int32
		EditSaveUrl        string
	}
	type Data struct {
		*CoreData
		Blog               *BlogRow
		Comments           []*BlogComment
		Offset             int
		IsReplyable        bool
		Text               string
		EditUrl            string
		Languages          []*db.Language
		SelectedLanguageId int
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := Data{
		CoreData:           r.Context().Value(hcommon.KeyCoreData).(*CoreData),
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: int(corelanguage.ResolveDefaultLanguageID(r.Context(), queries, config.AppRuntimeConfig.DefaultLanguage)),
		EditUrl:            fmt.Sprintf("/blogs/blog/%d/edit", blogId),
	}

	languageRows, err := data.CoreData.Languages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	blog, err := queries.GetBlogEntryForUserById(r.Context(), db.GetBlogEntryForUserByIdParams{
		ViewerIdusers: uid,
		ID:            int32(blogId),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_ = templates.GetCompiledSiteTemplates(r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
			return
		default:
			log.Printf("getBlogEntryForUserById_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	editUrl := ""
	if uid == blog.UsersIdusers {
		editUrl = fmt.Sprintf("/blogs/blog/%d/edit", blog.Idblogs)
	}

	data.Blog = &BlogRow{
		GetBlogEntryForUserByIdRow: blog,
		EditUrl:                    editUrl,
	}

	if !blog.ForumthreadID.Valid {
		data.IsReplyable = false
	} else {
		threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
			ViewerID:      uid,
			ThreadID:      blog.ForumthreadID.Int32,
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("GetThreadLastPosterAndPerms: %v", err)
			}
			data.IsReplyable = false
		} else if threadRow.Locked.Valid && threadRow.Locked.Bool {
			data.IsReplyable = false
		}
	}

	replyType := r.URL.Query().Get("type")
	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	if blog.ForumthreadID.Valid {
		pthid := blog.ForumthreadID.Int32

		if commentIdString != "" {
			comment, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
				ViewerID: uid,
				ID:       int32(commentId),
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				log.Printf("getCommentByIdForUser Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			switch replyType {
			case "full":
				data.Text = a4code.FullQuoteOf(comment.Username.String, comment.Text.String)
			default:
				data.Text = a4code.QuoteOfText(comment.Username.String, comment.Text.String)
			}
		}

		rows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
			ViewerID: uid,
			ThreadID: pthid,
			UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("getCommentsByThreadIdForUser Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		for i, row := range rows {
			editUrl := ""
			editSaveUrl := ""
			if data.CoreData.CanEditAny() || row.IsOwner {
				editUrl = fmt.Sprintf("/blogs/blog/%d/comments?comment=%d#edit", blog.Idblogs, row.Idcomments)
				editSaveUrl = fmt.Sprintf("/blogs/blog/%d/comment/%d", blog.Idblogs, row.Idcomments)
				if commentId != 0 && int32(commentId) == row.Idcomments {
					data.IsReplyable = false
				}
			}
			data.Comments = append(data.Comments, &BlogComment{
				GetCommentsByThreadIdForUserRow: row,
				ShowReply:                       true,
				EditUrl:                         editUrl,
				EditSaveUrl:                     editSaveUrl,
				Editing:                         commentId != 0 && int32(commentId) == row.Idcomments,
				Offset:                          i + offset,
				Idblogs:                         blog.Idblogs,
				Languages:                       languageRows,
				SelectedLanguageId:              row.LanguageIdlanguage,
			})
		}
	}

	hcommon.TemplateHandler(w, r, "commentPage.gohtml", data)
}
