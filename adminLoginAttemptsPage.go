package goa4web

import (
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func adminLoginAttemptsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Attempts []*LoginAttempt
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	items, err := queries.ListLoginAttempts(r.Context())
	if err != nil {
		log.Printf("list login attempts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Attempts = items
	if err := templates.RenderTemplate(w, "loginAttemptsPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
