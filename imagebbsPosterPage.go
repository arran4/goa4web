package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func imagebbsPosterPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Posts    []*GetImagePostsByUserDescendingRow
		Username string
		IsOffset bool
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username := vars["username"]

	queries := r.Context().Value(common.KeyQueries).(*Queries)
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

	rows, err := queries.GetImagePostsByUserDescending(r.Context(), GetImagePostsByUserDescendingParams{
		UsersIdusers: u.Idusers,
		Limit:        15,
		Offset:       int32(offset),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetImagePostsByUserDescending Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Posts:    rows,
		Username: username,
		IsOffset: offset != 0,
	}

	CustomImageBBSIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "posterPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
