package admin

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/eventbus"
)

type addAnnouncementTask struct{ eventbus.BasicTaskEvent }
type deleteAnnouncementTask struct{ eventbus.BasicTaskEvent }

func AdminAnnouncementsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Announcements []*db.ListAnnouncementsWithNewsRow
		NewsID        string
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.ListAnnouncementsWithNews(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list announcements: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Announcements = rows
	data.NewsID = r.FormValue("news_id")
	common.TemplateHandler(w, r, "announcementsPage.gohtml", data)
}

func (addAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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

func (deleteAnnouncementTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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
