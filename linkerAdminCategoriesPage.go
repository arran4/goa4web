package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func linkerAdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*Linkercategory
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	categoryRows, err := queries.GetAllLinkerCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("adminCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Categories = categoryRows

	CustomLinkerIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "linkerAdminCategoriesPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminCategoriesUpdatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	if err := queries.RenameLinkerCategory(r.Context(), RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminCategoriesRenamePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	if err := queries.RenameLinkerCategory(r.Context(), RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Idlinkercategory: int32(cid),
	}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminCategoriesDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	if err := queries.DeleteLinkerCategory(r.Context(), int32(cid)); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminCategoriesCreatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	title := r.PostFormValue("title")
	if err := queries.CreateLinkerCategory(r.Context(), sql.NullString{Valid: true, String: title}); err != nil {
		log.Printf("renameLinkerCategory Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}
