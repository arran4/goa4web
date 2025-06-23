package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func newsAnnouncementActivateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := queries.GetLatestAnnouncementByNewsID(r.Context(), int32(pid))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getLatestAnnouncementByNewsID: %v", err)
		}
	}
	if ann == nil {
		if err := queries.CreateAnnouncement(r.Context(), int32(pid)); err != nil {
			log.Printf("create announcement: %v", err)
		}
	} else if !ann.Active {
		if err := queries.SetAnnouncementActive(r.Context(), SetAnnouncementActiveParams{Active: true, ID: ann.ID}); err != nil {
			log.Printf("activate announcement: %v", err)
		}
	}
	taskDoneAutoRefreshPage(w, r)
}

func newsAnnouncementDeactivateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := queries.GetLatestAnnouncementByNewsID(r.Context(), int32(pid))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getLatestAnnouncementByNewsID: %v", err)
		}
		taskDoneAutoRefreshPage(w, r)
		return
	}
	if ann != nil && ann.Active {
		if err := queries.SetAnnouncementActive(r.Context(), SetAnnouncementActiveParams{Active: false, ID: ann.ID}); err != nil {
			log.Printf("deactivate announcement: %v", err)
		}
	}
	taskDoneAutoRefreshPage(w, r)
}
