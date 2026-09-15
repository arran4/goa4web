package user

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core"
)

func userLogoutPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Logout"
	_ = tasks.Template("domains/user/logout.gohtml").Handle(w, r, nil)
}

func userLogoutAction(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session, err := core.GetSession(r)
	if err != nil {
		core.SessionError(w, r, err)
	}
	uid, _ := session.Values["UID"].(int32)

	sm := cd.SessionManager()
	if ref, ok := session.Values["SessionRef"].(string); ok && ref != "" {
		if err := sm.DeleteSessionByID(r.Context(), core.HashSessionRef(ref)); err != nil {
			log.Printf("delete session: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	// Clear all values
	for k := range session.Values {
		delete(session.Values, k)
	}
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	// Issue 3104: explicit destruction of CSRF session state upon logout
	csrfSessionName := core.SessionName + "_csrf"
	csrfSession, _ := core.Store.Get(r, csrfSessionName)
	if csrfSession != nil {
		for k := range csrfSession.Values {
			delete(csrfSession.Values, k)
		}
		csrfSession.Options.MaxAge = -1
		if err := csrfSession.Save(r, w); err != nil {
			log.Printf("csrf session save error: %v", err)
		}
	}

	log.Printf("logout success uid=%d", uid)
	clearLoggedOutCoreData(cd)
	handlers.DisableCaching(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearLoggedOutCoreData(cd *common.CoreData) {
	cd.UserID = 0
	cd.CustomIndexItems = nil
}
