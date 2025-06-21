package main

import (
	_ "embed"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
	"net/http"
)

func adminForumPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	renderTemplate(w, r, "adminForumPage.gohtml", data)
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

	if err := queries.RecalculateAllForumThreadMetaData(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("recalculateForumThreadByIdMetaData_firstpost: %w", err).Error())
	}
	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
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
	if err := queries.RebuildAllForumTopicMetaColumns(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rebuildForumTopicByIdMetaColumns_lastaddition_lastposter: %w", err).Error())
	}
	renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
}
