package main

import (
	"database/sql"
	"errors"
	"net/http"
)

func blogsBloggersPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []*GetCountOfBlogPostsByUserRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.GetCountOfBlogPostsByUser(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	CustomBlogIndex(data.CoreData, r)

	renderTemplate(w, r, "blogsBloggersPage.gohtml", data)
}
