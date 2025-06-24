package goa4web

import (
	"database/sql"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func writingsWriterPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Abstracts []*GetPublicWritingsByUserRow
		Username  string
		IsOffset  bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			log.Printf("GetUserByUsername Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	rows, err := queries.GetPublicWritingsByUser(r.Context(), GetPublicWritingsByUserParams{
		UsersIdusers: u.Idusers,
		Limit:        15,
		Offset:       int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetPublicWritingsByUser Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData:  r.Context().Value(ContextValues("coreData")).(*CoreData),
		Abstracts: rows,
		Username:  username,
		IsOffset:  offset != 0,
	}

	CustomWritingsIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "writingsWriterPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
