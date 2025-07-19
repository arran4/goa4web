package user

import (
	"log"
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func adminSessionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Sessions []*db.ListSessionsRow
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	items, err := queries.ListSessions(r.Context())
	if err != nil {
		log.Printf("list sessions: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Sessions = items
	handlers.TemplateHandler(w, r, "sessionsPage.gohtml", data)
}

func adminSessionsDeletePage(w http.ResponseWriter, r *http.Request) {
	sid := r.PostFormValue("sid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/admin/sessions",
	}
	if sid == "" {
		data.Errors = append(data.Errors, "missing sid")
	} else {
		if err := r.Context().Value(common.KeyQueries).(*db.Queries).DeleteSessionByID(r.Context(), sid); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
	}
	handlers.TemplateHandler(w, r, "tasks/run_task.gohtml", data)
}
