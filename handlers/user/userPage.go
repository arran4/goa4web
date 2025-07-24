package user

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/middleware"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	if data.CoreData.UserID == 0 {
		session, err := data.CoreData.GetSession(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		middleware.RedirectToLogin(w, r, session)
		return
	}

	handlers.TemplateHandler(w, r, "userPage", data)
}
