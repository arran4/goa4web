package user

import (
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
)

func userLogoutPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Logout"
	session, err := core.GetSession(r)
	if err != nil {
		core.SessionError(w, r, err)
	}
	uid, _ := session.Values["UID"].(int32)
	log.Printf("logout request session=%s uid=%d", handlers.HashSessionID(session.ID), uid)

	// session retrieved earlier
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")
	sm := cd.SessionManager()
	if session.ID != "" {
		if err := sm.DeleteSessionByID(r.Context(), session.ID); err != nil {
			log.Printf("delete session: %v", err)
		}
	}

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	log.Printf("logout success session=%s", handlers.HashSessionID(session.ID))

	cd.UserID = 0

	UserLogoutPage.Handle(w, r, struct{}{})
}

const UserLogoutPage tasks.Template = "user/logoutPage.gohtml"
