package user

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
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
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}

	// session retrieved earlier
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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

	if err := templates.RenderTemplate(w, "logoutPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
