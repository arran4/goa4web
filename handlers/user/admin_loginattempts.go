package user

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminLoginAttemptsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Attempts []*db.LoginAttempt
	}
	data := Data{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	items, err := queries.AdminListLoginAttempts(r.Context())
	if err != nil {
		log.Printf("list login attempts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Attempts = items
	handlers.TemplateHandler(w, r, "loginAttemptsPage.gohtml", data)
}
