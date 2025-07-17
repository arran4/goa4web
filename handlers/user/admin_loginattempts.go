package user

import (
	"log"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func adminLoginAttemptsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*handlers.CoreData
		Attempts []*db.LoginAttempt
	}
	data := Data{CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData)}
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	items, err := queries.ListLoginAttempts(r.Context())
	if err != nil {
		log.Printf("list login attempts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Attempts = items
	handlers.TemplateHandler(w, r, "loginAttemptsPage.gohtml", data)
}
