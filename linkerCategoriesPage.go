package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func linkerCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*Linkercategory
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	categories, err := queries.GetAllLinkerCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Categories = categories

	CustomLinkerIndex(data.CoreData, r)
	if err := templates.RenderTemplate(w, "categoriesPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
