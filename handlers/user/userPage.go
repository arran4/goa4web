package user

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/middleware"

	"github.com/arran4/goa4web/core"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "User Preferences"
	data := Data{
		CoreData: cd,
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
