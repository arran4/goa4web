package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strconv"
)

func blogsBloggerPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*show_latest_blogsRow
		EditUrl string
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
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	userLanguagePref := 0

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	buid, err := queries.usernametouid(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})

	rows, err := queries.show_latest_blogs(r.Context(), show_latest_blogsParams{
		UsersIdusers:       buid,
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
		UID:      strconv.Itoa(int(buid)),
	}

	for _, row := range rows {
		editUrl := ""
		if uid == row.UsersIdusers {
			editUrl = fmt.Sprintf("/blogs/blog/%d/edit", row.Idblogs)
		}
		data.Rows = append(data.Rows, &BlogRow{
			show_latest_blogsRow: row,
			EditUrl:              editUrl,
		})
	}
	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "blogsBloggerPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
