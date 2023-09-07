package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func adminUsersPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []*User
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.AllUsers(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Rows = rows

	err = getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "adminUsersPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUsersDoNothingPage(w http.ResponseWriter, r *http.Request) {
	uid := r.PostFormValue("uid")
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users",
	}
	if _, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	}
	err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "adminRunTaskPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
