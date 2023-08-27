package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func linkerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Offset      int
		CatId       int
		CommentOnId int
		ReplyToId   int
		Links       []*showLatestRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	data.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	data.CatId, _ = strconv.Atoi(r.URL.Query().Get("category"))
	data.CommentOnId, _ = strconv.Atoi(r.URL.Query().Get("comment"))
	data.ReplyToId, _ = strconv.Atoi(r.URL.Query().Get("reply"))

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	linkerPosts, err := queries.showLatest(r.Context(), int32(data.CatId))
	if err != nil {
		log.Printf("showLatest Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Links = linkerPosts

	CustomLinkerIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomLinkerIndex(data *CoreData, r *http.Request) {
	data.RSSFeedUrl = "/linker/rss"
	data.AtomFeedUrl = "/linker/atom"

	userHasAdmin := true // TODO
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "User Permissions",
			Link: "/linker/admin/users/levels",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Category Controls",
			Link: "/linker/admin/categories",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Approve links",
			Link: "/linker/admin/queue",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Add link",
			Link: "/linker/admin/add",
		})
	}
	vars := mux.Vars(r)
	categoryId, _ := vars["category"]
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if categoryId == "" {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/linker?offset=%d", offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/linker?offset=%d", offset-15),
			})
		}
	} else {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/linker/category/%s?offset=%d", categoryId, offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/linker/category/%s?offset=%d", categoryId, offset-15),
			})
		}
	}

}
