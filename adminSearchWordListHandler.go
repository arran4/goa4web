package main

import (
	"database/sql"
	_ "embed"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"net/http"
)

func adminSearchWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []sql.NullString
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.CompleteWordList(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	renderTemplate(w, r, "adminSearchWordListPage.gohtml", data)
}
