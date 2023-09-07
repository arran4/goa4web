package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func blogsBlogPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*Show_blogRow
		EditUrl     string
		IsReplyable bool
	}
	type BlogComment struct {
		*User_get_all_comments_for_threadRow
		ShowReply bool
		EditUrl   string
		Editing   bool
		Offset    int
		Idblogs   int32
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

	languageRows, err := queries.FetchLanguages(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Languages = languageRows

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	blog, err := queries.Show_blog(r.Context(), int32(blogId))
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
		Show_blogRow: blog,
		EditUrl:      editUrl,
		IsReplyable:  true, // TODO
	}

	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "blogsBlogPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
