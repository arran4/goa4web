package user

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/middleware"

	"github.com/arran4/goa4web/core"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "User Preferences"

	if cd.UserID == 0 {
		session, err := core.GetSession(r)
		if err != nil {
			log.Printf("get session: %v", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
		_ = middleware.RedirectToLogin(w, r, session)
		return
	}

	handlers.TemplateHandler(w, r, "userPage", struct{}{})
}
