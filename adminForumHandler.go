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
	err := renderTemplate(w, r, "adminForumPage.gohtml", data)
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
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/forum",
	}

	data.Messages = append(data.Messages, "Recalculating forum thread metadata...")

	if err := queries.RecalculateAllForumThreadMetaData(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("recalculateForumThreadByIdMetaData_firstpost: %w", err).Error())
	} else {
		data.Messages = append(data.Messages, "Thread metadata rebuilt.")
	}
	err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
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
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/forum",
	}
	data.Messages = append(data.Messages, "Rebuilding forum topic metadata...")
	if err := queries.RebuildAllForumTopicMetaColumns(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rebuildForumTopicByIdMetaColumns_lastaddition_lastposter: %w", err).Error())
	} else {
		data.Messages = append(data.Messages, "Topic metadata rebuilt.")
	}
	err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
