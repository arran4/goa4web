package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func writingsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories                       []*Writingcategory
		EditingCategoryId                int32 // TODO
		CategoryId                       int32 // TODO
		WritingcategoryIdwritingcategory int32 // TODO
		IsAdmin                          int32 // TODO
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.GetAllWritingCategories(r.Context(), data.CategoryId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllWritingCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Categories = categoryRows

	CustomWritingsIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "writingsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomWritingsIndex(data *CoreData, r *http.Request) {
	// TODO
	// TODO RSS
	userHasAdmin := true // TODO
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "User Permissions",
			Link: "/writings/user/permissions",
		})
	}
	userHasWriter := true // TODO
	if userHasWriter || userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Write writings",
			Link: "/writings/add",
		})
	}

	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "Return to list",
		Link: fmt.Sprintf("/writings?offset=%d", 0),
	})
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset != 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "The start",
			Link: fmt.Sprintf("/writings?offset=%d", 0),
		})
	}
	data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
		Name: "Next 10",
		Link: fmt.Sprintf("/writings?offset=%d", offset+10),
	})
	if offset > 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Previous 10",
			Link: fmt.Sprintf("/writings?offset=%d", offset-10),
		})
	}
}
