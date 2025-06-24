package goa4web

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func blogsBloggerPage(w http.ResponseWriter, r *http.Request) {
	type BlogRow struct {
		*GetBlogEntriesForUserDescendingLanguagesRow
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
	username := vars["username"]
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	bu, err := queries.GetUserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	})
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

	buid := bu.Idusers

	rows, err := queries.GetBlogEntriesForUserDescendingLanguages(r.Context(), GetBlogEntriesForUserDescendingLanguagesParams{
		UsersIdusers:  buid,
		ViewerIdusers: uid,
		Limit:         15,
		Offset:        int32(offset),
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
			GetBlogEntriesForUserDescendingLanguagesRow: row,
			EditUrl: editUrl,
		})
	}
	CustomBlogIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "bloggerPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
