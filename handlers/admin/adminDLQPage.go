package admin

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core/templates"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminDLQPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Errors []*db.DeadLetter
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.ListDeadLetters(r.Context(), 100)
	if err != nil {
		log.Printf("list dead letters: %v", err)
	} else {
		data.Errors = rows
	}
	if err := templates.RenderTemplate(w, "admin/dlqPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminDLQAction(w http.ResponseWriter, r *http.Request) {
	DeleteDLQTask.Action()(w, r)
}

func (deleteDLQTask) action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	switch r.PostFormValue("task") {
	case "Delete":
		for _, idStr := range r.Form["id"] {
			if idStr == "" {
				continue
			}
			id, _ := strconv.Atoi(idStr)
			if err := queries.DeleteDeadLetter(r.Context(), int32(id)); err != nil {
				log.Printf("delete error: %v", err)
			}
		}
	case "Purge":
		before := r.PostFormValue("before")
		t := time.Now()
		if before != "" {
			if tt, err := time.Parse("2006-01-02", before); err == nil {
				t = tt
			}
		}
		if err := queries.PurgeDeadLettersBefore(r.Context(), t); err != nil {
			log.Printf("purge errors: %v", err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
