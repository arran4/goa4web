package main

import (
	"database/sql"
	_ "embed"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"log"
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

	err = renderTemplate(w, r, "adminSearchWordListPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// adminSearchWordListDownloadPage sends the full word list as a text file.
func adminSearchWordListDownloadPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.CompleteWordList(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=wordlist.txt")
	for _, row := range rows {
		if row.Valid {
			_, _ = w.Write([]byte(row.String + "\n"))
		}
	}
}
