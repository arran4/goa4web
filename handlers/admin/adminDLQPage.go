package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"
	"time"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	handlers "github.com/arran4/goa4web/handlers"

	db "github.com/arran4/goa4web/internal/db"
)

// DeleteDLQTask deletes entries from the dead letter queue.
type DeleteDLQTask struct{ tasks.TaskString }

var deleteDLQTask = &DeleteDLQTask{TaskString: TaskDelete}

func AdminDLQPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*common.CoreData
		Errors []*db.DeadLetter
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	rows, err := queries.ListDeadLetters(r.Context(), 100)
	if err != nil {
		log.Printf("list dead letters: %v", err)
	} else {
		data.Errors = rows
	}
	handlers.TemplateHandler(w, r, "admin/dlqPage.gohtml", data)
}

func (DeleteDLQTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	switch r.PostFormValue("task") {
	case string(TaskDelete):
		for _, idStr := range r.Form["id"] {
			if idStr == "" {
				continue
			}
			id, _ := strconv.Atoi(idStr)
			if err := queries.DeleteDeadLetter(r.Context(), int32(id)); err != nil {
				log.Printf("delete error: %v", err)
			}
		}
	case string(TaskPurge):
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
