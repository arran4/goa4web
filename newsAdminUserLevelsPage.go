package main

import (
	"database/sql"
	"log"
	"net/http"
)

func newsAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// SKIP. TODO replace completely
	// Custom Index???

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "newsAdminUserLevelsPage.gohtml", data); err != nil {
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
	// TODO
	/*
			char *postid = cont.post.getS("id");
		if (postid == NULL)
			postid = "0";
		a4string query("DELETE FROM permissions WHERE idpermissions=%d",
				atoi(postid));
		a4mysqlResult *result = cont.sql.query(query.raw());
		delete result;

	*/
	taskDoneAutoRefreshPage(w, r)
}
