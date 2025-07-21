package user

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
	db "github.com/arran4/goa4web/internal/db"
)

func userLogoutPage(w http.ResponseWriter, r *http.Request) {
	session, err := core.GetSession(r)
	if err != nil {
		core.SessionError(w, r, err)
	}
	uid, _ := session.Values["UID"].(int32)
	log.Printf("logout request session=%s uid=%d", session.ID, uid)
	type Data struct {
		*common.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	// session retrieved earlier
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if session.ID != "" {
		_ = queries.DeleteSessionByID(r.Context(), session.ID)
	}

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("logout success session=%s", session.ID)

	data.CoreData.UserID = 0

	handlers.TemplateHandler(w, r, "logoutPage.gohtml", data)
}
