package news

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func NewsAnnouncementActivateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := cd.NewsAnnouncement(int32(pid))
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
		if err := queries.SetAnnouncementActive(r.Context(), db.SetAnnouncementActiveParams{Active: true, ID: ann.ID}); err != nil {
			log.Printf("activate announcement: %v", err)
		}
	}
	hcommon.TaskDoneAutoRefreshPage(w, r)
}

func NewsAnnouncementDeactivateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := cd.NewsAnnouncement(int32(pid))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getLatestAnnouncementByNewsID: %v", err)
		}
		hcommon.TaskDoneAutoRefreshPage(w, r)
		return
	}
	if ann != nil && ann.Active {
		if err := queries.SetAnnouncementActive(r.Context(), db.SetAnnouncementActiveParams{Active: false, ID: ann.ID}); err != nil {
			log.Printf("deactivate announcement: %v", err)
		}
	}
	hcommon.TaskDoneAutoRefreshPage(w, r)
}
