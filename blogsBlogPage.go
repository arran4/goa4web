package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func blogsBlogPage(w http.ResponseWriter, r *http.Request) {
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
	}
	type Data struct {
		*CoreData
		Blog        *BlogRow
		Comments    []*BlogComment
		Offset      int
		IsReplyable bool
		Text        string
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
		IsReplyable:  true, // TODO
	}

	CustomBlogIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsBlogPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsBlogReplyPostPage(w http.ResponseWriter, r *http.Request) {
}
