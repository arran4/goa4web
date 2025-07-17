package user

import (
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/middleware"

	"github.com/arran4/goa4web/core"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*handlers.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
	}

	if data.CoreData.UserID == 0 {
		session, err := core.GetSession(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		middleware.RedirectToLogin(w, r, session)
		return
	}

	handlers.TemplateHandler(w, r, "userPage", data)
}
