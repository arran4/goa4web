package goa4web

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
)

// adminForumFlaggedPostsPage displays posts flagged for moderator review.
func adminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData)}
	if err := templates.RenderTemplate(w, "forumFlaggedPostsPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
