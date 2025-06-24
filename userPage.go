package goa4web

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if data.CoreData.UserID == 0 {
		session, err := core.GetSession(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		redirectToLogin(w, r, session)
		return
	}

	if err := templates.RenderTemplate(w, "page.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
