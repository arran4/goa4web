package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
)

func imagebbsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Boards      []*Imageboard
		IsSubBoard  bool
		BoardNumber int
	}

	data := Data{
		CoreData:    r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsSubBoard:  false,
		BoardNumber: 0,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	subBoardRows, err := queries.GetAllBoardsByParentBoardId(r.Context(), 0)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Boards = subBoardRows

	CustomImageBBSIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "imagebbsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomImageBBSIndex(data *CoreData, r *http.Request) {

	data.RSSFeedUrl = "/imagebbs/rss"
	data.AtomFeedUrl = "/imagebbs/atom"

	userHasAdmin := data.HasRole("administrator")
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Admin",
			Link: "/admin",
		}, IndexItem{
			Name: "Modify Boards",
			Link: "/imagebbs/admin/boards",
		}, IndexItem{
			Name: "New Board",
			Link: "/imagebbs/admin/board",
		})
	}
}
