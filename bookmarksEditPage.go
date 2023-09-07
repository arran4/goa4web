package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func bookmarksEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		BookmarkContent string
		Bid             interface{}
	}

	data := Data{
		CoreData:        r.Context().Value(ContextValues("coreData")).(*CoreData),
		BookmarkContent: "Category: Example 1\nhttp://www.google.com.au Google\nColumn\nCategory: Example 2\nhttp://www.google.com.au Google\nhttp://www.google.com.au Google\n",
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	bookmarks, err := queries.Show_bookmarks(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("error show_bookmarks: %s", err)
			http.Error(w, "ERROR", 500)
			return
		}
	} else {
		data.BookmarkContent = bookmarks.List.String
		data.Bid = bookmarks.Idbookmarks
	}
	bookmarksCustomIndex(data.CoreData)

	if err := getCompiledTemplates().ExecuteTemplate(w, "bookmarksEditPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func bookmarksEditSaveActionPage(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	if err := queries.Update_bookmarks(r.Context(), Update_bookmarksParams{
		List: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/bookmarks/mine", http.StatusTemporaryRedirect)

}

func bookmarksEditCreateActionPage(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("text")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	uid, _ := session.Values["UID"].(int32)

	if err := queries.Add_bookmarks(r.Context(), Add_bookmarksParams{
		List: sql.NullString{
			String: text,
			Valid:  true,
		},
		UsersIdusers: uid,
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/bookmarks/mine", http.StatusTemporaryRedirect)

}
