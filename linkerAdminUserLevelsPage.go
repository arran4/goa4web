package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

func linkerAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		UserLevels []*Permission
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	rows, err := queries.getUsersPermissions(r.Context())
	if err != nil {
		log.Printf("getUsersPermissions Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.UserLevels = rows

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerAdminUserLevelsPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	username := r.PostFormValue("username")
	where := r.PostFormValue("where")
	level := r.PostFormValue("level")
	uid, err := queries.usernametouid(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("usernametouid Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.user_allow(r.Context(), user_allowParams{
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
		log.Printf("user_allow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	permid, _ := strconv.Atoi(r.PostFormValue("permid"))
	if err := queries.userDisallow(r.Context(), int32(permid)); err != nil {
		log.Printf("userDisallow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}
