package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
)

func writingsCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories                       []*Writingcategory
		CategoryBreadcrumbs              []*Writingcategory
		EditingCategoryId                int32 // TODO
		IsAdmin                          bool  // TODO
		IsWriter                         bool  // TODO
		Abstracts                        []*FetchPublicWritingsInCategoryRow
		WritingcategoryIdwritingcategory int32
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsWriter: true,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.FetchAllCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("fetchCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	writingsRows, err := queries.FetchPublicWritingsInCategory(r.Context(), 0)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("fetchPublicWritingsInCategory Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	categoryMap := map[int32]*Writingcategory{}
	for _, cat := range categoryRows {
		categoryMap[cat.Idwritingcategory] = cat
		if cat.WritingcategoryIdwritingcategory == 0 {
			data.Categories = append(data.Categories, cat)
		}
	}

	data.Abstracts = writingsRows

	CustomWritingsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsCategoriesPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
