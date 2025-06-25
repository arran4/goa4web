package goa4web

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
)

func taskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	data.AutoRefresh = true

	if err := templates.RenderTemplate(w, "taskDoneAutoRefreshPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func taskRedirectWithoutQueryArgs(w http.ResponseWriter, r *http.Request) {
	u := r.URL
	u.RawQuery = ""
	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
}
