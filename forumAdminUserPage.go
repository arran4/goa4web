package main

import (
	"database/sql"
	"errors"
	"net/http"
)

func forumAdminUserPage(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		*CoreData
		Rows []*User
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.AllUsers(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	CustomForumIndex(data.CoreData, r)

	renderTemplate(w, r, "forumAdminUserPage.gohtml", data)
}
