package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func newsPage(w http.ResponseWriter, r *http.Request) {
	type Post struct {
		*getLatestNewsPostsRow
		ShowReply bool
		ShowEdit  bool
		Editing   bool
	}
	type Data struct {
		*CoreData
		News []*Post
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	posts, err := queries.getLatestNewsPosts(r.Context())
	if err != nil {
		log.Printf("getLatestNewsPosts Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	editingId, _ := strconv.Atoi(r.URL.Query().Get("reply"))

	for _, post := range posts {
		data.News = append(data.News, &Post{
			getLatestNewsPostsRow: post,
			ShowReply:             true, // TODO
			ShowEdit:              true, // TODO
			Editing:               editingId == int(post.Idsitenews),
		})
	}

	CustomNewsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "newsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomNewsIndex(data *CoreData, r *http.Request) {
	// TODO
	// TODO RSS
	userHasAdmin := true // TODO
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "User Permissions",
			Link: "/news/user/permissions",
		})
	}
	userHasWriter := true // TODO
	if userHasWriter || userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Add News",
			Link: "/news/post",
		})
	}

	vars := mux.Vars(r)
	newsId, _ := vars["news"]
	if newsId != "" {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Return to list",
			Link: fmt.Sprintf("/?offset=%d", 0),
		})
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if offset != 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "The start",
				Link: fmt.Sprintf("/?offset=%d", 0),
			})
		}
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 10",
			Link: fmt.Sprintf("/?offset=%d", offset+10),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 10",
				Link: fmt.Sprintf("/?offset=%d", offset-10),
			})
		}
	}
}
