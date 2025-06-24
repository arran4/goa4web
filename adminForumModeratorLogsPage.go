package goa4web

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func adminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData)}
	if err := templates.RenderTemplate(w, "forumModeratorLogsPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
