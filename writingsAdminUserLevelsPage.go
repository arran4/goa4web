package main

import (
	"log"
	"net/http"
)

func writingsAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsAdminUserLevelsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsAdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO
	/*
		userAllow(cont, cont.post.getS("username"), cont.post.getS("level"));
	*/
}

func writingsAdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	// TODO

	/*
		userDisallow(cont, atoiornull(cont.post.getS("permid")));
	*/
}
