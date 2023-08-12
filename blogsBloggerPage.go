package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func blogsBloggerPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*show_latest_blogsRow
		IsEditable bool
	}
	type Data struct {
		*CoreData
		Rows     []*BlogRow
		IsOffset bool
		UID      string
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	username, _ := vars["username"]

	userLanguagePref := 0

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	uid, err := queries.usernametouid(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})

	rows, err := queries.show_latest_blogs(r.Context(), show_latest_blogsParams{
		UsersIdusers:       uid,
		LanguageIdlanguage: int32(userLanguagePref),
		Limit:              15,
		Offset:             int32(offset),
	})
	if err != nil {
		log.Printf("Query Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsOffset: offset != 0,
		UID:      strconv.Itoa(int(uid)),
	}

	for _, row := range rows {
		data.Rows = append(data.Rows, &BlogRow{
			show_latest_blogsRow: row,
			IsEditable:           true, // TODO atoiornull(result->getColumn(4)) == cont.user.UID || level == auth_administrator || level == auth_moderator
		})
	}
	CustomBlogIndex(data.CoreData, r)

	if err := compiledTemplates.ExecuteTemplate(w, "blogsBloggerPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
