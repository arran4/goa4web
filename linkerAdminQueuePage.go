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
		Queue []*GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetailsRow
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	queue, err := queries.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Queue = queue

	CustomLinkerIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "linkerAdminQueuePage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func linkerAdminQueueDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	if err := queries.DeleteLinkerQueuedItem(r.Context(), int32(qid)); err != nil {
		log.Printf("updateLinkerQueuedItem Error: %s", err)
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
	if err := queries.UpdateLinkerQueuedItem(r.Context(), UpdateLinkerQueuedItemParams{
		LinkercategoryIdlinkercategory: int32(category),
		Title:                          sql.NullString{Valid: true, String: title},
		Url:                            sql.NullString{Valid: true, String: URL},
		Description:                    sql.NullString{Valid: true, String: desc},
		Idlinkerqueue:                  int32(qid),
	}); err != nil {
		log.Printf("updateLinkerQueuedItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	taskDoneAutoRefreshPage(w, r)
}

func linkerAdminQueueApproveActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	lid, err := queries.SelectInsertLInkerQueuedItemIntoLinkerByLinkerQueueId(r.Context(), int32(qid))
	if err != nil {
		log.Printf("updateLinkerQueuedItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(lid))
	if err != nil {
		log.Printf("getLinkerItemById Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, text := range []string{link.Title.String, link.Description.String} {
		wordIds, done := SearchWordIdsFromText(w, r, text, queries)
		if done {
			return
		}
		if InsertWordsToLinkerSearch(w, r, wordIds, queries, lid) {
			return
		}
	}
	taskDoneAutoRefreshPage(w, r)
}
