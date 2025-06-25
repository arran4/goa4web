package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

func adminAnnouncementsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Announcements []*ListAnnouncementsWithNewsRow
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	rows, err := queries.ListAnnouncementsWithNews(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list announcements: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Announcements = rows
	if err := templates.RenderTemplate(w, "announcementsPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminAnnouncementsAddActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	nid, err := strconv.Atoi(r.PostFormValue("news_id"))
	if err != nil {
		log.Printf("news id: %v", err)
		common.TaskDoneAutoRefreshPage(w, r)
		return
	}
	if err := queries.CreateAnnouncement(r.Context(), int32(nid)); err != nil {
		log.Printf("create announcement: %v", err)
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func adminAnnouncementsDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeleteAnnouncement(r.Context(), int32(id)); err != nil {
			log.Printf("delete announcement: %v", err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
