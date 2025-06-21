package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func blogsBloggerPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*GetBlogEntriesForUserDescendingRow
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
	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	userLanguagePref := 0

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	bu, _ := queries.GetUserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})

	buid := bu.Idusers

	rows, err := queries.GetBlogEntriesForUserDescending(r.Context(), GetBlogEntriesForUserDescendingParams{
		UsersIdusers:       buid,
		LanguageIdlanguage: int32(userLanguagePref),
		Limit:              15,
		Offset:             int32(offset),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Query Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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
			GetBlogEntriesForUserDescendingRow: row,
			EditUrl:                            editUrl,
		})
	}
	CustomBlogIndex(data.CoreData, r)

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "blogsBloggerPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
