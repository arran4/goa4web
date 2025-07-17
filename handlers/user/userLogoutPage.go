package user

import (
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"

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
		*corecommon.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecommon.CoreData),
	}

	// session retrieved earlier
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
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

	common.TemplateHandler(w, r, "logoutPage.gohtml", data)
}
