package user

import (
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func adminSessionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Sessions []*db.ListSessionsRow
	}
	data := Data{CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecommon.CoreData)}
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	items, err := queries.ListSessions(r.Context())
	if err != nil {
		log.Printf("list sessions: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Sessions = items
	common.TemplateHandler(w, r, "sessionsPage.gohtml", data)
}

func adminSessionsDeletePage(w http.ResponseWriter, r *http.Request) {
	sid := r.PostFormValue("sid")
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecommon.CoreData),
		Back:     "/admin/sessions",
	}
	if sid == "" {
		data.Errors = append(data.Errors, "missing sid")
	} else {
		if err := r.Context().Value(corecommon.KeyQueries).(*db.Queries).DeleteSessionByID(r.Context(), sid); err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
	}
	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
