package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func forumAdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*GetAllGetAllForumCategoriesWithSubcategoryCountRow
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	categoryRows, err := queries.GetAllGetAllForumCategoriesWithSubcategoryCount(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	CustomForumIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "forumAdminCategoriesPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func forumAdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	if err := queries.UpdateForumCategory(r.Context(), UpdateForumCategoryParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		Idforumcategory:              int32(categoryId),
		ForumcategoryIdforumcategory: int32(pcid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/forum/admin/categories", http.StatusTemporaryRedirect)
}

func forumAdminCategoryCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if err := queries.CreateForumCategory(r.Context(), CreateForumCategoryParams{
		ForumcategoryIdforumcategory: int32(pcid),
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/forum/admin/categories", http.StatusTemporaryRedirect)
}
