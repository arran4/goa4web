package user

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

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
	type Data struct {
		*common.CoreData
	}

	data := Data{
		CoreData: cd,
	}

	// session retrieved earlier
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")
	queries := cd.Queries()
	if session.ID != "" {
		if err := queries.DeleteSessionByID(r.Context(), session.ID); err != nil {
			log.Printf("delete session: %v", err)
		}
	}

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("logout success session=%s", handlers.HashSessionID(session.ID))

	data.CoreData.UserID = 0

	handlers.TemplateHandler(w, r, "logoutPage.gohtml", data)
}
