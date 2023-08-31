package main

import (
	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"strconv"
)

func writingsCategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories          []*Writingcategory
		CategoryBreadcrumbs []*Writingcategory
		EditingCategoryId   int32 // TODO
		CategoryId          int32 // TODO
		IsAdmin             bool  // TODO
		IsWriter            bool  // TODO
		Abstracts           []*fetchPublicWritingsInCategoryRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsWriter: true,
	}

	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])
	data.CategoryId = int32(categoryId)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.fetchAllCategories(r.Context())
	if err != nil {
		log.Printf("fetchCategories Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writingsRows, err := queries.fetchPublicWritingsInCategory(r.Context(), data.CategoryId)
	if err != nil {
		log.Printf("fetchPublicWritingsInCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsCategoryPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
