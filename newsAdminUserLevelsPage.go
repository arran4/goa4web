package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func newsAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		UserLevels []*Permission
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.GetUsersPermissions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getUsersPermissions Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.UserLevels = rows

	CustomNewsIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "newsAdminUserLevelsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func newsAdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	username := r.PostFormValue("username")
	where := r.PostFormValue("where")
	level := r.PostFormValue("level")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.PermissionUserAllow(r.Context(), PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section: sql.NullString{
			String: where,
			Valid:  true,
		},
		Level: sql.NullString{
			String: level,
			Valid:  true,
		},
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)

}

func newsAdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	permid, err := strconv.Atoi(r.PostFormValue("permid"))
	if err != nil {
		log.Printf("strconv.Atoi(permid) Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := queries.PermissionUserDisallow(r.Context(), int32(permid)); err != nil {
		log.Printf("permissionUserDisallow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}
