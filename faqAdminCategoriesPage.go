package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func faqAdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []*Faqcategory
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.GetAllFAQCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	CustomFAQIndex(data.CoreData)

	renderTemplate(w, r, "faqAdminCategoriesPage.gohtml", data)
}

func faqCategoriesRenameActionPage(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.RenameFAQCategory(r.Context(), RenameFAQCategoryParams{
		Name: sql.NullString{
			String: text,
			Valid:  true,
		},
		Idfaqcategories: int32(cid),
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}

func faqCategoriesDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.DeleteFAQCategory(r.Context(), int32(cid)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}

func faqCategoriesCreateActionPage(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if err := queries.CreateFAQCategory(r.Context(), sql.NullString{
		String: text,
		Valid:  true,
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}
