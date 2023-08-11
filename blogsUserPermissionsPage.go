package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func blogsUserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	// TODO add guard
	type Data struct {
		*CoreData
		Rows []*blogsUserPermissionsRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.blogsUserPermissions(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	err = compiledTemplates.ExecuteTemplate(w, "adminUsersPermissionsPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsUsersPermissionsUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	username := r.PostFormValue("username")
	where := "blogs"
	level := r.PostFormValue("level")
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/blogs/bloggers",
	}
	if uid, err := queries.usernametouid(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("usernametouid: %w", err).Error())
	} else if err := queries.user_allow(r.Context(), user_allowParams{
		UsersIdusers: uid,
		Section: sql.NullString{
			String: where,
			Valid:  true,
		},
		Level: sql.NullString{
			String: level,
			Valid:  true,
		},
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("user_allow: %w", err).Error())
	}
	err := compiledTemplates.ExecuteTemplate(w, "adminRunTaskPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func blogsUsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/blogs/bloggers",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.userDisallow(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("createLanguage: %w", err).Error())
	}
	err := compiledTemplates.ExecuteTemplate(w, "adminRunTaskPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
