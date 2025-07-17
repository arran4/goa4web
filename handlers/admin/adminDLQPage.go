package admin

import (
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"
	"time"

	handlers "github.com/arran4/goa4web/handlers"

	db "github.com/arran4/goa4web/internal/db"
)

type deleteDLQTask struct{ tasks.BasicTaskEvent }

func AdminDLQPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Errors []*db.DeadLetter
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*CoreData),
	}
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	rows, err := queries.ListDeadLetters(r.Context(), 100)
	if err != nil {
		log.Printf("list dead letters: %v", err)
	} else {
		data.Errors = rows
	}
	handlers.TemplateHandler(w, r, "admin/dlqPage.gohtml", data)
}

func (deleteDLQTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
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
	handlers.TaskDoneAutoRefreshPage(w, r)
}
