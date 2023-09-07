package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func linkerAdminQueuePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Queue []*ShowAdminQueueRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	queue, err := queries.ShowAdminQueue(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("showAdminQueue Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Queue = queue

	CustomLinkerIndex(data.CoreData, r)

	if err := getCompiledTemplates().ExecuteTemplate(w, "linkerAdminQueuePage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminQueueDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	if err := queries.DeleteQueueItem(r.Context(), int32(qid)); err != nil {
		log.Printf("updateQueue Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminQueueUpdateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	title := r.URL.Query().Get("title")
	URL := r.URL.Query().Get("URL")
	desc := r.URL.Query().Get("desc")
	category, _ := strconv.Atoi(r.URL.Query().Get("category"))
	if err := queries.UpdateQueue(r.Context(), UpdateQueueParams{
		LinkercategoryIdlinkercategory: int32(category),
		Title:                          sql.NullString{Valid: true, String: title},
		Url:                            sql.NullString{Valid: true, String: URL},
		Description:                    sql.NullString{Valid: true, String: desc},
		Idlinkerqueue:                  int32(qid),
	}); err != nil {
		log.Printf("updateQueue Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminQueueApproveActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	if err := queries.MoveToLinker(r.Context(), int32(qid)); err != nil {
		log.Printf("updateQueue Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	/*
		// TODO
			addToGeneralSearch(cont, description, value, "linkerSearch", "linker_idlinker");
			addToGeneralSearch(cont, title, value, "linkerSearch", "linker_idlinker");
	*/
	taskDoneAutoRefreshPage(w, r)
}
