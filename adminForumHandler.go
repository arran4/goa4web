package main

import (
	_ "embed"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"log"
	"net/http"
)

func adminForumHandler(w http.ResponseWriter, r *http.Request) {
	// Data holds the data needed for rendering the template.
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	err := compiledTemplates.ExecuteTemplate(w, "adminForumPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminForumRemakeForumThreadHandler(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/forum",
	}
	if err := queries.update_forumthread_lastaddition(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumthread_lastaddition: %w", err).Error())
	}
	if err := queries.update_forumthread_comments(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumthread_comments: %w", err).Error())
	}
	if err := queries.update_forumthread_lastposter(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumthread_lastposter: %w", err).Error())
	}
	if err := queries.update_forumthread_firstpost(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumthread_firstpost: %w", err).Error())
	}
	err := compiledTemplates.ExecuteTemplate(w, "adminRunTaskPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminForumRemakeForumTopicHandler(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	data := struct {
		*CoreData
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/forum",
	}
	if err := queries.update_forumtopic_threads(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumtopic_threads: %w", err).Error())
	}
	if err := queries.update_forumtopic_comments(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumtopic_comments: %w", err).Error())
	}
	if err := queries.update_forumtopic_lastaddition_lastposter(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update_forumtopic_lastaddition_lastposter: %w", err).Error())
	}
	err := compiledTemplates.ExecuteTemplate(w, "adminRunTaskPage.tmpl", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
