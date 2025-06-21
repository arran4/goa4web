package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"strconv"
)

func writingsCategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories                       []*Writingcategory
		CategoryBreadcrumbs              []*Writingcategory
		EditingCategoryId                int32
		CategoryId                       int32
		WritingcategoryIdwritingcategory int32
		IsAdmin                          bool
		IsWriter                         bool
		Abstracts                        []*GetPublicWritingsInCategoryRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	data.IsAdmin = data.CoreData.HasRole("administrator")
	data.IsWriter = data.CoreData.HasRole("writer") || data.IsAdmin
	editID, _ := strconv.Atoi(r.URL.Query().Get("edit"))
	data.EditingCategoryId = int32(editID)

	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])
	data.CategoryId = int32(categoryId)
	data.WritingcategoryIdwritingcategory = data.CategoryId

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.FetchAllCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllWritingCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	writingsRows, err := queries.GetPublicWritingsInCategory(r.Context(), GetPublicWritingsInCategoryParams{
		WritingcategoryIdwritingcategory: data.CategoryId,
		Limit:                            15,
		Offset:                           0,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:

			log.Printf("getPublicWritingsInCategory Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	categoryMap := map[int32]*Writingcategory{}
	for _, cat := range categoryRows {
		categoryMap[cat.Idwritingcategory] = cat
		if cat.WritingcategoryIdwritingcategory == data.CategoryId {
			data.Categories = append(data.Categories, cat)
		}
	}
	for cid := data.CategoryId; len(data.CategoryBreadcrumbs) < len(categoryRows); {
		cat, ok := categoryMap[cid]
		if ok {
			data.CategoryBreadcrumbs = append(data.CategoryBreadcrumbs, cat)
			cid = cat.WritingcategoryIdwritingcategory
		} else {
			break
		}
	}
	slices.Reverse(data.CategoryBreadcrumbs)
	data.Abstracts = writingsRows

	CustomWritingsIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "writingsCategoryPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
