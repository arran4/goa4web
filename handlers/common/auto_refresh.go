package common

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func TaskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	data.AutoRefresh = true

	if err := templates.RenderTemplate(w, "taskDoneAutoRefreshPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
