package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func writingsAdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
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

	if err := getCompiledTemplates().ExecuteTemplate(w, "writingsAdminCategoriesPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func writingsAdminCategoriesUpdatePage(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func writingsAdminCategoriesModifyPage(w http.ResponseWriter, r *http.Request) {
	// TODO

	/*

			int pwcid = atoiornull(cont.post.getS("pwcid"));
		int wcid = atoiornull(cont.post.getS("wcid"));
		char *name = cont.post.getS("name");
		char *description = cont.post.getS("desc");
		changeWritingCategory(cont, wcid, name, description, pwcid);


	*/
}

func writingsAdminCategoriesDeletePage(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func writingsAdminCategoriesCreatePage(w http.ResponseWriter, r *http.Request) {
	// TODO

	/*

			int pwcid = atoiornull(cont.post.getS("pwcid"));
		char *name = cont.post.getS("name");
		char *description = cont.post.getS("desc");
		makeWritingCategory(cont, pwcid, name, description);


	*/
}
