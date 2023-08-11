package main

import (
	"github.com/gorilla/mux"
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
		ShowReply bool
		Editable  bool
		Offset    int
		Idblogs   int32
	}
	type Data struct {
		*CoreData
		Blog        *BlogRow
		Comments    []*BlogComment
		Offset      int
		IsReplyable bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	data := Data{
		CoreData:    r.Context().Value(ContextValues("coreData")).(*CoreData),
		Offset:      offset,
		IsReplyable: true,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])

	blog, err := queries.show_blog(r.Context(), int32(blogId))
	if err != nil {
		log.Printf("show_blog_comments Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Blog = &BlogRow{
		show_blogRow: blog,
		IsEditable:   true, // TODO
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
			Editable:       true,
			Offset:         i + offset,
			Idblogs:        blog.Idblogs,
		})
	}

	CustomIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsCommentPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
