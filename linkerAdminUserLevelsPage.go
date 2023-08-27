package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func linkerAdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	// Custom Index???

	CustomLinkerIndex(data.CoreData, r)
	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerAdminUserLevelsPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
}

func linkerAdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
}
