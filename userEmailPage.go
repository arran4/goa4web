package main

import (
	"log"
	"net/http"
)

func userEmailPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "userEmailPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userEmailSaveActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		query.set("SELECT emailforumupdates FROM preferences WHERE users_idusers=%d", cont.user.UID);
		result = cont.sql.query(query.raw());
		int updates = cont.post.getS("emailupdates") != NULL;
		if (result->hasRow())
		{
			query.set("UPDATE preferences SET emailforumupdates=%d WHERE users_idusers=%d", updates, cont.user.UID);
		} else {
			query.set("INSERT INTO preferences (emailforumupdates, users_idusers) VALUES (%d, %d)", updates, cont.user.UID);
		}
		delete result;
		result = cont.sql.query(query.raw());
		delete result;
	*/
}

func userEmailTestActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		a4string url("http://%s%s", cont.env.get("HTTP_HOST"), cont.env.get("SCRIPT_NAME"));
		notifyChange(cont, cont.user.UID, url.raw());
		printf("Sent testmail.<br>\n");
	*/
}
