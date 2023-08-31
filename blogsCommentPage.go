package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func blogsCommentPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*show_blogRow
		EditUrl string
	}
	type BlogComment struct {
		*user_get_all_comments_for_threadRow
		ShowReply          bool
		EditUrl            string
		Editing            bool
		Offset             int
		Idblogs            int32
		Languages          []*Language
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
		Languages          []*Language
		SelectedLanguageId int
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	data := Data{
		CoreData:           r.Context().Value(ContextValues("coreData")).(*CoreData),
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: 1,
		EditUrl:            fmt.Sprintf("/blogs/blog/%d/edit", blogId),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	blog, err := queries.show_blog(r.Context(), int32(blogId))
	if err != nil {
		log.Printf("show_blog_comments Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	editUrl := ""
	if uid == blog.UsersIdusers {
		editUrl = fmt.Sprintf("/blogs/blog/%d/edit", blog.Idblogs)
	}

	data.Blog = &BlogRow{
		show_blogRow: blog,
		EditUrl:      editUrl,
	}

	replyType := r.URL.Query().Get("type")
	commentIdString := r.URL.Query().Get("comment")
	commentId, _ := strconv.Atoi(commentIdString)
	if blog.ForumthreadIdforumthread > 0 { // TODO make nullable.

		if commentIdString != "" {
			comment, err := queries.user_get_comment(r.Context(), user_get_commentParams{
				UsersIdusers: uid,
				Idcomments:   int32(commentId),
			})
			if err != nil {
				log.Printf("user_get_comment Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			switch replyType {
			case "full":
				data.Text = processCommentFullQuote(comment.Username.String, comment.Text.String)
			default:
				data.Text = processCommentQuote(comment.Username.String, comment.Text.String)
			}
		}

		rows, err := queries.user_get_all_comments_for_thread(r.Context(), user_get_all_comments_for_threadParams{
			UsersIdusers:             uid,
			ForumthreadIdforumthread: blog.ForumthreadIdforumthread,
		})
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("user_get_all_comments_for_thread Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		for i, row := range rows {
			editUrl := ""
			editSaveUrl := ""
			if uid == row.UsersIdusers {
				editUrl = fmt.Sprintf("/blogs/blog/%d/comments?comment=%d#edit", blog.Idblogs, row.Idcomments)
				editSaveUrl = fmt.Sprintf("/blogs/blog/%d/comment/%d", blog.Idblogs, row.Idcomments)
				if commentId != 0 && int32(commentId) == row.Idcomments {
					data.IsReplyable = false
				}
			}
			data.Comments = append(data.Comments, &BlogComment{
				user_get_all_comments_for_threadRow: row,
				ShowReply:                           true,
				EditUrl:                             editUrl,
				EditSaveUrl:                         editSaveUrl,
				Editing:                             commentId != 0 && int32(commentId) == row.Idcomments,
				Offset:                              i + offset,
				Idblogs:                             blog.Idblogs,
				Languages:                           languageRows,
				SelectedLanguageId:                  row.LanguageIdlanguage,
			})
		}
	}

	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "blogsCommentPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
