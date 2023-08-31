package main

import (
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func writingsAdminUserAccessPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	CustomWritingsIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsAdminUserAccessPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsAdminUserAccessAllowActionPage(w http.ResponseWriter, r *http.Request) {
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

func writingsAdminUserUpdateAllowActionPage(w http.ResponseWriter, r *http.Request) {
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

func writingsAdminUserAccessRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO

	/*

		{
		int wid = atoiornull(cont.post.getS("wid"));
		int uid = atoiornull(cont.post.getS("uid"));
		int isuserwriter = iswriter(cont, cont.user.UID, wid);
		if (level == auth_administrator || isuserwriter)
			deleteApprovedUser(cont, wid, uid);

	*/
}
