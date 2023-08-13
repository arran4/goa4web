package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func blogsCommentEditPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*show_blogRow
		EditUrl     string
		IsReplyable bool
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
	}
	type Data struct {
		*CoreData
		Blog               *BlogRow
		Comments           []*BlogComment
		Offset             int
		IsReplyable        bool
		Text               string
		Languages          []*Language
		SelectedLanguageId int
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	data := Data{
		CoreData:           r.Context().Value(ContextValues("coreData")).(*CoreData),
		Offset:             offset,
		IsReplyable:        true,
		SelectedLanguageId: 1,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	languageRows, err := queries.fetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])
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
		IsReplyable:  true,
	}

	if blog.ForumthreadIdforumthread > 0 { // TODO make nullable.

		replyType := r.URL.Query().Get("type")
		commentIdString := vars["comment"]
		commentId, _ := strconv.Atoi(commentIdString)
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
			log.Printf("show_blog_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for i, row := range rows {
			editUrl := ""
			if uid == row.UsersIdusers {
				editUrl = fmt.Sprintf("/blogs/blog/%d/comment/%d/edit#edit", blog.Idblogs, row.Idcomments)
			}

			data.Comments = append(data.Comments, &BlogComment{
				user_get_all_comments_for_threadRow: row,
				ShowReply:                           true,
				EditUrl:                             editUrl,
				Editing:                             commentId != 0 && int32(commentId) == row.Idcomments,
				Offset:                              i + offset,
				Idblogs:                             blog.Idblogs,
				Languages:                           languageRows,
				SelectedLanguageId:                  row.LanguageIdlanguage,
			})
		}
	}

	CustomBlogIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsCommentEditPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsCommentEditPostPage(w http.ResponseWriter, r *http.Request) {

	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])
	commentId, _ := strconv.Atoi(vars["comment"])

	err = queries.update_comment(r.Context(), update_commentParams{
		Idcomments:         int32(commentId),
		LanguageIdlanguage: int32(languageId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/blogs/blog/%d/comments", blogId), http.StatusTemporaryRedirect)

}
