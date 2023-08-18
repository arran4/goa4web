package main

import (
	_ "embed"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"log"
	"net/http"
)

func adminForumPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	err := getCompiledTemplates().ExecuteTemplate(w, "adminForumPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminForumRemakeForumThreadPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/forum",
	}
	if err := queries.update_forumthreads(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumthread_firstpost: %w", err).Error())
	}
	err := getCompiledTemplates().ExecuteTemplate(w, "adminRunTaskPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminForumRemakeForumTopicPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/forum",
	}
	if err := queries.update_forumtopics(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumtopic_lastaddition_lastposter: %w", err).Error())
	}
	err := getCompiledTemplates().ExecuteTemplate(w, "adminRunTaskPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
