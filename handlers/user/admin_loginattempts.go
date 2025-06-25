package user

import (
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func adminLoginAttemptsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Attempts []*db.LoginAttempt
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	items, err := queries.ListLoginAttempts(r.Context())
	if err != nil {
		log.Printf("list login attempts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Attempts = items
	if err := templates.RenderTemplate(w, "loginAttemptsPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
