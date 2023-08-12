package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func blogsCommentPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*show_blogRow
		IsEditable  bool
		IsReplyable bool
	}
	type BlogComment struct {
		*printThreadRow
		ShowReply          bool
		Editable           bool
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

	data.Blog = &BlogRow{
		show_blogRow: blog,
		IsEditable:   uid == blog.UsersIdusers,
		IsReplyable:  true,
	}

	if blog.ForumthreadIdforumthread > 0 { // TODO make nullable.

		replyType := r.URL.Query().Get("type")
		commentIdString := r.URL.Query().Get("comment")
		if commentIdString != "" {
			commentId, _ := strconv.Atoi(commentIdString)
			comment, err := queries.show_comment(r.Context(), int32(commentId))
			if err != nil {
				log.Printf("show_comment Error: %s", err)
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

		rows, err := queries.printThread(r.Context(), blog.ForumthreadIdforumthread)
		if err != nil {
			log.Printf("show_blog_comments Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for i, row := range rows {
			data.Comments = append(data.Comments, &BlogComment{
				printThreadRow: row,
				ShowReply:      true,
				Editable:       uid == row.Idusers,
				Offset:         i + offset,
				Idblogs:        blog.Idblogs,
			})
		}
	}

	CustomBlogIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsCommentPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
