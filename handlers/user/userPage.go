package user

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/middleware"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
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

	if err := templates.RenderTemplate(w, "userPage", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
