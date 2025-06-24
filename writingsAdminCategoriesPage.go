package goa4web

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

func writingsAdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*Writingcategory
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	categoryRows, err := queries.FetchAllCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("writingCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	CustomWritingsIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "categoriesPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsAdminCategoriesModifyPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	wcid, err := strconv.Atoi(r.PostFormValue("wcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	categoryId, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.UpdateWritingCategory(r.Context(), UpdateWritingCategoryParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		Idwritingcategory:                int32(categoryId),
		WritingcategoryIdwritingcategory: int32(wcid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	taskDoneAutoRefreshPage(w, r)
}

func writingsAdminCategoriesCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if err := queries.InsertWritingCategory(r.Context(), InsertWritingCategoryParams{
		WritingcategoryIdwritingcategory: int32(pcid),
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
	taskDoneAutoRefreshPage(w, r)
}
